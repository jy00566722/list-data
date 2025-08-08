import { defineConfig } from 'wxt';
import path from 'node:path';

export default defineConfig({
  version: "0.0.2",
  // 在这里为项目根目录定义一个别名 '@'
  alias: {
    '@': path.resolve(__dirname, './'),
  },
  manifest: {
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
