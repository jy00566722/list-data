<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElTable, ElTableColumn, ElButton, ElMessage, ElInput, ElForm, ElFormItem, ElCard, ElImage } from 'element-plus'

// --- 响应式状态 ---
const allProducts = ref<Product[]>([]) // 存储从后端获取的所有商品
const loading = ref(false)
const searchFilters = ref({
  shop_code: '',
  spu: '',
  skc: '',
  sku: '',
  skc_code: ''
})

// --- TypeScript 接口定义 ---
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

// --- API 基地址 ---
const API_BASE_URL = 'http://localhost:8080/api/v1'

// --- 计算属性 ---

// 1. 根据搜索条件过滤商品
const filteredProducts = computed(() => {
  let products = allProducts.value
  const { shop_code, spu, skc, sku, skc_code } = searchFilters.value

  // 为确保合并功能正确，先进行排序
  products.sort((a, b) => {
    if (a.skc < b.skc) return -1
    if (a.skc > b.skc) return 1
    if (a.color_cn < b.color_cn) return -1
    if (a.color_cn > b.color_cn) return 1
    return 0
  })

  return products.filter(p => 
    (!shop_code || p.shop_code.toLowerCase().includes(shop_code.toLowerCase())) &&
    (!spu || p.spu.toLowerCase().includes(spu.toLowerCase())) &&
    (!skc || p.skc.toLowerCase().includes(skc.toLowerCase())) &&
    (!sku || p.sku.toLowerCase().includes(sku.toLowerCase())) &&
    (!skc_code || p.skc_code.toLowerCase().includes(skc_code.toLowerCase()))
  )
})

// 2. 计算表格行合并信息
const spanMap = computed(() => {
  const map = new Map<number, number>()
  let currentGroupStart = 0
  
  for (let i = 0; i < filteredProducts.value.length; i++) {
    // 当到达列表末尾，或者当前项与下一项不属于同一组时，计算合并
    if (i === filteredProducts.value.length - 1 || 
        filteredProducts.value[i].skc !== filteredProducts.value[i + 1].skc ||
        filteredProducts.value[i].color_cn !== filteredProducts.value[i + 1].color_cn) {
      
      const groupSize = i - currentGroupStart + 1
      map.set(currentGroupStart, groupSize)
      currentGroupStart = i + 1
    }
  }
  return map
})


// --- API 函数 ---

async function fetchProducts() {
  loading.value = true
  try {
    const response = await fetch(`${API_BASE_URL}/products`)
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
    const data = await response.json()
    allProducts.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('获取商品列表失败:', error)
    ElMessage.error('获取商品列表失败')
  } finally {
    loading.value = false
  }
}

async function deleteProduct(id: number) {
  try {
    const response = await fetch(`${API_BASE_URL}/products/${id}`, { method: 'DELETE' })
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
    ElMessage.success('商品删除成功')
    await fetchProducts() // 刷新列表
  } catch (error) {
    console.error('删除商品失败:', error)
    ElMessage.error('删除商品失败')
  }
}

// --- 事件处理 ---

function resetFilters() {
  searchFilters.value = { shop_code: '', spu: '', skc: '', sku: '', skc_code: '' }
}

// 表格合并方法
const objectSpanMethod = ({ row, column, rowIndex, columnIndex }: { row: Product, column: any, rowIndex: number, columnIndex: number }) => {
  // 只对“缩略图”列进行操作
  if (columnIndex === 0) {
    const rowspan = spanMap.value.get(rowIndex)
    if (rowspan) {
      return { rowspan: rowspan, colspan: 1 }
    } else {
      return { rowspan: 0, colspan: 0 } // 隐藏单元格
    }
  }
}


// --- 生命周期钩子 ---
onMounted(() => {
  fetchProducts()
})
</script>

<template>
  <div class="product-view-container">
    <!-- 查询区 (固定顶部) -->
    <el-card class="search-card">
      <el-form :model="searchFilters" class="search-form">
        <el-form-item label="店铺代码">
          <el-input v-model="searchFilters.shop_code" placeholder="店铺代码" clearable />
        </el-form-item>
        <el-form-item label="SPU">
          <el-input v-model="searchFilters.spu" placeholder="SPU" clearable />
        </el-form-item>
        <el-form-item label="SKC">
          <el-input v-model="searchFilters.skc" placeholder="SKC" clearable />
        </el-form-item>
        <el-form-item label="SKU">
          <el-input v-model="searchFilters.sku" placeholder="SKU" clearable />
        </el-form-item>
        <el-form-item label="SKC货号">
          <el-input v-model="searchFilters.skc_code" placeholder="SKC货号" clearable />
        </el-form-item>
        <el-form-item>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 表格区 (自动撑满并可滚动) -->
    <div class="table-container">
      <el-table 
        :data="filteredProducts" 
        v-loading="loading" 
        border 
        stripe 
        height="100%"
        :span-method="objectSpanMethod"
        class="product-table"
      >
        <el-table-column label="缩略图" width="120">
          <template #default="scope">
            <el-image 
              style="width: 100px; height: 100px"
              :src="scope.row.image_url" 
              :preview-src-list="[scope.row.image_url]"
              fit="cover" 
              lazy
            />
          </template>
        </el-table-column>
        <el-table-column prop="sku" label="SKU" width="180" />
        <el-table-column prop="skc" label="SKC" width="180" />
        <el-table-column prop="size" label="尺码" width="100" />
        <el-table-column prop="color_cn" label="颜色" width="100" />
        <el-table-column prop="shop_code" label="店铺代码" width="120" />
        <el-table-column prop="spu" label="SPU" width="180" />
        <el-table-column prop="skc_code" label="SKC货号" width="150" />
        <el-table-column prop="sku_code" label="SKU货号" width="150" />
        <el-table-column prop="bar_code" label="条码" width="150" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="scope">
            <el-button type="danger" size="small" @click="deleteProduct(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.product-view-container {
  display: flex;
  flex-direction: column;
  /* 减去顶部导航栏的高度和一些边距 */
  height: calc(100vh - 60px - 40px); 
  width: 100%;
}

.search-card {
  flex-shrink: 0; /* 防止查询区被压缩 */
  margin-bottom: 20px;
}

.search-form {
  display: flex;
  flex-wrap: nowrap; /* 强制不换行 */
  gap: 15px;
  align-items: center;
}

.search-form .el-form-item {
  margin-bottom: 0; /* 移除表单项的下边距 */
  flex-shrink: 1; /* 允许表单项收缩 */
  min-width: 150px; /* 设置一个最小宽度 */
}

.table-container {
  flex-grow: 1; /* 让表格容器填满剩余空间 */
  overflow-y: auto; /* 容器内部滚动 */
}

.product-table {
  width: 100%;
}
</style>
