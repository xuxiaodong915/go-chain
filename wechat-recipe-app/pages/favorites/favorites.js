// pages/favorites/favorites.js
const api = require('../../utils/api')
const util = require('../../utils/util')

const CATEGORY_EMOJIS = {
  '川菜': '🌶️', '粤菜': '🥟', '湘菜': '🔥',
  '甜点': '🍰', '早餐': '🌅', '汤羹': '🍲',
  '素菜': '🥬', '面食': '🍜'
}

Page({
  data: {
    favorites: [],
    loading: true
  },

  onShow() {
    this.loadFavorites()
  },

  async loadFavorites() {
    this.setData({ loading: true })
    try {
      const favs = await api.getFavorites(util.getUserId())
      const items = favs.map(f => ({
        ...f,
        recipe: {
          ...f.recipe,
          emoji: CATEGORY_EMOJIS[f.recipe.category] || '🍳',
          image_url: util.getImageUrl(f.recipe.image_url)
        }
      }))
      this.setData({ favorites: items, loading: false })
    } catch (err) {
      this.setData({ loading: false })
      util.showToast(err.msg || '加载失败')
    }
  },

  async removeFavorite(e) {
    const recipeId = e.currentTarget.dataset.id
    try {
      await api.removeFavorite(recipeId, util.getUserId())
      util.showToast('已取消收藏')
      this.loadFavorites()
    } catch (err) {
      util.showToast(err.msg || '操作失败')
    }
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({ url: `/pages/recipe-detail/recipe-detail?id=${id}` })
  },

  goHome() {
    wx.switchTab({ url: '/pages/index/index' })
  }
})
