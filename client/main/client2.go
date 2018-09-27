package main

import (
	. "../config"
	"../msg"
	"fmt"
	"net"
)

var myInfo msg.UserInfo
var otherInfo msg.UserInfo

func main() {
	conn, _ := net.Dial("tcp", IP)
	fmt.Println("连接成功")

	go SendMsg(conn)
	go ReceiveMsg(conn)
	select {}
}

func SendMsg(conn net.Conn) {
	registerInfo := msg.LoginInfo{
		Name:     "bbb",
		Password: "bbb",
	}
	send, _ := msg.Serialize(C2S_LOGIN, registerInfo)
	conn.Write(send)
	var input string
	for {
		fmt.Scanln(&input)
		switch input {
		default:
			if len(input) != 0 {
				inpubyte := []byte(input)
				chat := msg.Chat{
					Name:    "aaa",
					Content: inpubyte,
				}
				send, _ := msg.Serialize(C2S_CHAT, chat)
				conn.Write(send)
			}
		}
		input = ""
	}
}

func ReceiveMsg(conn net.Conn) {
	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			return
		}
		cmd, body, err := msg.Deserialize(b[:n])
		if err != nil {
			fmt.Println(err)
		}
		switch cmd {
		case C2S_CHAT:
			fmt.Println(myInfo.Name+": ", string(body.(msg.Chat).Content))
		}
	}
}

func StartLogin(conn net.Conn) {

}
