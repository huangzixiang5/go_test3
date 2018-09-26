package room

import (
	. "../config"
	"../msg"
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
}

func (this *User) SetLogout() {
	this.LogoutTime = time.Now()
	this.TotalTime = this.LogoutTime.Second() - this.LoginTime.Second()
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
	fmt.Println("发送  ", this.Uid, cmd, data)

	_, err := this.Conn.Write(b)
	if err != nil {
		log.Print(err)
	}
}
