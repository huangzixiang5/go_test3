package room

import (
	. "../config"
	"../msg"
	"../mysql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

const MAX_MSG_BUF = 20

type ChatRoom struct {
	Rid          int
	Name         string
	OnlineNum    int
	OfflineNum   int
	OnlineUsers  []*User
	//OfflineUsers []*User
	msg          chan []byte
	mu           sync.Mutex

	msgBuf []mysql.RoomInfo
	msgNum int
}

type LevelChange struct {
	Uid      int
	OldTitle string
	NewTitle string
}

func CreateAndStartChatRoom(m *Manager, id int, name string) {
	room := &ChatRoom{
		Rid:          id,
		Name:         name,
		OnlineNum:    0,
		OfflineNum:   0,
		OnlineUsers:  make([]*User, 0, 20),
		OfflineUsers: make([]*User, 0, 0),
		msg:          make(chan []byte, 50),
		mu:           sync.Mutex{},

		msgBuf: make([]mysql.RoomInfo, 0, MAX_MSG_BUF),
		msgNum: 0,
	}
	m.AddRoom(room)
	go room.Run()
	go room.TraceLevel()
}

func (this *ChatRoom) TraceLevel() {
	for range time.Tick(time.Second * 5) {
		for _, user := range this.OnlineUsers {
			pre := user.Title
			_, _, title := UserLevelConversion(user.TotalTime + int(time.Now().Sub(user.LoginTime).Seconds()))
			//fmt.Printf("等级:pre=%s,title=%s,uid=%d,totaltime=%d,dtime=%d",pre,title,user.Uid,user.TotalTime,int(time.Now().Sub(user.LoginTime).Seconds()))
			if pre != title && pre != "" {
				this.BroadcastLevelChanged(&LevelChange{Uid: user.Uid, OldTitle: pre, NewTitle: title})
			}
			user.Title = title
		}
	}
}

func (this *ChatRoom) Run() {
	log.Print("ChartRoom Run with id/name  :  ", this.Rid, "/", this.Name)
	for {
		select {
		case b := <-this.msg:
			this.SendMsg(b)
		}
	}
}

func (this *ChatRoom) SetUserIn(user *User) {
	user.Room = this
	this.addUser(user)
	this.BroadcastLogin(user)
}

func (this *ChatRoom) SetUserOut(user *User) {
	user.Room = nil
	this.removeUser(user)
	this.BroadcastLogout(user)
}

func (this *ChatRoom) SetUserOffline(user *User) {
	this.offlineUser(user)
	this.BroadcastOffline(user)
}

func (this *ChatRoom) addUser(user *User) {
	this.mu.Lock()
	this.OnlineUsers = append(this.OnlineUsers, user)
	this.OnlineNum++
	this.mu.Unlock()

}

func (this *ChatRoom) removeUser(user *User) {
	this.mu.Lock()
	for i := 0; i < len(this.OnlineUsers); i++ {
		if this.OnlineUsers[i].Uid == user.Uid {
			fmt.Println("remove user ", user.Uid)
			this.OnlineUsers = append(this.OnlineUsers[:i], this.OnlineUsers[i+1:]...)
			this.OnlineNum--
			break
		}
	}
	this.mu.Unlock()
}

func (this *ChatRoom) offlineUser(user *User) {
	this.removeUser(user)
	//this.OfflineUsers = append(this.OfflineUsers, user)
	this.OnlineNum--
	this.OfflineNum++
}

func (this *ChatRoom) BroadcastLogin(user *User) {
	body := msg.LoginInfo{Uid: user.Uid, Name: user.Name, Rid: this.Rid}
	b, _ := msg.Serialize(S2C_BROADCAST_ENTER_ROOM, body)
	for _, u := range this.OnlineUsers {
		if u.Uid != user.Uid {
			u.SendMsg(b)
		}
	}
}

func (this *ChatRoom) BroadcastLogout(user *User) {
	body := msg.LoginInfo{Uid: user.Uid, Name: user.Name, Rid: this.Rid}
	b, _ := msg.Serialize(S2C_BROADCAST_OUT_ROOM, body)
	for _, u := range this.OnlineUsers {
		if u.Uid != user.Uid {
			u.SendMsg(b)
		}
	}
}

func (this *ChatRoom) BroadcastOffline(user *User) {
	body := msg.LoginInfo{Uid: user.Uid, Name: user.Name, Rid: this.Rid}
	b, _ := msg.Serialize(S2C_BROADCAST_OFFLINE, body)
	for _, u := range this.OnlineUsers {
		if u.Uid != user.Uid {
			u.SendMsg(b)
		}
	}
}

func (this *ChatRoom) BroadcastLevelChanged(levelChange *LevelChange) {
	for _, u := range this.OnlineUsers {
		b, _ := msg.Serialize(S2C_BROADCAST_LEVELCHANGED, levelChange)
		u.SendMsg(b)
	}
}

func (this *ChatRoom) ConvertToRoomInfo() msg.RoomInfo {
	others := make([]msg.UserInfo, 0)
	for _, user := range this.OnlineUsers {
		others = append(others, user.ConvertToUserInfo())
	}
	//for _, user := range this.OfflineUsers {
	//	others = append(others, user.ConvertToUserInfo())
	//}

	others = append(others, this.GetUserInfoFromSql(others)...)
	return msg.RoomInfo{Rid: this.Rid, Others: others}
}

func (this *ChatRoom) SendMsg(b []byte) {
	for _, user := range this.OnlineUsers {
		user.SendMsg(b)
	}
}

func (this *ChatRoom) ReceiveMsg(b []byte) {
	fmt.Println("ChatRoom ReceiveMsg ", b)
	this.msg <- b
	t := time.Now().Format("2006-01-02 15:04:05")
	this.msgBuf = append(this.msgBuf, mysql.RoomInfo{SendMsg: string(b), SendTime: t})
	this.msgNum++
	if this.msgNum > MAX_MSG_BUF {
		this.AddMsgToSql(this.msgBuf)
		this.msgNum = 0
		this.msgBuf = make([]mysql.RoomInfo, 0, MAX_MSG_BUF)
	}

}

func (this *ChatRoom) AddMsgToSql(infos []mysql.RoomInfo) {
	for _, info := range infos {
		mysql.AddRoomInfoByRoomId(this.Rid, &info)
	}
}

//查询离线玩家信息
func (this *ChatRoom) GetUserInfoFromSql(onlineUsers []msg.UserInfo) []msg.UserInfo {
	result := make([]msg.UserInfo, 0)
	t := mysql.GetAllUserInfoByRoomId(this.Rid)
	for _, u := range t {
		isOnline := false
		for _, uu := range onlineUsers {
			if uu.Uid == u.Uid {
				isOnline = true //当前玩家已经在线，，但数据库中 状态还没改
			}
		}
		if isOnline {
			continue
		}
		t, _ := strconv.Atoi(u.OfflineTime)
		level, exp, title := UserLevelConversion(t)
		result = append(result, msg.UserInfo{Uid: u.Uid, Name: u.UserName, Level: level, Exp: exp, Title: title, Status: STATUS_OFFLINE})
	}
	return result
}

//获取离线消息
func (this *ChatRoom) GetMsgByTime(t string) []mysql.RoomInfo {
	fmt.Println("ChatRoom GetMsgByTime", t)
	messages := make([]mysql.RoomInfo, 0)
	for _, info := range this.msgBuf {
		offlineTime, _ := time.Parse("2006-01-02 15:04:05", t)         //离线时间
		msgTime, _ := time.Parse("2006-01-02 15:04:05", info.SendTime) //消息发送时间
		if offlineTime.Sub(msgTime) < 0 {
			messages = append(messages, info)
		}
	}
	s := mysql.GetRoomInfoByTimes(t, time.Now().Format("2006-01-02 15:04:05"), this.Rid)
	for _, info := range s {
		messages = append(messages, info)
	}
	return messages
}

func (this *ChatRoom) SendOfflineMsg(user *User) {
	t := user.LogoutTime.Format("2006-01-02 15:04:05") //用户上次离线时间
	if messages := this.GetMsgByTime(t); len(messages) > 0 {
		for _, m := range messages {
			if m.SendMsg != "" {
				user.SendMsg([]byte(m.SendMsg))
			}
		}
	}
}
