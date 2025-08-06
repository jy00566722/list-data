# 项目技术方案文档 (V1.0)
## 项目概述

##项目名称： Temu卖家数据洞察平台

## 项目目标：
- 开发一套工具，用于自动从Temu卖家后台收集关键业务数据（如销量、库存等），并提供一个Web界面对这些数据进行聚合、分析和可视化展示，帮助卖家更好地进行经营决策。

## 核心架构：
- 本项目采用前后端分离的微服务思想，由三个核心部分组成：
- 数据采集端 (Chrome插件): 负责在用户浏览Temu卖家后台时，无侵入式地捕获API数据，进行初步处理后上报。
- 后端服务 (Golang API): 负责接收、校验、处理并持久化前端上报的数据，同时提供数据查询和分析的API接口。
- 数据展示端 (Vue Web应用): 负责调用后端API，以图表、表格等形式多维度地展示和分析数据。

## 数据采集端 (Chrome 插件) 详细设计
-  预先注入的、非阻塞的fetch代理方式获取数据，再上报给content.js.再由content.js报给backgrouns.js，再报给后端服务器保存
- 时机：run_at: document_start
我们在manifest.json中配置content_scripts，让它在文档开始创建时就立即执行，甚至在页面的HTML、CSS、JS加载之前。
```javascript
// manifest.json
"content_scripts": [
  {
    "matches": ["https://seller.kuajingmaihuo.com/*", "https://agentseller.temu.com/*"],
    "js": ["content.js"],
    "run_at": "document_start" 
  }
],
// 还需要 web_accessible_resources 来让页面能加载我们的js文件
"web_accessible_resources": [{
    "resources": ["interceptor.js"],
    "matches": ["<all_urls>"]
}]
```

- content.js：唯一的任务是“注入”
- content.js的代码变得极其简单。它的唯一使命，就是在第一时间，向页面的主世界（Main World）注入另一个JS文件，我们称之为interceptor.js。这个interceptor.js才是真正干活的。同时，content.js要负责接收interceptor.js发来的数据，处理这些数据，整理成后端需要的格式再上报。

```javascript
// content.js
const s = document.createElement('script');
s.src = chrome.runtime.getURL('interceptor.js');
s.onload = function() {
    this.remove(); // 注入后就移除script标签，保持DOM干净
};
(document.head || document.documentElement).appendChild(s);

// 同时，在这里监听来自 interceptor.js 的消息
window.addEventListener("message", (event) => {
    // 确保消息来源是我们自己
    if (event.source === window && event.data.type === 'FROM_INTERCEPTOR') {
        const capturedData = event.data.payload;
        console.log("Content script received data:", capturedData);
        // 在这里调用上报后端的api.js模块
        // chrome.runtime.sendMessage({ type: 'UPLOAD_DATA', data: capturedData });
    }
}, false);
```

interceptor.js：核心代理逻辑（解决你的疑问的关键）
这个脚本被注入后，会立即运行在页面的JS环境中。它会在Temu自己的任何JS之前（或非常早的阶段）重写window.fetch。但它的重写方式是“非侵入式”和“非阻塞式”的。
```javascript
// interceptor.js
// 保存一份原始的、未被修改的fetch函数
const originalFetch = window.fetch;

window.fetch = function(...args) {
    // 获取请求的URL
    const url = args[0] instanceof Request ? args[0].url : args[0];

    // 调用原始的fetch，让请求正常发出。我们不阻塞、不修改请求本身。
    // 这会返回一个Promise
    const promise = originalFetch.apply(this, args);

    // 我们只对我们关心的API的返回结果感兴趣
    if (url.includes('/mms/venom/api/supplier/sales/management/querySkuSalesNumber')) {
        // 对返回的Promise进行处理
        promise.then(response => {
            // 关键：克隆一份response。因为response的body只能被读取一次。
            // 我们把克隆的给我们的插件用，原始的留给页面自己的JS代码用。
            const clonedResponse = response.clone();

            clonedResponse.json().then(data => {
                console.log('Interceptor captured data:', data);
                // 通过 postMessage 将捕获的数据发送给 content.js
                // 这是从页面主世界到Content Script隔离世界最安全的通信方式
                window.postMessage({ type: 'FROM_INTERCEPTOR', payload: data }, '*');
            }).catch(err => {
                // console.error('Error reading cloned response:', err);
            });
        });
    }

    // 最重要的一步：将原始的、未经任何处理的Promise返回给调用者。
    // 这样，Temu页面自身的JS代码收到的返回值和行为跟我们不存在时一模一样。
    return promise;
};
```

## 要做到可扩展性，后面不光只是上报销量数据，还有库存数据，备货数据等等

