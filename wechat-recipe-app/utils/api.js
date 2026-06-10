const app = getApp()

const BASE_URL = app.globalData.baseUrl

/**
 * HTTP request wrapper
 */
function request(url, method = 'GET', data = {}) {
  return new Promise((resolve, reject) => {
    wx.showNavigationBarLoading()
    wx.request({
      url: BASE_URL + url,
      method: method,
      data: data,
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
        } else {
          reject({ code: res.statusCode, msg: res.data?.error || '请求失败' })
        }
      },
      fail: (err) => {
        reject({ code: -1, msg: '网络错误，请检查服务器是否启动' })
      },
      complete: () => {
        wx.hideNavigationBarLoading()
      }
    })
  })
}

/**
 * Recipe APIs
 */
function getDailyRecommend() {
  return request('/recipes/daily')
}

function getRecipes(categoryId = 0, page = 1, size = 20) {
  let data = { page, size }
  if (categoryId > 0) data.category_id = categoryId
  return request('/recipes', 'GET', data)
}

function getRecipeDetail(id) {
  return request('/recipes/' + id)
}

function getCategories() {
  return request('/categories')
}

/**
 * Favorite APIs
 */
function addFavorite(recipeId, userId) {
  return request('/favorites', 'POST', { user_id: userId, recipe_id: recipeId })
}

function removeFavorite(recipeId, userId) {
  return request('/favorites/' + recipeId + '?user_id=' + encodeURIComponent(userId), 'DELETE')
}

function getFavorites(userId) {
  return request('/favorites', 'GET', { user_id: userId })
}

/**
 * Shopping List APIs
 */
function getShoppingList(userId) {
  return request('/shopping-list', 'GET', { user_id: userId })
}

function addShoppingItem(userId, itemName, quantity = '') {
  return request('/shopping-list', 'POST', { user_id: userId, item_name: itemName, quantity })
}

function addIngredientsFromRecipe(userId, recipeId) {
  return request('/shopping-list/from-recipe', 'POST', { user_id: userId, recipe_id: recipeId })
}

function updateShoppingItem(id, checked) {
  return request('/shopping-list/' + id, 'PUT', { checked })
}

function deleteShoppingItem(id) {
  return request('/shopping-list/' + id, 'DELETE')
}

function clearCheckedItems(userId) {
  return request('/shopping-list/checked/clear?user_id=' + encodeURIComponent(userId), 'DELETE')
}

module.exports = {
  getDailyRecommend,
  getRecipes,
  getRecipeDetail,
  getCategories,
  addFavorite,
  removeFavorite,
  getFavorites,
  getShoppingList,
  addShoppingItem,
  addIngredientsFromRecipe,
  updateShoppingItem,
  deleteShoppingItem,
  clearCheckedItems
}
