package room

import (
	. "../config"
	"../msg"
	"../mysql"
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
	OfflineUsers []*User
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
	for range time.Tick(time.Second * 60) {
		for _, user := range this.OnlineUsers {
			pre := user.Title
			_, _, title := UserLevelConversion(user.TotalTime + time.Now().Second() - user.LoginTime.Second())
			if pre != title {
				this.BroadcastLevelChanged(&LevelChange{Uid: user.Uid, OldTitle: pre, NewTitle: title})
			}
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
	if user.GetRoomId() == this.Rid { //之前就在房间中，发送离线消息
		this.OfflineNum--
		if messages := this.GetMsgByTime(time.Now().Format("2006-01-02 15:04:05")); len(messages) > 0 {
			for _, m := range messages {
				user.SendMsg([]byte(m.SendMsg))
			}
		}
	}
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
	for i := 0; i < this.OnlineNum; i++ {
		if this.OnlineUsers[i].Uid == user.Uid {
			this.OnlineUsers = append(this.OnlineUsers[:i], this.OnlineUsers[i+1:]...)
			this.OnlineNum--
			break
		}
	}
	this.mu.Unlock()
}

func (this *ChatRoom) offlineUser(user *User) {
	this.removeUser(user)
	this.OfflineUsers = append(this.OfflineUsers, user)
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
	others = append(others, this.GetUserInfoFromSql()...)
	return msg.RoomInfo{Rid: this.Rid, Others: others}
}

func (this *ChatRoom) SendMsg(b []byte) {
	for _, user := range this.OnlineUsers {
		user.SendMsg(b)
	}
}

func (this *ChatRoom) ReceiveMsg(b []byte) {
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
func (this *ChatRoom) GetUserInfoFromSql() []msg.UserInfo {
	result := make([]msg.UserInfo, 0)
	t := mysql.GetAllUserInfoByRoomId(this.Rid)
	for _, u := range t {
		t, _ := strconv.Atoi(u.OfflineTime)
		level, exp, title := UserLevelConversion(t)
		result = append(result, msg.UserInfo{Uid: u.Uid, Name: u.UserName, Level: level, Exp: exp, Title: title, Status: STATUS_OFFLINE})
	}
	return result
}

//获取离线消息
func (this *ChatRoom) GetMsgByTime(t string) []mysql.RoomInfo {
	messages := make([]mysql.RoomInfo, 0)
	messages = append(messages, this.msgBuf...)
	s := mysql.GetRoomInfoByTimes(t, time.Now().Format("2006-01-02 15:04:05"))
	for _, info := range s {
		messages = append(messages, info)
	}
	return messages
}