可扩展性设计: 定义一个“API处理器”映射表（ApiHandlerMap）。
```javascript
const ApiHandlerMap = {
  '/mms/venom/api/supplier/sales/management/querySkuSalesNumber': {
    parser: 'parseSkuSalesData', // 指向解析函数的名称
    endpoint: '/api/v1/data/daily-sales' // 指向上报到后端的API路径
  },
  '/another/api/for/total/sales': {
    parser: 'parseTotalSalesData',
    endpoint: '/api/v1/data/total-sales'
  }
  // ... 未来在这里添加新的API监控
};
```

## 后端采用golang+mysql的方案。
 后端应用架构

采用经典的分层架构，清晰、可维护、易扩展。

main.go: 程序入口，初始化路由、数据库连接、配置等。

config/: 配置文件加载。

router/: 定义API路由。使用 Gin 或 Echo 等主流Web框架。

POST /api/v1/data/daily-sales: 接收每日销量数据。

POST /api/v1/data/inventory: (未来扩展) 接收库存数据。

GET /api/v1/analysis/monthly-summary: (未来扩展) 获取月度汇总数据。

handler/ (或 controller/): 处理HTTP请求，参数校验，调用service层。

daily_sales_handler.go

service/: 业务逻辑核心。处理复杂逻辑，如数据转换、计算等。

daily_sales_service.go

repository/ (或 dao/): 数据访问层。负责与数据库交互，执行SQL。

daily_sales_repo.go (使用 sqlx 或 gorm 等库)

model/: 定义数据结构体，与数据库表映射。

daily_sales.go

可扩展性体现:
当需要支持新的数据类型（如库存）时，只需：

在 model/ 中添加 inventory.go。

在 repository/ 中添加 inventory_repo.go 并实现其数据库操作。

在 service/ 中添加 inventory_service.go。

在 handler/ 中添加 inventory_handler.go。

在 router/ 中注册新的API路由。
各个模块职责单一，互不干扰。

## 前端展示阶段
好的，非常理解你的需求和考量。这是一个非常典型的“数据采集 -> 数据处理与存储 -> 数据分析与展示”的全栈项目。你对数据采集部分的担忧和思考非常到位，这正是项目初期最需要明确的关键点。

在开始编写任何代码之前，我们先来构建一个全面、可扩展的技术方案文档。这份文档将作为我们后续开发的蓝图。

项目技术方案文档 (V1.0)
1. 项目概述
项目名称： Temu卖家数据洞察平台

项目目标：
开发一套工具，用于自动从Temu卖家后台收集关键业务数据（如销量、库存等），并提供一个Web界面对这些数据进行聚合、分析和可视化展示，帮助卖家更好地进行经营决策。

核心架构：
本项目采用前后端分离的微服务思想，由三个核心部分组成：

数据采集端 (Chrome插件): 负责在用户浏览Temu卖家后台时，无侵入式地捕获API数据，进行初步处理后上报。

后端服务 (Golang API): 负责接收、校验、处理并持久化前端上报的数据，同时提供数据查询和分析的API接口。

数据展示端 (Vue Web应用): 负责调用后端API，以图表、表格等形式多维度地展示和分析数据。

2. 数据采集端 (Chrome 插件) 详细设计
这部分是项目的起点，也是你最关心的部分。你的方案是可行的，但我们可以优化得更健壮、用户体验更好。

2.1. 数据捕获方案对比与选择

方案A: JS注入/请求劫持 (已排除)

优点: 可以实时获取数据。

缺点: 正如你所说，Temu前端逻辑复杂，JS混淆、反调试、以及动态生成的anti签名参数都使得此方案极不稳定，容易因平台更新而失效，维护成本极高。同意你的判断，放弃此方案。

方案B: chrome.devtools.network API

优点: 对页面无侵入，能稳定获取网络请求的完整信息（包括response body），可以绕过anti参数的破解难题。

缺点: 致命缺陷是它必须在开发者工具(DevTools)面板打开的情况下才能工作。这意味着用户每次想采集数据，都必须手动打开F12面板，这极大地影响了用户体验，无法做到“无感”采集。

方案C: chrome.webRequest API (推荐方案)

优点:

后台运行: 可以在插件的Background Script中静默运行，无需用户打开开发者工具，实现真正的无感、自动化采集。

对页面无侵入: 同样是监听网络请求，不修改页面DOM或JS，稳定性高。

功能强大: 可以监听到请求的URL、Header等，并且通过特定事件（如 onCompleted）可以判断请求成功。

挑战与解决方案:

挑战: chrome.webRequest API 无法直接获取Response Body。这是Chrome为了安全和隐私所做的限制。

