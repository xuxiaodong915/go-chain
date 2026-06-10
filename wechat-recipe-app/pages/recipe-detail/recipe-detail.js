// pages/recipe-detail/recipe-detail.js
const api = require('../../utils/api')
const util = require('../../utils/util')

const CATEGORY_EMOJIS = {
  '川菜': '🌶️', '粤菜': '🥟', '湘菜': '🔥',
  '甜点': '🍰', '早餐': '🌅', '汤羹': '🍲',
  '素菜': '🥬', '面食': '🍜'
}

Page({
  data: {
    recipe: {},
    loading: true,
    isFavorited: false
  },

  onLoad(options) {
    const id = options.id
    if (id) {
      this.loadRecipe(id)
      this.checkFavorite(id)
    }
  },

  async loadRecipe(id) {
    this.setData({ loading: true })
    try {
      const recipe = await api.getRecipeDetail(id)
      recipe.emoji = CATEGORY_EMOJIS[recipe.category] || '🍳'
      recipe.image_url = util.getImageUrl(recipe.image_url)
      this.setData({ recipe, loading: false })
      wx.setNavigationBarTitle({ title: recipe.name })
    } catch (err) {
      this.setData({ loading: false })
      util.showToast(err.msg || '加载失败')
    }
  },

  async checkFavorite(recipeId) {
    try {
      const favs = await api.getFavorites(util.getUserId())
      const favorited = favs.some(f => f.recipe.id == recipeId)
      this.setData({ isFavorited: favorited })
    } catch (err) {
      console.warn('Check favorite failed:', err)
    }
  },

  async toggleFavorite() {
    const { recipe, isFavorited } = this.data
    try {
      if (isFavorited) {
        await api.removeFavorite(recipe.id, util.getUserId())
        this.setData({ isFavorited: false })
        util.showToast('已取消收藏')
      } else {
        await api.addFavorite(recipe.id, util.getUserId())
        this.setData({ isFavorited: true })
        util.showToast('收藏成功', 'success')
      }
    } catch (err) {
      util.showToast(err.msg || '操作失败')
    }
  },

  async addAllToList() {
    const { recipe } = this.data
    util.showLoading('添加中...')
    try {
      await api.addIngredientsFromRecipe(util.getUserId(), recipe.id)
      util.hideLoading()
      wx.showModal({
        title: '添加成功',
        content: '已将食材加入购物清单',
        confirmText: '去查看',
        cancelText: '继续浏览',
        success: (res) => {
          if (res.confirm) {
            wx.switchTab({ url: '/pages/shopping-list/shopping-list' })
          }
        }
      })
    } catch (err) {
      util.hideLoading()
      util.showToast(err.msg || '添加失败')
    }
  },

  /**
   * Image failed to load - show emoji fallback instead of yellow background
   */
  onImageError() {
    this.setData({ 'recipe.image_url': '' })
  },

  /**
   * Replace recipe image - pick from album and upload
   */
  replaceImage() {
    const { recipe } = this.data
    wx.showActionSheet({
      itemList: ['从手机相册选择', '拍照'],
      success: (res) => {
        wx.chooseMedia({
          mediaType: ['image'],
          sourceType: [res.tapIndex === 0 ? 'album' : 'camera'],
          count: 1,
          sizeType: ['compressed'],
          success: (mediaRes) => {
            this.uploadImage(mediaRes.tempFiles[0].tempFilePath, recipe.id)
          }
        })
      }
    })
  },

  async uploadImage(filePath, recipeId) {
    const app = getApp()
    // baseUrl = "http://192.168.x.x:8080/api", upload endpoint = "/api/upload"
    const uploadUrl = app.globalData.baseUrl + '/upload'

    util.showLoading('上传中...')
    try {
      const res = await new Promise((resolve, reject) => {
        wx.uploadFile({
          url: uploadUrl,
          filePath: filePath,
          name: 'file',
          formData: { recipe_id: recipeId },
          success: (res) => {
            if (res.statusCode >= 200 && res.statusCode < 300) {
              resolve(JSON.parse(res.data))
            } else {
              reject({ msg: '上传失败' })
            }
          },
          fail: () => reject({ msg: '网络错误' })
        })
      })

      // Mark this recipe as updated (for cross-page refresh)
      app.globalData.updatedImages[recipeId] = Date.now()

      util.hideLoading()
      util.showToast('图片已更换', 'success')

      // Reload recipe, force cache bust with timestamp
      const recipe = await api.getRecipeDetail(recipeId)
      recipe.emoji = CATEGORY_EMOJIS[recipe.category] || '🍳'
      if (recipe.image_url) {
        recipe.image_url = util.getImageUrl(recipe.image_url) + '?_t=' + app.globalData.updatedImages[recipeId]
      }
      this.setData({ recipe, loading: false })
      wx.setNavigationBarTitle({ title: recipe.name })
    } catch (err) {
      util.hideLoading()
      util.showToast(err.msg || '上传失败')
    }
  }
})
