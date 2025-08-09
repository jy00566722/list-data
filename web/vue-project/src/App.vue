<template>
  <el-container class="app-container">
    <el-header>
      <el-menu
        :default-active="activeIndex"
        class="el-menu-demo"
        mode="horizontal"
        :router="true"
        @select="handleSelect"
      >
        <el-menu-item index="/">商品列表</el-menu-item>
        <el-menu-item index="/upload">上传数据</el-menu-item>
      </el-menu>
    </el-header>
    <el-main>
      <RouterView />
    </el-main>
  </el-container>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { RouterView } from 'vue-router'

const activeIndex = ref('/')
const handleSelect = (key: string, keyPath: string[]) => {
  console.log(key, keyPath)
  activeIndex.value = key
}
</script>

<style>
/* Global styles */
body {
  margin: 0;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB',
    'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
}

.app-container .el-header {
  padding: 0;
}

.app-container .el-main {
  padding: 20px;
}
</style>
