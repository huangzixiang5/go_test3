package main

import (
	. "../config"
	"../msg"
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	//打开连接:
	conn, err := net.Dial("tcp", "localhost:50000")
	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		fmt.Println("Error dialing", err.Error())
		return // 终止程序
	}

	inputReader := bufio.NewReader(os.Stdin)
	// fmt.Printf("CLIENTNAME %s", clientName)
	// 给服务器发送信息直到程序退出：
	for {
		fmt.Println("What to send to the server? Type Q to quit.")
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\r\n")
		if trimmedInput == "Q" {
			return
		}

		id, _ := strconv.Atoi(trimmedInput)
		///登录
		t := msg.LoginInfo{
			Uid:      id,
			Name:     "EricHuang",
			Password: "password",
		}
		tt, _ := msg.Serialize(C2S_LOGIN, t)
		_, err = conn.Write(tt)
		time.Sleep(time.Second)
		//进房间
		b := msg.LoginInfo{
			Uid:      id,
			Rid:      1,
			Name:     "EricHuang",
			Password: "password",
		}
		bb, _ := msg.Serialize(C2S_ENTER_ROOM, b)
		_, err = conn.Write(bb)

		time.Sleep(time.Second)
		//发消息
		v := msg.Chat{Uid: id, Name: "EricHuang", Type: 1, Content: []byte("hello")}
		vv, _ := msg.Serialize(C2S_CHAT, v)
		_, err = conn.Write(vv)
	}
}
