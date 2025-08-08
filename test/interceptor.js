// interceptor.js
// 保存一份原始的、未被修改的fetch函数
console.log("interceptor.js start...")
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
                //window.postMessage({ type: 'FROM_INTERCEPTOR', payload: data }, '*');
                window.postMessage({ type: 'FROM_INTERCEPTOR', payload: data }, '*');
                
            }).catch(err => {
                console.error('Error reading cloned response:', err);
            });
        });
    }

    // 最重要的一步：将原始的、未经任何处理的Promise返回给调用者。
    // 这样，Temu页面自身的JS代码收到的返回值和行为跟我们不存在时一模一样。
    return promise;
};