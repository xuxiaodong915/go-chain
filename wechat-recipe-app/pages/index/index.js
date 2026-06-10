// pages/index/index.js
const api = require('../../utils/api')
const util = require('../../utils/util')
const app = getApp()

// Category-specific emojis for recipe images
const CATEGORY_EMOJIS = {
  '川菜': '🌶️', '粤菜': '🥟', '湘菜': '🔥',
  '甜点': '🍰', '早餐': '🌅', '汤羹': '🍲',
  '素菜': '🥬', '面食': '🍜'
}

const DEFAULT_EMOJIS = ['🍳', '🥘', '🍝', '🥗', '🍛', '🥩', '🍕', '🌮']

Page({
  data: {
    loading: true,
    error: '',
    todayText: '',
    currentRecipe: {},
    recipeList: [],
    currentIndex: 0,
    backupList: [],
    trackOffset: 0,
    cardWidth: 375,
    _moving: false,
    isFavorited: false,
    favoriteIds: new Set()
  },

  // Touch state
  touchStartX: 0,
  isSwiping: false,
  dragOffset: 0,

  onLoad() {
    // .page has padding: 30rpx on each side = 60rpx total
    // 750rpx = windowWidth, so card-area width = (750-60)/750 * windowWidth
    const { windowWidth } = wx.getSystemInfoSync()
    const cardWidth = windowWidth * (690 / 750)
    // The gap between adjacent cards (matches .card-track gap: 20rpx)
    const gapPx = windowWidth * (20 / 750)
    this.actualCardWidth = cardWidth
    this.stepPx = cardWidth + gapPx
    this.setData({
      todayText: this.getTodayText(),
      cardWidth: cardWidth
    })
  },

  onShow() {
    // 如果有菜品图片被更新过，强制刷新
    const hasUpdates = Object.keys(app.globalData.updatedImages).length > 0
    const needsRefresh = hasUpdates && this.data.recipeList.some(r => app.globalData.updatedImages[r.id])

    if (needsRefresh) {
      // Clear the flags so we don't keep re-fetching
      this.setData({ recipeList: [] }, () => {
        this.loadData(true)
      })
      // Clear the update tracking
      app.globalData.updatedImages = {}
    } else {
      // 首次加载显示骨架屏，切 tab 回来时静默刷新（不闪）
      this.loadData(this.data.recipeList.length === 0)
    }
    this.loadFavorites()
  },

  getTodayText() {
    const now = new Date()
    const weekdays = ['日', '一', '二', '三', '四', '五', '六']
    const m = String(now.getMonth() + 1).padStart(2, '0')
    const d = String(now.getDate()).padStart(2, '0')
    return `${m}月${d}日 星期${weekdays[now.getDay()]}`
  },

  getCategoryEmoji(category) {
    return CATEGORY_EMOJIS[category] || DEFAULT_EMOJIS[Math.floor(Math.random() * DEFAULT_EMOJIS.length)]
  },

  async loadData(showLoading = false) {
    if (showLoading) {
      this.setData({ loading: true, error: '' })
    }
    try {
      const res = await api.getDailyRecommend()

      const current = res.recipe
      if (!current) {
        this.setData({ loading: false, error: '暂无菜谱数据' })
        return
      }

      current.emoji = this.getCategoryEmoji(current.category)
      current.ingredientsPreview = current.ingredients ? current.ingredients.join('、') : ''
      current.image_url = util.getImageUrl(current.image_url)

      const backupList = (res.backup || []).map(item => ({
        ...item,
        emoji: this.getCategoryEmoji(item.category),
        image_url: util.getImageUrl(item.image_url)
      }))

      const recipeList = [current, ...backupList]

      this.setData({
        loading: false,
        currentRecipe: current,
        recipeList: recipeList,
        currentIndex: 0,
        backupList: backupList,
        trackOffset: 0,
        isFavorited: this.data.favoriteIds.has(current.id)
      })
    } catch (err) {
      this.setData({
        loading: false,
        error: err.msg || '加载失败，请检查网络连接'
      })
    }
  },

  async loadFavorites() {
    try {
      const favs = await api.getFavorites(util.getUserId())
      const ids = new Set(favs.map(f => f.recipe.id))
      this.data.favoriteIds = ids
      if (this.data.currentRecipe.id) {
        this.setData({ isFavorited: ids.has(this.data.currentRecipe.id) })
      }
    } catch (err) {
      console.warn('Load favorites failed:', err)
    }
  },

  // === Track swipe handlers ===
  onTouchStart(e) {
    this.touchStartX = e.touches[0].clientX
    this.isSwiping = true
    this.dragOffset = 0
    this.setData({ _moving: true })
  },

  onTouchMove(e) {
    if (!this.isSwiping) return
    this.dragOffset = e.touches[0].clientX - this.touchStartX
    const { currentIndex, recipeList } = this.data
    const step = this.stepPx
    const maxOffset = 0
    const minOffset = -((recipeList.length - 1) * step)
    const newOffset = Math.max(minOffset, Math.min(maxOffset, -(currentIndex * step) + this.dragOffset))
    this.setData({ trackOffset: newOffset })
  },

  onTouchEnd(e) {
    if (!this.isSwiping) return
    this.isSwiping = false
    this.setData({ _moving: false })

    const { currentIndex, recipeList, cardWidth } = this.data
    const touch = e.changedTouches[0]
    const finalDiffX = touch.clientX - this.touchStartX
    const threshold = cardWidth * 0.25

    if (finalDiffX < -threshold && currentIndex < recipeList.length - 1) {
      // Swipe left → next
      this.goToCard(currentIndex + 1)
    } else if (finalDiffX > threshold && currentIndex > 0) {
      // Swipe right → previous
      this.goToCard(currentIndex - 1)
    } else {
      // Snap back
      this.setData({ trackOffset: -(currentIndex * this.stepPx) })
    }
  },

  goToCard(index) {
    const { recipeList, cardWidth } = this.data
    if (index < 0 || index >= recipeList.length) return

    const recipe = recipeList[index]
    this.setData({
      currentRecipe: recipe,
      currentIndex: index,
      trackOffset: -(index * this.stepPx),
      isFavorited: this.data.favoriteIds.has(recipe.id)
    })

    // If this backup recipe has no ingredients, fetch the full detail
    if (!recipe.ingredients || !Array.isArray(recipe.ingredients)) {
      this.loadIngredients(index)
    }
  },

  async loadIngredients(index) {
    const { recipeList } = this.data
    const recipe = recipeList[index]
    if (!recipe || !recipe.id) return

    try {
      const detail = await api.getRecipeDetail(recipe.id)
      recipe.ingredients = detail.ingredients || []
      recipe.ingredientsPreview = recipe.ingredients.join('、')
      recipe.steps = detail.steps || []

      // Update recipeList cache and re-render current card if it's this one
      const listKey = 'recipeList[' + index + ']'
      this.setData({
        [listKey + '.ingredients']: recipe.ingredients,
        [listKey + '.ingredientsPreview']: recipe.ingredientsPreview,
        [listKey + '.steps']: recipe.steps
      })

      // If this is the currently displayed card, also update currentRecipe
      if (this.data.currentIndex === index) {
        this.setData({
          'currentRecipe.ingredients': recipe.ingredients,
          'currentRecipe.ingredientsPreview': recipe.ingredientsPreview
        })
      }
    } catch (err) {
      console.warn('Failed to load ingredients for recipe', recipe.id, err)
    }
  },

  switchToRecipe(e) {
    const id = e.currentTarget.dataset.id
    const index = this.data.recipeList.findIndex(r => r.id === id)
    if (index >= 0) {
      this.goToCard(index)
    }
  },

  async toggleFavorite() {
    const { currentRecipe, isFavorited } = this.data
    try {
      if (isFavorited) {
        await api.removeFavorite(currentRecipe.id, util.getUserId())
        this.data.favoriteIds.delete(currentRecipe.id)
        this.setData({ isFavorited: false })
        util.showToast('已取消收藏')
      } else {
        await api.addFavorite(currentRecipe.id, util.getUserId())
        this.data.favoriteIds.add(currentRecipe.id)
        this.setData({ isFavorited: true })
        util.showToast('收藏成功', 'success')
      }
    } catch (err) {
      util.showToast(err.msg || '操作失败')
    }
  },

  async addToShoppingList() {
    const { currentRecipe } = this.data
    util.showLoading('添加中...')
    try {
      await api.addIngredientsFromRecipe(util.getUserId(), currentRecipe.id)
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

  goToDetail() {
    const { currentRecipe } = this.data
    wx.navigateTo({
      url: `/pages/recipe-detail/recipe-detail?id=${currentRecipe.id}`
    })
  }
})
