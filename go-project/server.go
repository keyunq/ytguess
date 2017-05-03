package main
import (
    "golang.org/x/net/websocket"
    "fmt"
    "log"
    "net/http"
    "github.com/go-redis/redis"
    "encoding/json"
    "strconv"
)
const max_room_num = 2

var (  
    JSON          = websocket.JSON              // codec for JSON  
    Message       = websocket.Message           // codec for string, []byte  
    ActiveClients = make(map[string]ClientConn) // map containing clients  //在线websocket列表
    User          = make(map[string]string)
)  

type ClientConn struct {
    websocket *websocket.Conn  
}

type UserMsg struct {
    Room string
    Cmd string
    User string
    AvatarUrl string
    Content string
    Uuid string
    HandNum string
    GuessNum string
}

type UserInfo struct {
    User string
    AvatarUrl string
    Uuid string
}

type ReplyMsg struct {
    Room string
    Cmd string
    Data string
}

type GuessResult struct {
    Result string
    CurrentNum int
    HandRecord map[string]string
    GuessRecord map[string]string
}

func echoHandler(ws *websocket.Conn) {
    var err error  
    var userMsg UserMsg
    
    for {  

        var data []byte
        if err = websocket.Message.Receive(ws, &data); err != nil {  
            fmt.Println("can't receive")  
            break  
        }

        err = json.Unmarshal(data, &userMsg)  
        fmt.Println(userMsg)

        go wsHandler(ws,userMsg)

    }  

}