解决方案: 我们可以采用一种“组合拳”的方式来解决：

Background Script (background.js): 使用 chrome.webRequest.onCompleted 监听特定API的成功请求，例如 .../querySkuSalesNumber。当监听到这个URL的请求成功完成时，我们知道了“此刻数据已经加载到了页面中”。

Content Script (content.js): 当Background Script监听到请求成功后，它会向当前页面的Content Script发送一个消息。

数据注入与提取: Content Script接收到消息后，执行一小段JS代码。此时，由于原始的API请求已经完成，数据实际上已经存在于页面的某个JS变量中，或者已经被渲染到了DOM的某个隐藏部分。Content Script的任务就是从这个“已知”的位置把数据提取出来。这比劫持请求要简单和稳定得多。如果数据没有直接暴露在全局变量中，我们可以通过注入一个 script 标签，重写 fetch 或 XMLHttpRequest.prototype.open/send，用一个极简的代理来捕获返回数据，然后通过 window.postMessage 安全地传递给Content Script。 这种方式比你最初担心的“劫持”要轻量和可控，我们只在自己的沙箱里操作。

2.2. 插件架构设计 (基于方案C)

manifest.json (配置文件):

声明插件名称、版本、权限等。

关键权限: webRequest, storage, activeTab, 以及对Temu卖家后台网址的访问权限 (host_permissions)。

注册background脚本。

注册content_scripts，并指定注入到Temu的两个域名下。

background.js (核心后台脚本):

职责: 网络监听、任务分发。

实现:

使用 chrome.webRequest.onCompleted 监听一个URL列表。

可扩展性设计: 定义一个“API处理器”映射表（ApiHandlerMap）。

JavaScript

const ApiHandlerMap = {
  '/mms/venom/api/supplier/sales/management/querySkuSalesNumber': {
    parser: 'parseSkuSalesData', // 指向解析函数的名称
    endpoint: '/api/v1/data/daily-sales' // 指向上报到后端的API路径
  },
  '/another/api/for/total/sales': {
    parser: 'parseTotalSalesData',
    endpoint: '/api/v1/data/total-sales'
  }
  // ... 未来在这里添加新的API监控
};
当监听到匹配的URL时，查找映射表，并向content.js发送消息，告知其需要执行哪个解析函数。

content.js (页面交互脚本):

职责: 提取数据、调用解析器。

实现:

监听background.js发来的消息。

通过上文提到的“注入script标签”的方式，安全地获取API的Response Body。

根据消息中的parser名称，调用对应的解析函数。

parsers/ (数据解析模块目录):

职责: 格式化数据。

实现:

skuSalesParser.js: 包含parseSkuSalesData函数。此函数接收原始API返回的JSON，将其转换成后端需要的标准格式，如 { shopName, skc, spu, itemCode, date, sales }。

totalSalesParser.js: 包含parseTotalSalesData函数，处理另一种API数据。

可扩展性: 未来需要支持新的API时，只需在此目录下新增一个对应的parser文件即可，无需改动核心逻辑。

api.js (后端通信模块):

职责: 将处理好的数据上报到Golang后端。

实现: 封装一个uploadData函数，处理POST请求、认证（如JWT Token）和错误处理。

这个架构将“监听”、“提取”、“解析”、“上报”四个环节解耦，具备极强的扩展性。增加对新API的支持，只需要在ApiHandlerMap中增加一条记录，并实现一个新的parser函数即可。

3. 后端服务 (Golang) 详细设计
3.1. 数据库选型分析

MySQL (关系型数据库):

优点: 结构化数据存储的王者，事务支持（ACID）非常完善，数据一致性高。对于店铺-SKC-日期-销量这种高度结构化的数据非常适合。聚合查询（如月度汇总）能力强大。

缺点: 在需要频繁变更表结构，或存储半结构化/非结构化数据时，灵活性较差。

MongoDB (文档型数据库):

优点: Schema-less（模式自由），非常灵活，适合快速迭代和存储结构多变的数据。对于不同API返回的不同格式数据，可以直接存为一个文档，扩展方便。

缺点: 事务支持相对较弱（虽然近年来已增强）。复杂的关联查询（JOIN）不如SQL方便，需要通过应用层代码实现，可能导致性能问题。对于需要强一致性和复杂聚合分析的场景，可能不是最佳选择。

TimescaleDB / InfluxDB (时间序列数据库):

优点: 这是此场景下的最佳选择。

为时间而生: 我们的核心数据“每日销量”是典型的时间序列数据（metric在timestamp的值）。TSDB 在这种场景下的写入和查询性能远超通用数据库。

高效聚合: 内置了大量针对时间的聚合函数（如下采样、滑动窗口、最新值等），计算“某一款的总销量”这类需求性能极高。

