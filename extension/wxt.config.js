import { defineConfig } from 'wxt';
import vue from '@wxt-dev/module-vue';
import path from 'node:path';

// See https://wxt.dev/api/config.html
export default defineConfig({
  modules: ['@wxt-dev/module-vue'],
  // 在这里为项目根目录定义一个别名 '@'
  alias: {
    '@': path.resolve(__dirname, './'),
  },
  manifest: {
    name: "Temu数据分析",
    short_name: "Temu Interceptor",
    description: "Temu销量分析，补货分析，条码打印.",
    version: "0.0.2",
    web_accessible_resources: [
      {
        resources: ['/interceptor.js'],
        matches: ['*://*.temu.com/*'],
      },
    ],
    permissions: [
      "tabs",
    ],
    host_permissions: [
      "*://*.temu.com/*"
    ]
  },
});
