# Temu 卖家数据洞察平台 - 技术开发文档 (V2.0)

## 1. 项目概述

### 1.1. 项目名称
Temu卖家数据洞察平台

### 1.2. 项目目标
开发一套自动化工具，用于从Temu卖家后台收集关键业务数据（如商品销量、库存、备货建议等），并提供一个现代化的Web界面对这些数据进行聚合、分析和可视化展示，从而帮助卖家做出更精准的经营决策。

### 1.3. 核心架构
项目采用前后端分离的设计思想，由三个核心部分组成：

1.  **数据采集端 (Chrome 插件):** 以浏览器插件的形式，在用户正常浏览Temu卖家后台时，静默、无感地捕获关键API的返回数据，并将其安全地上报给后端服务。
2.  **后端服务 (Golang API):** 负责接收并持久化由插件上报的数据。同时，提供一套标准化的API接口，供前端进行数据查询、聚合与分析。
3.  **数据展示端 (Vue Web 应用):** 一个数据可视化仪表盘，调用后端API，以图表、表格等形式多维度地展示和分析业务数据，为卖家提供直观的洞察。

---

## 2. 数据采集端 (Chrome 插件) 详细设计

### 2.1. 最终方案：`fetch` 代理注入
为确保数据采集的稳定性、无感化和高成功率，我们采用**非侵入式的 `fetch` 代理注入方案**。

*   **原理:** 通过 `content_script` 向页面主环境（Main World）注入一个 `interceptor.js` 脚本。该脚本会重写页面原生的 `window.fetch` 函数。当页面自身的JS代码调用 `fetch` 请求数据时，我们的代理逻辑会检查请求的URL。如果URL是我们关心的目标API，代理会在请求正常返回后，**克隆**一份返回结果（Response），并将数据通过 `window.postMessage` 安全地发送给 `content_script`。原始的返回结果会原封不动地交给页面，确保对Temu后台的正常运行无任何干扰。

*   **优势:**
    *   **高稳定性:** 不依赖页面DOM结构或CSS选择器，只关心API接口，极大降低了因Temu前端更新而失效的风险。
    *   **无需破解:** 无需关心和破解Temu的 `anti-signature` 等动态签名参数，因为我们只是“搭便车”，让页面自己完成请求。
    *   **用户无感:** 整个过程在后台静默运行，用户无需打开开发者工具（F12），实现了真正的自动化采集。
    *   **非阻塞式:** 代理逻辑不会阻塞或延迟页面的正常网络请求。

### 2.2. 实现流程

1.  **`manifest.json`**:
    *   配置 `content_scripts`，使其在文档加载初期（`document_start`）就注入 `content.js`。
    *   通过 `web_accessible_resources` 声明，允许页面加载我们插件中的 `interceptor.js` 文件。

2.  **`content.js`**:
    *   **职责一 (注入):** 创建一个 `<script>` 标签，将其 `src` 指向 `interceptor.js`，并添加到文档的 `<head>` 中。注入后立即移除该标签，保持DOM干净。
    *   **职责二 (监听与转发):** 监听来自 `interceptor.js` 的 `message` 事件，接收到数据后，调用上报模块，将数据发送到Golang后端。

3.  **`interceptor.js`**:
    *   **核心代理:** 保存原始 `fetch` 函数的引用，然后使用自定义函数覆盖 `window.fetch`。
    *   **URL匹配:** 在自定义函数内部，检查请求URL是否包含在我们的目标API列表中。
    *   **数据捕获与发送:** 如果URL匹配，就克隆 `response`，读取其 `json` 内容，并通过 `window.postMessage` 将数据发送给 `content.js`。
    *   **返回原始Promise:** 无论是否捕获数据，都必须将原始 `fetch` 调用的 `Promise` 返回，确保页面行为一致。

### 2.3. 扩展性设计
为了方便未来增加对更多API（如库存、备货等）的监控，我们设计一个可配置的API处理器映射表。

```javascript
// interceptor.js 或 content.js 中
const ApiHandlerMap = {
  // Key: API URL中用于识别的独特部分
  '/mms/venom/api/supplier/sales/management/querySkuSalesNumber': {
    dataType: 'DAILY_SALES', // 数据类型标识
    // 可选：如果需要，可以在这里指定一个前端解析函数
    // parser: 'parseSkuSalesData', 
  },
  '/mms/venom/api/supplier/stock/management/queryStock': {
    dataType: 'INVENTORY_LEVELS',
  },
  // ... 未来在此处添加新的API监控配置
};

// content.js 在接收到数据后，可以连同dataType一起上报
// 后端根据dataType来决定数据存入哪个表、如何处理
```

---

## 3. 后端服务 (Golang API) 详细设计

### 3.1. 技术栈
*   **语言:** Golang
*   **Web框架:** Gin (或 Echo)
*   **数据库:** MySQL
*   **ORM/数据库驱动:** GORM (或 sqlx)

### 3.2. 数据库设计

*   **选型说明:**
    经过讨论，最终决定使用 **MySQL**。
    *   **理由:** 虽然项目的核心数据（如每日销量）具有时间序列特征，使用TimescaleDB等专业时序数据库在理论上性能更优，但考虑到项目初期的SKU总量和数据增长速率可控，MySQL的性能完全可以满足需求。同时，MySQL拥有广泛的社区支持、成熟的生态和稳定的事务处理能力，团队对其也更为熟悉，有利于快速开发和后期维护。

