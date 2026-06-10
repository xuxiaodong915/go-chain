// pages/categories/categories.js
const api = require('../../utils/api')
const util = require('../../utils/util')

const CATEGORY_EMOJIS = {
  '川菜': '🌶️', '粤菜': '🥟', '湘菜': '🔥',
  '甜点': '🍰', '早餐': '🌅', '汤羹': '🍲',
  '素菜': '🥬', '面食': '🍜'
}

Page({
  data: {
    categories: [],
    currentCategory: 0,
    recipes: [],
    loading: true,
    page: 1,
    hasMore: true
  },

  onLoad() {
    this.loadCategories()
    this.loadRecipes()
  },

  async loadCategories() {
    try {
      const cats = await api.getCategories()
      this.setData({ categories: cats })
    } catch (err) {
      console.warn('Load categories failed:', err)
    }
  },

  async loadRecipes(reset = false) {
    if (reset) {
      this.setData({ page: 1, recipes: [], hasMore: true })
    }
    if (!this.data.hasMore) return

    this.setData({ loading: true })
    try {
      const res = await api.getRecipes(
        this.data.currentCategory,
        this.data.page,
        20
      )
      const newRecipes = res.recipes.map(r => ({
        ...r,
        emoji: CATEGORY_EMOJIS[r.category] || '🍳',
        image_url: util.getImageUrl(r.image_url)
      }))
      this.setData({
        recipes: reset ? newRecipes : [...this.data.recipes, ...newRecipes],
        loading: false,
        hasMore: this.data.recipes.length + newRecipes.length < res.total
      })
    } catch (err) {
      this.setData({ loading: false })
      util.showToast(err.msg || '加载失败')
    }
  },

  selectCategory(e) {
    const id = parseInt(e.currentTarget.dataset.id)
    if (id === this.data.currentCategory) return
    this.setData({ currentCategory: id }, () => {
      this.loadRecipes(true)
    })
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/recipe-detail/recipe-detail?id=${id}` })
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({ page: this.data.page + 1 })
      this.loadRecipes()
    }
  }
})
