import { createRouter, createWebHistory } from 'vue-router'
import ProductView from '../views/ProductView.vue'
import UploadView from '../views/UploadView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'products',
      component: ProductView
    },
    {
      path: '/upload',
      name: 'upload',
      component: UploadView
    }
  ]
})

export default router
