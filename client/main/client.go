package main

import (
	. "../config"
	"../msg"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

//用户名
var uName string

//接收数据通道
var recvDataClient = make(chan []byte)

//发送数据通道
var sendDataClient = make(chan []byte)

//关闭信息显示通道
//var stopMessage = make(chan bool)
var stopMessage bool

//显示数据通道
var messageChan = make(chan []byte)

//大厅信息
var hallInfo msg.HallInfo

func main() {
	connectServer()
}

//连接服务器
func connectServer() {
	conn, err := net.Dial("tcp", IP)
	checkErrClient(err)
	fmt.Println("连接成功")
	go recvClient(recvDataClient, conn)
	go sendClient(sendDataClient, conn)
	go func() {
		heartInterval := time.Tick(HEART_TIME)
		for {
			select {
			case <-heartInterval:
				sendHeardPacket()
				//case <-stopChan:
				//	return
			}
		}
	}()
	fmt.Println("1.注册  2.登录")
	clientLog(conn)
}

//发送心跳包
func sendHeardPacket() {
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	send, err := msg.Serialize(C2S_HEART, timeStamp)
	checkErrClient(err)
	sendDataClient <- send
}

//接收数据
func recvClient(recvDataClient chan []byte, conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		checkErrClient(err)
		recvDataClient <- buf[:n]
		//for i := 0; i < 10; i++ {
		//	messageChan <- []byte("111")
		//}
	}
}

////接收数据
//func recvClient(conn net.Conn) {
//	for {
//		buf := make([]byte, 1024)
//		n, err := conn.Read(buf)
//		checkErrClient(err)
//		recvDataClient <- buf[:n]
//		//for i := 0; i < 10; i++ {
//		//	messageChan <- []byte("111")
//		//}
//	}
//}

//发送数据
func sendClient(sendData chan []byte, conn net.Conn) {
	for {
		buf := <-sendData
		//fmt.Println("send: ", buf)
		_, err := conn.Write(buf)
		checkErrClient(err)
	}
}

//注册登录
func clientLog(conn net.Conn) {
	fmt.Println("请输入“1”或“2”来选择注册或登录：")
	var option string
	fmt.Scanln(&option)
	switch option {
	case "1":
		register(conn)
	case "2":
		login(conn)
	default:
		fmt.Println("输入的序号错误！")
		clientLog(conn)
	}
}

//注册程序
func register(conn net.Conn) {
	////跳转到登录页面
	//s, _ := msg.Serialize(C2S_REGISTER, nil)
	//sendDataClient <- s

	for {
		var userName, pwd, pwdCheck string
		fmt.Println("请输入用户名：")
		fmt.Scanln(&userName)
		fmt.Println("请输入密码：")
		fmt.Scanln(&pwd)
		fmt.Println("请确认密码：")
		fmt.Scanln(&pwdCheck)
		if pwdCheck != pwd {
			fmt.Println("两次密码不一样，请重新注册！")
			register(conn)
		} else {
			registerInfo := msg.LoginInfo{
				Name:     userName,
				Password: pwd,
			}
			send, err := msg.Serialize(C2S_REGISTER, registerInfo)
			checkErrClient(err)
			sendDataClient <- send
			res := <-recvDataClient
			//buf := make([]byte, 1024)
			//n, err := conn.Read(buf)
			//checkErrClient(err)
			//res := buf[:n]
			cmd, registerRes, err := msg.Deserialize(res)
			checkErrClient(err)
			switch cmd {
			case S2C_REGISTER_SUCCESS:
				fmt.Println("注册成功,请登录")
				login(conn) //登录
			case S2C_REGISTER_FAIL:
				fmt.Println("注册失败!", registerRes)
				register(conn)
			}
		}
	}
}

//登录程序
func login(conn net.Conn) {
	for {
		var userName, password string
		fmt.Println("请输入用户名：")
		fmt.Scanln(&userName)
		if userName == "" {
			fmt.Println("用户名为空!")
			login(conn)
		}
		fmt.Println("请输入密码：")
		fmt.Scanln(&password)
		loginInfo := msg.LoginInfo{
			Name:     userName,
			Password: password,
		}

		send, err := msg.Serialize(C2S_LOGIN, loginInfo)
		checkErrClient(err)
		sendDataClient <- send
		res := <-recvDataClient
		//buf := make([]byte, 1024)
		//n, err := conn.Read(buf)
		//checkErrClient(err)
		//res := buf[:n]
		cmd, loginRes, err := msg.Deserialize(res)
		switch cmd {
		case S2C_LOGIN_FAIL:
			fmt.Println("登录失败! ", loginRes)
			login(conn)
		case S2C_LOGIN_SUCCESS:
			uName = userName
			fmt.Println("登陆成功，进入大厅")
			hallinfo, ok := loginRes.(msg.HallInfo)
			if !ok {
				fmt.Println("接收大厅数据错误!")
			}
			hallInfo = hallinfo
			HallInfo(hallinfo, conn) //进入大厅
		case S2C_ENTER_ROOM_SUCCESS:
			uName = userName
			fmt.Println("登陆成功，进入聊天室")
			roomInfo, ok := loginRes.(msg.RoomInfo)
			if !ok {
				fmt.Println("接受聊天室信息失败!")
			}
			chatRoom(roomInfo, conn) //进入聊天室
		}
	}
}

