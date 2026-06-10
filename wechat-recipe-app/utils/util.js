/**
 * Utility functions
 */

// Show toast message
function showToast(title, icon = 'none') {
  wx.showToast({
    title: title,
    icon: icon,
    duration: 2000
  })
}

// Show loading
function showLoading(title = '加载中...') {
  wx.showLoading({ title, mask: true })
}

// Hide loading
function hideLoading() {
  wx.hideLoading()
}

// Format date
function formatDate(dateStr) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

// Get the app user ID
function getUserId() {
  return getApp().globalData.userId
}

// Convert relative image URL to absolute URL for WeChat image component
function getImageUrl(path) {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  const baseUrl = getApp().globalData.baseUrl
  // baseUrl is like "http://localhost:8080/api" - remove /api suffix
  const host = baseUrl.replace(/\/api$/, '')
  return host + path
}

module.exports = {
  showToast,
  showLoading,
  hideLoading,
  formatDate,
  getUserId,
  getImageUrl
}
