//app.js
App({
  onLaunch: function () {
    console.log("App生命周期函数——onLaunch函数");
  },
  checkSession:function(mysessionid) {
    return new Promise(function(resolve, reject) {
      wx.request({
        url: 'https://api.keyunq.com/check.php',
        header: {
          sessionid:mysessionid
        },
        success: function(res) {
          console.log("检查sessionid是否有效")
          resolve(res.data)
        },
        fail: function(e) {
          reject(e)
        }
      })
    })
  },
  login:function() {
    return new Promise(function(resolve, reject) {
      wx.login({
        success: function (res0) {
          if (res0.code) {
            wx.request({
              url: 'https://api.keyunq.com/login.php',
              data: {
                code: res0.code
              },
              header: {
                  'content-type': 'application/json'
              },
              success: function(res) {
                console.log("取得新的sessionid")
                console.log(res.data)
                var mysessionid = res.data.k
                wx.setStorageSync("mysessionid",mysessionid)
                var myuuid = res.data.v
                wx.setStorageSync("myuuid",myuuid)
                resolve(mysessionid)
              },
              fail: function(e) {
                reject(e)
              }
            })
          }
        }
      })
    })
  },
  getWxUserInfo:function() {
    return new Promise(function(resolve, reject) {
      wx.getUserInfo({
        withCredentials: false,
        success: function(res) {
          console.log("取得新的userInfo")
          var userInfo = res.userInfo
          wx.setStorageSync("userInfo",userInfo)
          console.log("setUserInfo")
          resolve(userInfo)
        }
      })
    })
  },
  getUserInfo:function() {
    var that = this
    return new Promise(function(resolve, reject) {
      var mysessionid = wx.getStorageSync('mysessionid')
      if(mysessionid) {
        console.log("sessionid存在")
        that.checkSession(mysessionid).then(function(sessionContent){
          if(sessionContent == 0) {
            console.log("sessionid无效-取userInfo存到本地")
            that.login().then(function(){
              that.getWxUserInfo().then(function(userInfo){
                resolve(userInfo)
              })
            })
          } else {
            console.log("sessionid有效-直接取本地userInfo")
            var userInfo = wx.getStorageSync("userInfo")
            resolve(userInfo)
          }
        })
        
      } else {
        console.log("sessionid不存在,重新走登录流程")
        that.login().then(function(){
          that.getWxUserInfo().then(function(userInfo){
            resolve(userInfo)
          })
        })
      }
    })
  },
  globalData:{
    userInfo:null,
    onlineList:[],
    onlineStatus:false,
    myHandNum:0,
    myGuessNum:0
  }
})