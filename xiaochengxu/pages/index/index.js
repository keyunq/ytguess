//index.js
//获取应用实例
var app = getApp()
Page({
  data: {
    userInfo: {},
    onlineList:{},
    status:0,
    statusStr:"等待中",
    guessBoxStatus:"hideBox",
    handList:['0','1','2','3','4','5'],
    handStyleList:['primary','default','default','default','default','default'],
    guessList:['0','1','2','3','4','5','6','7','8','9','10'],
    guessStyleList:['primary','default','default','default','default','default','default','default','default','default','default'],
    buttonList:['0','1','2'],
    buttonStrList:['准备','开始','提交'],
    buttonStyleList:['btnShow','btnHide','btnHide'],
    buttonFuncList:['ready','start','guess']
  },
  onLoad: function () {
    console.log("Page onLoad函数");
    /*
    wx.downloadFile({
      url: 'https://api.keyunq.com/8585.mp3', //仅为示例，并非真实的资源
      success: function(res) {
        console.log("下载音效成功"+res.tempFilePath);
        wx.playBackgroundAudio({
          dataUrl: res.tempFilePath,
          title: '古琴音效',
          coverImgUrl: 'http://avatar.csdn.net/4/D/B/1_qq_31383345.jpg',
          success: function() {
            console.log("播放音效")
          }
        })
      }
    })
    var tempFilePath = "https://api.keyunq.com/8585.mp3"
    wx.playVoice({
      filePath: tempFilePath,
      complete: function(){
        console.log("播放音效")
      }
    })*/
    wx.playBackgroundAudio({
      dataUrl: 'https://api.keyunq.com/8585.mp3',
      title: '古琴音效',
      coverImgUrl: 'https://api.keyunq.com/logo.png',
      success: function() {
        console.log("播放音效")
      }
    })
  },
  onHide: function() {
    console.log('发送注销消息')
    var myuuid = wx.getStorageSync('myuuid')
    var msg = new Object();
    msg.Room = '1';
    msg.Cmd = 'logout';
    msg.Uuid = myuuid;
    var str = JSON.stringify(msg)
    wx.sendSocketMessage({
      data:str
    })
    wx.closeSocket()
    app.globalData.onlineStatus = false
  },
  onShow: function() {
    var that = this
    app.getUserInfo().then(function(userInfo){
      that.setData({
        userInfo:userInfo
      })
      that.wsHandler(userInfo)
      that.initBox()
    })
  },
  wsHandler: function(userInfo) {
    var that = this

    //websocket
    wx.connectSocket({
      url: 'wss://ws.keyunq.com/echo'
    })
    wx.onSocketOpen(function(res) {
      console.log('WebSocket连接已打开！')
      var myuuid = wx.getStorageSync('myuuid')
      var msg = new Object();
      msg.Room = '1';
      msg.Cmd = 'login';
      msg.User = userInfo.nickName;
      msg.AvatarUrl = userInfo.avatarUrl;
      msg.Uuid = myuuid;
      var str = JSON.stringify(msg)
      wx.sendSocketMessage({
        data:str
      })
    })
    wx.onSocketMessage(function(res) {
      var msg = JSON.parse(res.data)
      if(msg.Cmd == 'init') {
        var userList = JSON.parse(msg.Data)
        app.globalData.onlineList = []
        for(var i=0;i<userList.length;i++){
          var user = JSON.parse(userList[i])
          app.globalData.onlineList.push(user)
        }
        /*
        var onlineNum = app.globalData.onlineList.length
        var user = msg.Data
        app.globalData.onlineList.push(user)
        */
        that.setData({
          onlineList:app.globalData.onlineList,
          status:0,
          statusStr:'等待中'
        })
      }
      if(msg.Cmd == 'full') {
        that.setData({
          status:1,
          statusStr:'准备开始'
        })
      }
      if(msg.Cmd == 'result') {
        
        var result = JSON.parse(msg.Data)

        var content = "总数为"+result.CurrentNum+"\n"
        for (var value in result.HandRecord) {
          content = content+value+"出拳："+result.HandRecord[value]+"\n";
        }
        for (var value in result.GuessRecord) {
          content = content+value+"猜拳："+result.GuessRecord[value]+"\n";
        }

        if(result.Result == 1) {
          content = "恭喜你，猜中啦\n" + content
          wx.showModal({
            content: content,
            showCancel: false,
            success: function (res) {
                if (res.confirm) {
                    that.initBox()
                }
            }
          });
        }
        if(result.Result == 0) {
          content = "很遗憾，猜错啦\n" + content
          wx.showModal({
            content: content,
            showCancel: false,
            success: function (res) {
                if (res.confirm) {
                    that.initBox()
                }
            }
          });
        }
        
      }
      if(msg.Cmd == 'start') {
        that.setData({
          status:2,
          statusStr:'游戏中',
          guessBoxStatus:'showBox',
          buttonStyleList:['btnHide','btnHide','btnShow'],
        })
      }
      
    })
  },
  setHandNum: function(event) {
    var that = this
    console.log(event.target.dataset.handnum)
    app.globalData.myHandNum = event.target.dataset.handnum
    var myList = that.data.handStyleList
    for(var i=0;i<myList.length;i++) {
      if(i == event.target.dataset.handnum) {
        myList[i] = 'primary'
      } else {
        myList[i] = 'default'
      }
    }
    that.setData({
      handStyleList:myList
    })
  },
  setGuessNum: function(event) {
    var that = this
    console.log(event.target.dataset.guessnum)
    app.globalData.myGuessNum = event.target.dataset.guessnum
    var myList = that.data.guessStyleList
    for(var i=0;i<myList.length;i++) {
      if(i == event.target.dataset.guessnum) {
        myList[i] = 'primary'
      } else {
        myList[i] = 'default'
      }
    }
    that.setData({
      guessStyleList:myList
    })
  },
  guess: function() {
    var that = this
    var userInfo = that.data.userInfo
    var myuuid = wx.getStorageSync('myuuid')
    var msg = new Object();
    msg.Room = '1';
    msg.Cmd = 'guess';
    msg.User = userInfo.nickName;
    msg.AvatarUrl = userInfo.avatarUrl;
    msg.Uuid = myuuid;
    msg.HandNum = app.globalData.myHandNum
    msg.GuessNum = app.globalData.myGuessNum
    var str = JSON.stringify(msg)
    wx.sendSocketMessage({
      data:str
    })
  },
  ready: function() {
    var that = this
    var userInfo = that.data.userInfo
    var myuuid = wx.getStorageSync('myuuid')
    var msg = new Object();
    msg.Room = '1';
    msg.Cmd = 'ready';
    msg.User = userInfo.nickName;
    msg.AvatarUrl = userInfo.avatarUrl;
    msg.Uuid = myuuid;
    var str = JSON.stringify(msg)
    wx.sendSocketMessage({
      data:str
    })
    that.setData({
      status:1,
      statusStr:'等待对手，准备开始',
      buttonStyleList:['btnHide','btnHide','btnHide'],
    })
    
  },
  start: function() {
    var that = this
    var userInfo = that.data.userInfo
    var myuuid = wx.getStorageSync('myuuid')
    var msg = new Object();
    msg.Room = '1';
    msg.Cmd = 'start';
    msg.User = userInfo.nickName;
    msg.AvatarUrl = userInfo.avatarUrl;
    msg.Uuid = myuuid;
    var str = JSON.stringify(msg)
    wx.sendSocketMessage({
      data:str
    })
  },
  initBox: function() {
    var that = this
    that.setData({
      status:0,
      statusStr:"等待中",
      guessBoxStatus:"hideBox",
      handList:['0','1','2','3','4','5'],
      handStyleList:['primary','default','default','default','default','default'],
      guessList:['0','1','2','3','4','5','6','7','8','9','10'],
      guessStyleList:['primary','default','default','default','default','default','default','default','default','default','default'],
      buttonList:['0','1','2'],
      buttonStrList:['准备','开始','提交'],
      buttonStyleList:['btnShow','btnHide','btnHide'],
      buttonFuncList:['ready','start','guess']
    })
  },
  getAudioStatus: function() {
    wx.getBackgroundAudioPlayerState({
      success: function(res) {
          var status = res.status
          var dataUrl = res.dataUrl
          var currentPosition = res.currentPosition
          var duration = res.duration
          var downloadPercent = res.downloadPercent
          console.log("音乐状态"+status)
          console.log("音乐长度"+duration)
      }
    })
  }
})
