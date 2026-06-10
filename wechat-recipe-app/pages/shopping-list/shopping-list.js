// pages/shopping-list/shopping-list.js
const api = require('../../utils/api')
const util = require('../../utils/util')

Page({
  data: {
    items: [],
    loading: true,
    inputValue: '',
    inputQty: '',
    hasChecked: false
  },

  onShow() {
    this.loadList()
  },

  async loadList() {
    this.setData({ loading: true })
    try {
      const items = await api.getShoppingList(util.getUserId())
      const hasChecked = items.some(i => i.checked)
      this.setData({ items, hasChecked, loading: false })
    } catch (err) {
      this.setData({ loading: false })
      util.showToast(err.msg || '加载失败')
    }
  },

  onInputChange(e) {
    this.setData({ inputValue: e.detail.value })
  },

  onQtyChange(e) {
    this.setData({ inputQty: e.detail.value })
  },

  async addItem() {
    const { inputValue, inputQty } = this.data
    if (!inputValue.trim()) return

    try {
      await api.addShoppingItem(util.getUserId(), inputValue.trim(), inputQty)
      this.setData({ inputValue: '', inputQty: '' })
      this.loadList()
    } catch (err) {
      util.showToast(err.msg || '添加失败')
    }
  },

  async toggleItem(e) {
    const { id, checked } = e.currentTarget.dataset
    try {
      await api.updateShoppingItem(id, !checked)
      this.loadList()
    } catch (err) {
      util.showToast(err.msg || '操作失败')
    }
  },

  async deleteItem(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这项吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await api.deleteShoppingItem(id)
            this.loadList()
          } catch (err) {
            util.showToast(err.msg || '删除失败')
          }
        }
      }
    })
  },

  async clearChecked() {
    wx.showModal({
      title: '确认清除',
      content: '确定要清除所有已购买的食材吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await api.clearCheckedItems(util.getUserId())
            this.loadList()
          } catch (err) {
            util.showToast(err.msg || '操作失败')
          }
        }
      }
    })
  },

  goHome() {
    wx.switchTab({ url: '/pages/index/index' })
  }
})
