import { ApiHandlerMap } from '@/api-config.js';

export default defineBackground({
  main() {
    console.log('后台脚本已加载 (V2 - 可扩展版)');

    // =================================================================================
    // 数据格式化模块
    // 负责将从API捕获的原始数据，转换为符合我们后端要求的、干净的结构。
    // =================================================================================

    /**
     * 格式化每日销量数据
     * @param {object} payload - 从API原始返回的JSON数据
     * @returns {Array|null} - 格式化后的数据数组，或在无效时返回null
     */
    function formatSalesData(payload) {
      // 根据我们之前分析的API返回结构，销量信息在 result.skuSalesNumberInfos 数组中
      if (!payload || !payload.result || !payload.result.length>0 || !payload.success) {
        console.warn('后台: 接收到的销量数据格式无效。', payload);
        return null;
      }

      return payload.result.map(item => ({
        sales_date: item.date,
        sku: item.prodSkuId,
        sales_number: item.salesNumber,
      }));
    }

    /**
     * 格式化库存数据 (未来扩展示例)
     * @param {object} payload - 从API原始返回的JSON数据
     * @returns {Array|null}
     */
    function formatInventoryData(payload) {
      // 这里是未来处理库存数据的逻辑
      console.log("格式化库存数据...", payload);
      // 假设库存数据结构是 payload.result.inventoryInfos
      // const inventory = payload.result.inventoryInfos.map(item => ({...}));
      // return inventory;
      return null; // 暂时返回null
    }


    // =================================================================================
    // 数据处理器中心
    // 核心分发逻辑：根据 dataType 查找对应的处理器，并执行。
    // 未来要支持新数据，只需在此处添加新的条目。
    // =================================================================================

    const dataProcessorMap = {
      'DAILY_SALES': {
        formatter: formatSalesData,
        getEndpoint: () => ApiHandlerMap['/mms/venom/api/supplier/sales/management/querySkuSalesNumber'].backendEndpoint,
      },
      'INVENTORY_LEVELS': {
        formatter: formatInventoryData,
        getEndpoint: () => ApiHandlerMap['/mms/venom/api/supplier/stock/management/queryStock']?.backendEndpoint,
      },
      // ... 在这里添加新的数据处理器
    };


    // =================================================================================
    // 后端通信与消息监听
    // =================================================================================

    /**
     * 将格式化后的数据发送到指定的后端接口
     * @param {string} endpoint - 目标后端URL
     * @param {object} data - 格式化后的数据
     */
    async function sendToBackend(endpoint, data) {
      if (!endpoint) {
        console.error('后台: 未能获取有效的后端接口地址。');
        return;
      }
      if (!data || data.length === 0) {
        console.log('后台: 没有有效数据可以发送。');
        return;
      }
      console.log(`后台: 准备将数据发送到 -> ${endpoint}`, data);

      try {
        const response = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        });

        if (!response.ok) {
          throw new Error(`后端返回错误: ${response.status} ${response.statusText}`);
        }
        const result = await response.json();
        console.log('后台: 成功发送数据到后端。响应:', result);
      } catch (error) {
        console.error('后台: 发送数据到后端时出错:', error);
      }
    }

    // 监听来自内容脚本的消息
    browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
      const { type, dataType, payload } = message;

      if (type === 'FROM_CONTENT') {
        console.log(`后台: 从内容脚本接收到数据, 类型: ${dataType}`, payload);
        
        const processor = dataProcessorMap[dataType];
        if (processor) {
          const formattedData = processor.formatter(payload);
          const endpoint = processor.getEndpoint();
          sendToBackend(endpoint, formattedData);
          sendResponse({ status: 'success', message: `数据 (类型: ${dataType}) 已被后台处理。` });
        } else {
          console.warn(`后台: 未找到针对数据类型 "${dataType}" 的处理器。`);
          sendResponse({ status: 'error', message: `未知的dataType: ${dataType}` });
        }
      }
      return true; // 异步发送响应
    });
  },
});