自动分区/分块: TimescaleDB（作为PostgreSQL的插件）会自动按时间对数据进行分区，查询近期数据时速度极快。

SQL的便利性: TimescaleDB 兼容完整的SQL，你可以继续使用熟悉的SQL语法进行查询，同时享受TSDB的性能优势，完美结合了MySQL的优点。

缺点: 相对MySQL来说，社区和运维经验可能稍少一些。但其基于PostgreSQL，生态非常成熟。

结论：
强烈推荐使用 PostgreSQL + TimescaleDB 插件。它既能满足当前结构化数据的存储和分析需求，又能以最高性能应对未来的时间序列分析。如果追求极致的简单和快速开发，MySQL 也是一个可靠的备选项。暂时不优先考虑MongoDB。

3.2. 数据库表结构设计 (以TimescaleDB/PostgreSQL为例)

SQL

-- 主表：每日SKC销量记录
CREATE TABLE daily_sales (
    time        TIMESTAMPTZ       NOT NULL, -- 数据记录的日期 (时间戳)
    shop_name   TEXT              NOT NULL, -- 店铺名称
    skc         TEXT              NOT NULL, -- SKC (Stock Keeping Unit Code)
    spu         TEXT,                       -- SPU (Standard Product Unit)
    item_code   TEXT,                       -- 货号
    sales       INTEGER           NOT NULL, -- 销量
    -- 可以增加一个元数据字段，用于存储原始API返回的JSON，便于排查问题
    raw_data    JSONB
);

-- 创建唯一索引，防止重复上报同一天同一个SKC的数据
CREATE UNIQUE INDEX idx_shop_skc_time ON daily_sales (shop_name, skc, time);

-- 将此普通表转换为TimescaleDB的超表 (Hypertable)，按时间分区
SELECT create_hypertable('daily_sales', 'time');

-- 未来可以有其他表，例如：
-- CREATE TABLE inventory_levels (...)
-- SELECT create_hypertable('inventory_levels', 'time');
3.3. 后端应用架构

采用经典的分层架构，清晰、可维护、易扩展。

main.go: 程序入口，初始化路由、数据库连接、配置等。

config/: 配置文件加载。

router/: 定义API路由。使用 Gin 或 Echo 等主流Web框架。

POST /api/v1/data/daily-sales: 接收每日销量数据。

POST /api/v1/data/inventory: (未来扩展) 接收库存数据。

GET /api/v1/analysis/monthly-summary: (未来扩展) 获取月度汇总数据。

handler/ (或 controller/): 处理HTTP请求，参数校验，调用service层。

daily_sales_handler.go

service/: 业务逻辑核心。处理复杂逻辑，如数据转换、计算等。

daily_sales_service.go

repository/ (或 dao/): 数据访问层。负责与数据库交互，执行SQL。

daily_sales_repo.go (使用 sqlx 或 gorm 等库)

model/: 定义数据结构体，与数据库表映射。

daily_sales.go

可扩展性体现:
当需要支持新的数据类型（如库存）时，只需：

在 model/ 中添加 inventory.go。

在 repository/ 中添加 inventory_repo.go 并实现其数据库操作。

在 service/ 中添加 inventory_service.go。

在 handler/ 中添加 inventory_handler.go。

在 router/ 中注册新的API路由。
各个模块职责单一，互不干扰。

4. 数据分析与展示前端 (Vue) 详细设计
4.1. 技术栈

框架: Vue 3 (Composition API)

构建工具: Vite

UI库: Element Plus 或 Ant Design Vue

图表库: ECharts

状态管理: Pinia

路由: Vue Router

4.2. 架构与功能模块

src/

main.js: 入口文件。

router/: 定义页面路由。

/dashboard: 仪表盘，展示核心KPI。

/sales/daily: 每日销量详情页。

/sales/analysis: 销量分析页。

store/: Pinia状态管理。

user.js: 管理用户信息。

sales.js: 管理销量相关数据、筛选条件等。

api/: 封装所有对后端Golang服务的请求。

views/: 页面级组件。

Dashboard.vue

DailySales.vue

components/: 可复用的UI组件。

DateRangePicker.vue: 日期范围选择器。

SalesChart.vue: 封装ECharts的销量图表。

ShopSelector.vue: 店铺选择器。

4.3. 可扩展性

基于组件: 新的分析图表或功能页面，都可以封装成独立的组件或视图，通过路由集成进来。

基于状态管理: 新的数据源（如库存）在后端准备好后，在store/中增加一个inventory.js模块来管理其状态，在api/中增加inventory.js来请求数据，然后创建新的view和component来展示，与现有功能完全解耦。