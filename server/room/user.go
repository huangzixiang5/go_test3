package room

import (
	. "../config"
	"../msg"
	"../mysql"
	"fmt"
	"log"
	"net"
	"time"
)

type User struct {
	Uid        int
	Name       string
	Sex        int8 // 0:unknown 1:man 2:woman
	Url        string
	Title      string
	LoginTime  time.Time
	LogoutTime time.Time
	TotalTime  int
	Conn       net.Conn
	Room       *ChatRoom
	Status     uint8
}

func (this *User) SetLogin() {
	this.LoginTime = time.Now()
	this.Status = STATUS_IN_HALL
}

func (this *User) SetLogout() {
	this.LogoutTime = time.Now()
	this.TotalTime = this.TotalTime + this.LogoutTime.Second() - this.LoginTime.Second()
}

func (this *User) SetOffline() {
	this.SetLogout()
	this.UpdateUserTimeInfo()
	if this.Room != nil {
		this.Room.SetUserOffline(this)
	}
}

func (this *User) SetConn(conn net.Conn) {
	this.Conn = conn
}

func (this *User) GetRoomId() int {
	if this.Room != nil {
		return this.Room.Rid
	} else {
		return -1
	}
}

func (this *User) ConvertToUserInfo() msg.UserInfo {
	level, exp, title := UserLevelConversion(this.TotalTime + time.Now().Second() - this.LoginTime.Second())
	return msg.UserInfo{Uid: this.Uid, Name: this.Name, Level: level, Exp: exp, Title: title, Status: STATUS_IN_HALL}
}

func (this *User) SendMsg(b []byte) {
	cmd, data, _ := msg.Deserialize(b)
	_, err := this.Conn.Write(b)
	if err != nil {
		log.Print(err)
	} else {
		fmt.Println("发送成功  ", this.Uid, cmd, data)
	}
}

func (this *User) UpdateUserTimeInfo() {
	Rid := -1
	if this.Room != nil {
		Rid = this.Room.Rid
	}

	t := &mysql.UserInfo{
		Status:        "offline",
		OfflineTime:   this.LogoutTime.Format("2006-01-02 15:04:05"),
		OnlineTime:    this.TotalTime,
		DefaultRoomId: Rid}
	fmt.Println("UpdateUserTimeInfo offline time : ", t)
	mysql.UpDateUserTimeInfo(this.Uid, t)
}
