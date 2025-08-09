<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElButton, ElUpload, ElMessage, ElDialog, ElIcon } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'

// --- Reactive State ---
const products = ref<Product[]>([])
const loading = ref(false)

// --- Interfaces ---
interface Product {
  id: number
  shop_id: number
  shop_code: string
  spu: string
  skc: string
  sku: string
  skc_code: string
  sku_code: string
  color_cn: string
  color_en: string
  size: string
  image_url: string
  bar_code: string
}

// --- API Base URL ---
const API_BASE_URL = 'http://localhost:8080/api/v1'

// --- API Functions ---

async function fetchProducts() {
  loading.value = true
  try {
    const response = await fetch(`${API_BASE_URL}/products`)
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    products.value = await response.json()
  } catch (error) {
    console.error('Failed to fetch products:', error)
    ElMessage.error('获取商品列表失败')
  } finally {
    loading.value = false
  }
}

async function deleteProduct(id: number) {
  try {
    const response = await fetch(`${API_BASE_URL}/products/${id}`, {
      method: 'DELETE'
    })
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    ElMessage.success('商品删除成功')
    await fetchProducts() // Refresh the list
  } catch (error) {
    console.error('Failed to delete product:', error)
    ElMessage.error('删除商品失败')
  }
}

function handleUploadSuccess() {
  ElMessage.success('文件上传成功，后台处理中...')
  fetchProducts() // Refresh the list after a short delay to allow backend processing
}

function handleUploadError(error: Error) {
  console.error('Upload failed:', error)
  ElMessage.error('文件上传失败')
}

// --- Lifecycle Hooks ---
onMounted(() => {
  fetchProducts()
})
</script>

<template>
  <div class="product-view">
    <h1>商品管理</h1>

    <div class="toolbar">
      <el-upload
        class="upload-component"
        :action="`${API_BASE_URL}/products/upload`"
        :show-file-list="false"
        :on-success="handleUploadSuccess"
        :on-error="handleUploadError"
        name="file"
      >
        <el-button type="primary">
          <el-icon class="el-icon--left"><UploadFilled /></el-icon>
          上传 Excel 新增/更新商品
        </el-button>
      </el-upload>
    </div>

    <el-table :data="products" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="sku" label="SKU" width="180" />
      <el-table-column prop="skc" label="SKC" width="180" />
      <el-table-column prop="shop_code" label="店铺代码" width="120" />
      <el-table-column prop="color_cn" label="颜色" width="100" />
      <el-table-column prop="size" label="尺码" width="100" />
      <el-table-column prop="skc_code" label="SKC货号" width="150" />
      <el-table-column prop="sku_code" label="SKU货号" width="150" />
      <el-table-column prop="bar_code" label="条码" width="150" />
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="scope">
          <el-button
            type="danger"
            size="small"
            @click="deleteProduct(scope.row.id)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.product-view {
  padding: 20px;
}
.toolbar {
  margin-bottom: 20px;
  display: flex;
  justify-content: flex-start;
}
</style>
