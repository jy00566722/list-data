// content.js
console.log('content.js start....')
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