func wsHandler(ws *websocket.Conn,userMsg UserMsg) {
    sockCli := ClientConn{ws}
    var err error


    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    //登录
    if userMsg.Cmd == "login" {
        fmt.Println("login") 
        //判断房间人数是否已满
        checkNumTmp := redisClient.SCard(userMsg.Room)
        checkNum := checkNumTmp.Val()
        if(checkNum < max_room_num) {
            fmt.Println("checkNum success") 
            //socket用户列表新增当前用户websocket连接
            ActiveClients[userMsg.Uuid] = sockCli
            //用户uuid保存到redis房间set集合内
            redisClient.SAdd("ROOM:"+userMsg.Room,userMsg.Uuid)

            var me UserInfo

            me.User = userMsg.User
            me.AvatarUrl = userMsg.AvatarUrl
            me.Uuid = userMsg.Uuid

            //生成用户信息json串
            b, err := json.Marshal(me)
            if err != nil {
                fmt.Println("Encoding User Faild")
            } else {
                //保存用户信息到redis
                redisClient.Set("USER:"+me.Uuid,b,0)

                //初始化用户
                initOnlineMsg(redisClient,userMsg)

            }
        } else {
            var rm ReplyMsg
            rm.Room = userMsg.Room
            rm.Cmd = "loginFailed"
            rm.Data = "登录失败，人数已满"

            sendMsg,err2 := json.Marshal(rm)
            sendMsgStr := string(sendMsg)
            fmt.Println(sendMsgStr) 
            if err2 != nil {
                
            } else {
                if err = websocket.Message.Send(ws, sendMsgStr); err != nil {  
                    log.Println("Could not send UsersList to ", userMsg.User, err.Error())  
                }
            }
        }

    //准备
    } else if userMsg.Cmd == "ready" {

        redisClient.Set("READY:"+userMsg.Uuid,"ready",0)

        //从redis取房间内的所有用户uuid
        roomSlice := redisClient.SMembers("ROOM:"+userMsg.Room)
        //用户uuid保存到一个go切片online 
        online := roomSlice.Val()

        i := 0

        //循环取在线用户个人信息
        if len(online) != 0 {
            for _, na := range online {  
                if na != "" {  
                    userJson := redisClient.Get("READY:"+na)
                    userJson2 := userJson.Val()
                    if userJson2 == "ready" {
                        i++
                    }
                }  
            }
        }
        if i == len(online) && i == max_room_num {
            var rm ReplyMsg
            rm.Room = userMsg.Room
            rm.Cmd = "start"
            rm.Data = ""

            broadcast(redisClient,userMsg,rm)
        }

    //退出
    } else if userMsg.Cmd == "logout" {
        fmt.Println("logout") 

        //socket用户列表删除该用户websocket连接
        delete(ActiveClients,userMsg.Uuid)
        //从redis房间set集合内删除该用户uuid
        redisClient.SRem("ROOM:"+userMsg.Room,userMsg.Uuid)

        //初始化用户
        initOnlineMsg(redisClient,userMsg)


    //出拳
    } else if userMsg.Cmd == "guess" {
        var result string
        fmt.Println("guess")
        fmt.Println(userMsg.HandNum)
        fmt.Println(userMsg.GuessNum)

        myHandNum,_ := strconv.Atoi(userMsg.HandNum)
        myGuessNum,_ := strconv.Atoi(userMsg.GuessNum)

        redisClient.Set("HANDNUM:"+userMsg.Uuid,myHandNum,0)
        redisClient.Set("GUESSNUM:"+userMsg.Uuid,myGuessNum,0)


        //从redis取房间内的所有用户uuid
        roomSlice := redisClient.SMembers("ROOM:"+userMsg.Room)
        //用户uuid保存到一个go切片online 
        online := roomSlice.Val()

        
        i := 0

        //循环取在线用户
        if len(online) != 0 {
            for _, na := range online {  
                if na != "" {  
                    handnumCmd := redisClient.Get("HANDNUM:"+na)
                    handnum := handnumCmd.Val()
                    if handnum != "" {
                        i++
                    }
                }  
            }
        }

        //房间内所有人都已提交，则计算最后结果
        if i == len(online) && i == max_room_num {
            var handRecordList map[string]string
            handRecordList = make(map[string]string)
            var guessRecordList map[string]string
            guessRecordList = make(map[string]string)

            //计算正确结果currentNum
            currentNum := 0
            //循环取在线用户
            if len(online) != 0 {
                for _, na := range online {  
                    if na != "" {  
                        //取某用户的出拳数据，已用户名为key，存入结果map
                        handnumCmd := redisClient.Get("HANDNUM:"+na)
                        handnum := handnumCmd.Val()
                        
                        guessnumCmd := redisClient.Get("GUESSNUM:"+na)
                        guessnum := guessnumCmd.Val()   

                        userJson := redisClient.Get("USER:"+na)
                        userJson2 := userJson.Val()

                        var user UserInfo
                        json.Unmarshal([]byte(userJson2), &user)

                        handRecordList[user.User] = handnum
                        guessRecordList[user.User] = guessnum

                        //计算结果
                        thandnum,_ := strconv.Atoi(handnum)
                        currentNum = currentNum + thandnum
                    }  
                }
            }

            //给各个用户发送结果消息
            if len(online) != 0 {
                for _, na := range online { 
                    if na != "" {
                        guessnumCmd := redisClient.Get("GUESSNUM:"+na)
                        guessnum := guessnumCmd.Val()
                        tguessnum ,_ := strconv.Atoi(guessnum)
                        if tguessnum == currentNum {
                            result = "1"
                        } else {
                            result = "0"
                        }
                        var guessResult GuessResult
                        guessResult.Result = result
                        guessResult.CurrentNum = currentNum
                        guessResult.HandRecord = handRecordList
                        guessResult.GuessRecord = guessRecordList

                        resultTmp,_ := json.Marshal(guessResult)
                        resultData := string(resultTmp)

                        //删除用户准备状态
                        redisClient.Del("READY:"+na)
                        //删除用户猜拳数据
                        redisClient.Del("HANDNUM:"+na)
                        redisClient.Del("GUESSNUM:"+na)

                        var rm ReplyMsg
                        rm.Room = userMsg.Room
                        rm.Cmd = "result"
                        rm.Data = resultData

                        sendMsg,_ := json.Marshal(rm)
                        sendMsgStr := string(sendMsg)

                        if err = websocket.Message.Send(ActiveClients[na].websocket, sendMsgStr); err != nil {  
                            log.Println("Could not send UsersList to ", "", err.Error())  
                        }


                    }
                }
            }
        }

    //发消息
    } else {
        /*
        //从redis取房间内的所有用户uuid
        roomSlice := redisClient.SMembers(userMsg.Room)
        //用户uuid保存到一个go切片online 
        online := roomSlice.Val()

        //循环给房间内用户发送消息
        if len(online) != 0 {
            for _, na := range online {  
                if na != "" {  
                    //ActiveClients[na].websocket就是用户对应的websocket链接
                    if err = websocket.Message.Send(ActiveClients[na].websocket, userMsg.User+"说："+userMsg.Content); err != nil {  
                        log.Println("Could not send message to ", userMsg.User, err.Error())  
                    }  
                }  
            }
        }*/


    }
}

