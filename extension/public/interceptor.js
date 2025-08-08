// public/interceptor.js

console.log("拦截器脚本已加载 (V5 - 最终健壮版)");

/**
 * 初始化 fetch 拦截器的主函数
 * @param {object} ApiHandlerMap - 从内容脚本获取到的API配置
 */
function initialize(ApiHandlerMap) {
  if (!ApiHandlerMap || Object.keys(ApiHandlerMap).length === 0) {
    console.error("拦截器: 初始化失败，未收到有效的API配置。");
    return;
  }

  const originalFetch = window.fetch;
  const targetApiUrls = Object.keys(ApiHandlerMap);
  console.log("拦截器: 成功接收配置，开始监控以下API:", targetApiUrls);

  // 重写 window.fetch
  window.fetch = function(...args) {
    const [resource, config] = args;
    let requestUrl;

    // 步骤 1: 健壮地获取请求的URL
    // 检查 resource 是字符串还是 Request 对象
    if (resource instanceof Request) {
      requestUrl = resource.url;
    } else {
      requestUrl = String(resource); // 确保是字符串
    }

    // 步骤 2: 检查URL是否匹配我们的目标
    const matchedUrl = targetApiUrls.find(target => requestUrl.includes(target));

    // 步骤 3: 无论是否匹配，都必须先调用原始的 fetch，并立即返回 Promise
    // 这是为了确保不阻塞页面的任何网络请求
    const fetchPromise = originalFetch(...args);

    // 步骤 4: 如果是我们关心的API，则在Promise上附加我们的处理逻辑
    if (matchedUrl) {
      console.log(`拦截器: 成功匹配到API -> ${requestUrl}`);
      
      fetchPromise.then(response => {
        // 必须克隆响应，这样我们才能读取它，同时不影响页面原始的读取流程
        const clonedResponse = response.clone();
        const apiConfig = ApiHandlerMap[matchedUrl];

        clonedResponse.json().then(data => {
          console.log("拦截器: 已捕获数据:", data);
          window.postMessage({
            type: 'FROM_INTERCEPTOR_DATA',
            dataType: apiConfig.dataType,
            payload: data
          }, '*');
        }).catch(err => {
          // 这个catch只处理 response.json() 的错误
          console.error("拦截器: 读取响应JSON时出错。", err);
        });
      }).catch(err => {
        // 这个catch处理原始fetch请求本身的错误 (例如网络问题)
        console.error(`拦截器: 原始请求失败 -> ${requestUrl}`, err);
      });
    }

    // 立即返回原始的Promise
    return fetchPromise;
  };
}

// 监听来自 content.js 的配置响应 (这部分逻辑不变)
window.addEventListener('message', (event) => {
  if (event.source === window && event.data && event.data.type === 'CONFIG_RESPONSE') {
    initialize(event.data.payload);
  }
});

// 脚本启动时，立即向 content.js 请求配置信息 (这部分逻辑不变)
console.log("拦截器: 正在向内容脚本请求API配置...");
window.postMessage({ type: 'GET_CONFIG_REQUEST' }, '*');