*   **表结构设计 (MySQL):**

    ```sql
    -- 每日SKC销量记录表
    CREATE TABLE `daily_sales` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
      `shop_name` VARCHAR(255) NOT NULL COMMENT '店铺名称',
      `skc` VARCHAR(255) NOT NULL COMMENT '商品SKC',
      `spu` VARCHAR(255) DEFAULT NULL COMMENT '商品SPU',
      `item_code` VARCHAR(255) DEFAULT NULL COMMENT '商家编码/货号',
      `sales_date` DATE NOT NULL COMMENT '销量统计日期',
      `sales_count` INT NOT NULL COMMENT '销量',
      `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
      `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
      `raw_data` JSON DEFAULT NULL COMMENT '原始API返回的JSON数据，用于调试和追溯',
      PRIMARY KEY (`id`),
      UNIQUE KEY `uk_shop_skc_date` (`shop_name`, `skc`, `sales_date`) COMMENT '店铺、SKC和日期的唯一索引，防止重复上报'
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='每日销量表';

    -- 未来可扩展的库存表 (示例)
    CREATE TABLE `inventory_levels` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
      -- ... 其他字段，如 skc, warehouse, quantity, record_time ...
      PRIMARY KEY (`id`)
    ) ENGINE=InnoDB COMMENT='库存水平表';
    ```

### 3.3. 应用架构
采用经典的分层架构，确保代码的清晰、可维护和可扩展。

*   **`main.go`**: 程序入口，负责初始化配置、数据库连接、路由等。
*   **`config/`**: 加载和管理应用配置（如数据库连接信息、服务器端口等）。
*   **`router/`**: 定义API路由。将HTTP路径映射到对应的`handler`函数。
    *   `POST /api/v1/data/daily-sales`: 接收并保存每日销量数据。
    *   `POST /api/v1/data/inventory`: (未来) 接收库存数据。
    *   `GET /api/v1/analysis/summary`: (未来) 提供数据分析汇总接口。
*   **`handler/` (或 `controller/`)**: 处理HTTP请求。负责解析和校验请求参数，并调用`service`层处理业务逻辑。
*   **`service/`**: 业务逻辑核心。执行数据转换、计算、聚合等复杂操作，不直接与数据库交互。
*   **`repository/` (或 `dao/`)**: 数据访问层。封装对数据库的增删改查操作（CRUD），供`service`层调用。
*   **`model/`**: 定义数据结构体，与数据库表进行映射。

### 3.4. 扩展流程
当需要支持新的数据类型（如库存）时，开发流程如下：
1.  **`model/`**: 新增 `inventory.go` 定义库存数据结构体。
2.  **`repository/`**: 新增 `inventory_repo.go` 实现库存数据的数据库操作。
3.  **`service/`**: 新增 `inventory_service.go` 处理库存相关的业务逻辑。
4.  **`handler/`**: 新增 `inventory_handler.go` 处理接收库存数据的HTTP请求。
5.  **`router/`**: 注册新的API路由，例如 `POST /api/v1/data/inventory`。

---

## 4. 前端应用 (Vue) 详细设计

### 4.1. 技术栈
*   **框架:** Vue 3 (使用 Composition API)
*   **构建工具:** Vite
*   **UI库:** Element Plus (或 Ant Design Vue)
*   **图表库:** Apache ECharts
*   **状态管理:** Pinia
*   **路由:** Vue Router

### 4.2. 功能模块与目录结构
```
src/
├── api/          # 封装所有对后端API的请求
│   ├── sales.js
│   └── index.js
├── assets/       # 静态资源 (CSS, images)
├── components/   # 可复用的UI组件
│   ├── charts/
│   │   └── SalesTrendChart.vue
│   └── layout/
│       └── AppLayout.vue
├── router/       # 路由配置
│   └── index.js
├── stores/       # Pinia状态管理
│   ├── user.js
│   └── salesData.js
├── views/        # 页面级组件
│   ├── Dashboard.vue
│   ├── DailySales.vue
│   └── SalesAnalysis.vue
├── App.vue       # 根组件
└── main.js       # 应用入口
```

### 4.3. 核心页面规划
1.  **仪表盘 (Dashboard):** 登录后的首页。展示最核心的KPI指标，如昨日总销量、本月累计销量、高销量SKU排行、店铺健康度等。
2.  **销量报表 (Sales Report):**
    *   提供按日、周、月查看销量数据的表格。
    *   支持按店铺、SKC、SPU进行筛选和搜索。
    *   提供数据导出功能。
3.  **销量分析 (Sales Analysis):**
    *   使用ECharts图表展示销量趋势、多SKU销量对比、销售额占比等。
    *   提供丰富的交互式筛选条件（如日期范围、产品分类等）。

---

## 5. 总结与后续步骤
本技术文档确立了项目的整体架构和关键技术选型。所有后续的开发工作都应严格遵循本文档中定义的设计方案和规范，以确保项目的高质量、可维护性和可扩展性。
