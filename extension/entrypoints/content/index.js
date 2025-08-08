// entrypoints/content/index.js
import { ApiHandlerMap } from '@/api-config.js';

export default defineContentScript({
  matches: ['*://*.temu.com/*'],
  runAt: 'document_start',

  main() {
    console.log('内容脚本已加载 (V4 - 主动请求配置方案)');

    // 注入拦截器脚本 (这部分不变)
    const interceptorScript = document.createElement('script');
    interceptorScript.src = browser.runtime.getURL('/interceptor.js');
    (document.head || document.documentElement).appendChild(interceptorScript);

    // 监听来自页面（包括注入脚本）的所有消息
    window.addEventListener('message', (event) => {
      if (event.source !== window || !event.data) return;

      const { type, dataType, payload } = event.data;

      // 分支1: 如果是注入脚本在请求配置
      if (type === 'GET_CONFIG_REQUEST') {
        console.log('内容脚本: 收到配置请求，正在响应...');
        // 将配置信息发送回注入脚本
        window.postMessage({
          type: 'CONFIG_RESPONSE',
          payload: ApiHandlerMap
        }, '*');
        return; // 处理完毕
      }

      // 分支2: 如果是注入脚本发送来了捕获到的数据
      if (type === 'FROM_INTERCEPTOR_DATA') {
        console.log(`内容脚本: 从拦截器接收到数据, 类型: ${dataType}`);
        // 将数据转发给后台脚本
        browser.runtime.sendMessage({ type: 'FROM_CONTENT', dataType, payload })
          .catch(err => console.error('内容脚本: 发送消息到后台脚本时出错:', err));
        return; // 处理完毕
      }
    });
  },
});
