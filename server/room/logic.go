package room

import (
	. "../config"
	"../msg"
	"fmt"
	"log"
	"net"
)

//注册
//func OnRegister(manager *Manager, conn net.Conn) {
//	for {
//		b := make([]byte, 1024)
//		n, err := conn.Read(b)
//		if err != nil {
//			log.Print(" OnLogin conn error ")
//			return
//		}
//		fmt.Println("receive ", b)
//		cmd, body, err := msg.Deserialize(b[:n])
//		fmt.Println("aaaaaaaaaaaaa", body)
//		if err != nil {
//			fmt.Println(err)
//		}
//		if cmd == C2S_REGISTER { // 只监注册
//			isLegal, user := manager.CheckRegisterInfo(body.(msg.LoginInfo))
//			if isLegal {
//				user.SetConn(conn)
//				user.SetLogin()
//				manager.SendMsgByConn(conn, S2C_REGISTER_SUCCESS, nil)
//				go OnEnterHall(manager, user)
//				return
//			} else {
//				manager.SendMsgByConn(conn, S2C_REGISTER_FAIL, nil)
//			}
//		} else if cmd == C2S_HEART {
//			manager.StartHeartBeat(conn)
//		} else if cmd == C2S_LOGIN {
//			go OnLogin(manager, conn)
//			return
//		}
//	}
//}

func Register(manager *Manager, conn net.Conn, body msg.LoginInfo) bool {
	fmt.Println("Register : ", body)
	isLegal, user := manager.CheckRegisterInfo(body)
	if isLegal {
		user.SetConn(conn)
		user.SetLogin()
		manager.SendMsgByConn(conn, S2C_REGISTER_SUCCESS, nil)
		go OnEnterHall(manager, user)
		return true
	} else {
		manager.SendMsgByConn(conn, S2C_REGISTER_FAIL, nil)
	}
	return false
}

//处理登录逻辑
func OnLogin(manager *Manager, conn net.Conn) {
	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			log.Print(" OnLogin conn error ", err)
			return
		}
		cmd, body, err := msg.Deserialize(b[:n])
		if err != nil {
			manager.SendMsgByConn(conn, S2C_LOGIN_FAIL, nil)
		}
		if cmd == C2S_LOGIN { // 只监听请求登录消息
			isLegal, user := manager.CheckLoginInfo(body.(msg.LoginInfo))
			if isLegal {
				user.SetConn(conn)
				user.SetLogin()
				if user.Room != nil {
					user.Room.SetUserIn(user)
					go OnEnterRoom(manager, user)
				} else {
					go OnEnterHall(manager, user)
				}
				fmt.Println("玩家", user.Name, "登录成功！！")
				return
			} else {
				manager.SendMsgByConn(conn, S2C_LOGIN_FAIL, nil)
			}
		} else if cmd == C2S_HEART {
			manager.StartHeartBeat(conn)
		} else if cmd == C2S_REGISTER {
			//go OnRegister(manager, conn)
			if Register(manager, conn, body.(msg.LoginInfo)) {
				return
			}
		}
	}
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Println("something error ")
	//	}
	//}()
}

//处理进入大厅逻辑
func OnEnterHall(manager *Manager, user *User) {
	log.Println("OnEnterHall uid : ", user.Uid)
	t := msg.HallInfo{user.ConvertToUserInfo(), manager.GetHallInfo()}
	manager.SendMsgByConn(user.Conn, S2C_LOGIN_SUCCESS, t)
	for {
		b := make([]byte, 1024)
		n, err := user.Conn.Read(b)
		if err != nil {
			log.Println("OnEnterHall conn error ", err)
			OnTimeout(manager, user)
			return
		}
		cmd, body, err := msg.Deserialize(b[:n])
		if err != nil {
		}
		switch cmd {
		case C2S_LOGOUT:
			OnLogout(manager, user)
			return
		case C2S_ENTER_ROOM:
			isLegal, room := manager.CheckRoomInfo(body.(msg.LoginInfo))
			if isLegal {
				room.SetUserIn(user)
				go OnEnterRoom(manager, user)
				return
			}
		case C2S_HEART:
			manager.StartHeartBeat(user.Conn)
		}
	}
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Println("time out uid : ", user.Uid)
	//	}
	//}()
}

//处理进入房间逻辑
func OnEnterRoom(manager *Manager, user *User) {
	log.Println("OnEnterRoom uid : ", user.Uid)
	t := user.Room.ConvertToRoomInfo()
	t.UserInfo = user.ConvertToUserInfo()
	manager.SendMsgByConn(user.Conn, S2C_ENTER_ROOM_SUCCESS, t)
	for {
		b := make([]byte, 1024)
		n, err := user.Conn.Read(b)
		if err != nil {
			log.Print("OnEnterRoom conn error ")
			OnTimeout(manager, user)
			return
		}
		cmd, _, err := msg.Deserialize(b[:n])
		if err != nil {
		}
		switch cmd {
		case C2S_OUT_ROOM:
			OnOutRoom(manager, user)
			go OnEnterHall(manager, user)
			return
		case C2S_CHAT:
			OnChatMsg(manager, user, b[:n])
		case C2S_HEART:
			manager.StartHeartBeat(user.Conn)
		case C2S_LOGOUT:
			OnLogout(manager, user)
			return
		}
	}
	//defer func() {
	//	if err := recover(); err != nil {
	//		log.Println("time out")
	//	}
	//}()
}

//下线
func OnLogout(manager *Manager, user *User) {
	fmt.Println("OnLogout", user.Uid)
	user.SetOffline()
}

//退出房间
func OnOutRoom(manager *Manager, user *User) {
	manager.SendMsgByConn(user.Conn, S2C_OUT_ROOM_SUCCESS, nil)
	user.Room.SetUserOut(user)
}

func OnChatMsg(manager *Manager, user *User, b []byte) {
	user.Room.ReceiveMsg(b)
}

func OnTimeout(manager *Manager, user *User) {
	user.SetOffline()
}
