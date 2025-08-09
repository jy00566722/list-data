<template>
  <div class="upload-view">
    <h1>上传商品数据</h1>
    <p>请选择包含商品信息的 Excel 文件进行上传。系统会自动根据 SKU 判断是新增还是更新商品。</p>
    <el-upload
      class="upload-component"
      drag
      :action="`${API_BASE_URL}/products/upload`"
      :on-success="handleUploadSuccess"
      :on-error="handleUploadError"
      name="file"
    >
      <el-icon class="el-icon--upload"><upload-filled /></el-icon>
      <div class="el-upload__text">
        将文件拖到此处，或<em>点击上传</em>
      </div>
      <template #tip>
        <div class="el-upload__tip">
          仅支持 Excel 文件 (.xls, .xlsx)，且数据格式需符合模板要求。
        </div>
      </template>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { ElUpload, ElMessage, ElIcon } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'

const API_BASE_URL = 'http://localhost:8080/api/v1'

function handleUploadSuccess(response: any) {
  ElMessage.success(response.message || '文件上传成功！')
}

function handleUploadError(error: Error) {
  console.error('Upload failed:', error)
  // 尝试解析后端返回的JSON错误信息
  try {
    const errorResponse = JSON.parse(error.message || '{}');
    ElMessage.error(errorResponse.error || '文件上传失败，请检查文件内容或联系管理员')
  } catch (e) {
    ElMessage.error('文件上传失败，请检查文件内容或联系管理员')
  }
}
</script>

<style scoped>
.upload-view {
  max-width: 600px;
  margin: 40px auto;
  text-align: center;
}

.upload-view h1 {
  margin-bottom: 10px;
}

.upload-view p {
  color: #666;
  margin-bottom: 30px;
}

.upload-component {
  width: 100%;
}
</style>
