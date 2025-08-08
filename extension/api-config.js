// extension/api-config.js

/**
 * API监控配置中心
 * 这是整个插件数据捕获功能的核心配置文件。
 * 未来如果需要监控新的API，只需要在此文件中添加新的条目即可。
 *
 * 键 (Key): API URL中独一无二、足以识别它的部分。
 * 值 (Value): 一个包含该API元数据的对象。
 *   - dataType: 一个独一无二的字符串，用于标识这是哪种数据。
 *   - backendEndpoint: 后端接收此数据的API接口地址。
 */
export const ApiHandlerMap = {
  // --- 每日销量API ---
  '/mms/venom/api/supplier/sales/management/querySkuSalesNumber': {
    dataType: 'DAILY_SALES',
    backendEndpoint: 'http://localhost:8080/api/v1/data/sales',
  },

  // --- 库存水平API (未来扩展示例) ---
  // '/mms/venom/api/supplier/stock/management/queryStock': {
  //   dataType: 'INVENTORY_LEVELS',
  //   backendEndpoint: 'http://localhost:8080/api/v1/data/inventory',
  // },
  
  // ... 在这里添加更多需要监控的API
};
