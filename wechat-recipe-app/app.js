// app.js
App({
  globalData: {
    // appid:wxecd056ac472b6f18
    // Will be replaced with actual WeChat openid after login
    userId: 'demo_user_001',

    // ===== API 地址配置 =====
    // 【本地开发】Go 后端在本地运行：
    baseUrl: 'http://192.168.66.141:8080/api',
    // 【云托管部署】上线后换成微信分配的域名：
    // baseUrl: 'https://你的云托管域名/api',

    // Color theme
    themeColor: '#FF6B35',

    // Track recently updated recipe images for cross-page refresh
    updatedImages: {} // { recipeId: timestamp }
  }
})