//房间成员初始化,有人加入或者退出都要重新初始化，相当于聊天室的在线用户列表的维护
func initOnlineMsg(redisClient *redis.Client,userMsg UserMsg) {
    
    var err error

    //从redis取房间内的所有用户uuid
    roomSlice := redisClient.SMembers("ROOM:"+userMsg.Room)
    //用户uuid保存到一个go切片online 
    online := roomSlice.Val()

    var onlineList []string

    //循环取在线用户个人信息
    if len(online) != 0 {
        for _, na := range online {  
            if na != "" {  
                userJson := redisClient.Get("USER:"+na)
                userJson2 := userJson.Val()
                onlineList = append(onlineList,userJson2)
            }  
        }
    }
    fmt.Println("get online success") 
    //生成在线用户信息json串
    //c, err := json.Marshal(onlineList)

    onlineListStr,err2 := json.Marshal(onlineList)

    var rm ReplyMsg
    rm.Room = userMsg.Room
    rm.Cmd = "init"
    rm.Data = string(onlineListStr)

    sendMsg,err2 := json.Marshal(rm)
    sendMsgStr := string(sendMsg)
    fmt.Println("init") 
    if err2 != nil {
        
    } else {
        //给所有用户发初始化消息
        if len(online) != 0 {
            for _, na := range online {  
                if na != "" {                   
                    if err = websocket.Message.Send(ActiveClients[na].websocket, sendMsgStr); err != nil {  
                        log.Println("Could not send UsersList to ", "", err.Error())  
                    }
                }
            }
        }
        //若房间人数满，发送就绪消息
        if len(online) >= max_room_num {
            fmt.Println("full")
            var rm ReplyMsg
            rm.Room = userMsg.Room
            rm.Cmd = "full"
            rm.Data = ""

            sendMsg,_ := json.Marshal(rm)
            sendMsgStr := string(sendMsg)

            for _, na := range online {  
                if na != "" {                   
                    if err = websocket.Message.Send(ActiveClients[na].websocket, sendMsgStr); err != nil {  
                        log.Println("Could not send UsersList to ", "", err.Error())  
                    }
                }
            }
        }
    }
    
}

//广播消息
func broadcast(redisClient *redis.Client,userMsg UserMsg,rm ReplyMsg) {
    var err error

    //从redis取房间内的所有用户uuid
    roomSlice := redisClient.SMembers("ROOM:"+userMsg.Room)
    //用户uuid保存到一个go切片online 
    online := roomSlice.Val()

    sendMsg,err2 := json.Marshal(rm)
    sendMsgStr := string(sendMsg)
    fmt.Println("broadcast") 
    if err2 != nil {
        
    } else {
        //给所有用户发消息
        if len(online) != 0 {
            for _, na := range online {  
                if na != "" {                   
                    if err = websocket.Message.Send(ActiveClients[na].websocket, sendMsgStr); err != nil {  
                        log.Println("Could not send UsersList to ", "", err.Error())  
                    }
                }
            }
        }
    }
}

func main() {
    http.Handle("/echo", websocket.Handler(echoHandler))
    http.Handle("/", http.FileServer(http.Dir(".")))

    err := http.ListenAndServe(":8929", nil)

    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}