//选择聊天室
func HallInfo(hallInfo msg.HallInfo, conn net.Conn) {
	var ordinal int
	for i, _ := range hallInfo.Rooms {
		fmt.Println("room_id: ", hallInfo.Rooms[i].Rid, "room_name: ", hallInfo.Rooms[i].Name, "online_num: ", hallInfo.Rooms[i].OnlineNum, "total_num: ", hallInfo.Rooms[i].OfflineNum)
	}
	fmt.Println("选择聊天室序号，请输入：", C2S_ENTER_ROOM, "选择登出，请输入：", C2S_LOGOUT)
	var choose uint8
	fmt.Scanln(&choose)
	if choose == C2S_LOGOUT {
		fmt.Println("登出成功")
		send, err := msg.Serialize(C2S_LOGOUT, nil)
		checkErrClient(err)
		sendDataClient <- send
		fmt.Println("再见")
		os.Exit(0)
	}
	fmt.Println("请选择聊天室序号：")
	fmt.Scanln(&ordinal)
	loginInfo := msg.LoginInfo{
		Rid: ordinal,
	}
	send, err := msg.Serialize(C2S_ENTER_ROOM, loginInfo)
	checkErrClient(err)
	sendDataClient <- send
	res := <-recvDataClient
	//buf := make([]byte, 1024)
	//n, err := conn.Read(buf)
	//checkErrClient(err)
	//res := buf[:n]
	cmd, roomRes, err := msg.Deserialize(res)
	checkErrClient(err)
	switch cmd {
	case S2C_ENTER_ROOM_SUCCESS:
		fmt.Println("进入聊天室")
		roomInfo, ok := roomRes.(msg.RoomInfo)
		if !ok {
			fmt.Println("接受聊天室信息失败")
		}
		chatRoom(roomInfo, conn) //进入聊天室
	}
}

//进入聊天室
func chatRoom(roomInfo msg.RoomInfo, conn net.Conn) {
	for i, _ := range roomInfo.Others {
		fmt.Println("Name:", roomInfo.Others[i].Name, " Status:", UserStatusConversion(roomInfo.Others[i].Status), " Title:", roomInfo.Others[i].Title)
	}
	go showMessage(conn) //启动聊天内容显示协程
	inputMessage(conn)   //启动聊天内容输入
}

//聊天内容
func showMessage(conn net.Conn) {
	for {
		//buf := make([]byte, 1024)
		//n, err := conn.Read(buf)
		//checkErrClient(err)
		//res := buf[:n]
		res := <-recvDataClient
		//fmt.Println("res: ", res)
		cmd, message, err := msg.Deserialize(res)
		//fmt.Println("cmd: ", cmd)
		checkErrClient(err)
		if cmd == C2S_CHAT {
			chat, ok := message.(msg.Chat)
			if !ok {
				fmt.Println("接受聊天信息失败")
			}
			fmt.Println(chat.Name + ":" + string(chat.Content))
		}
		if cmd == S2C_BROADCAST_LEVELCHANGED {
			levelchange, ok := message.(msg.LevelChange)
			if !ok {
				fmt.Println("接受等级变化失败")
			}
			fmt.Println("uid", levelchange.Uid, "oldtitle", levelchange.OldTitle, "newtitle", levelchange.NewTitle)
		}
		if stopMessage {
			fmt.Println("break  show ")
			break
		}
	}
	fmt.Println("退出显示聊天内容")
	return
}

//聊天内容输入
func inputMessage(conn net.Conn) {
	fmt.Println("登出，请输入：", C2S_LOGOUT, "退出当前聊天室，请输入：", C2S_OUT_ROOM)
	var input string
	for {
		fmt.Scanln(&input)
		switch input {
		case "3":
			send, err := msg.Serialize(C2S_LOGOUT, nil)
			checkErrClient(err)
			sendDataClient <- send
			fmt.Println("再见")
			os.Exit(0)
		case "7":
			send, err := msg.Serialize(C2S_OUT_ROOM, nil)
			checkErrClient(err)
			sendDataClient <- send
			res := <-recvDataClient
			//buf := make([]byte, 1024)
			//n, err := conn.Read(buf)
			//checkErrClient(err)
			//res := buf[:n]
			cmd, resinfo, err := msg.Deserialize(res)
			if cmd == S2C_LOGIN_SUCCESS {
				fmt.Println("返回大厅成功")
				stopMessage = true
				hallinfo, ok := resinfo.(msg.HallInfo)
				if !ok {
					fmt.Println("接收大厅数据错误!")
				}
				hallInfo = hallinfo
				HallInfo(hallinfo, conn) //进入大厅
			}
		default:
			if len(input) != 0 {
				inpubyte := []byte(input)
				chat := msg.Chat{
					Name:    uName,
					Content: inpubyte,
				}
				send, err := msg.Serialize(C2S_CHAT, chat)
				checkErrClient(err)
				sendDataClient <- send
			}
		}
		input = ""
	}
}

//检查错误
func checkErrClient(err error) {
	if err != nil {
		log.Println("an error: ", err.Error())
	}
}
