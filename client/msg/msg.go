package msg

import (
	. "../config"
	"encoding/json"
	"errors"
	"fmt"
)

//玩家信息
type UserInfo struct {
	Uid    int
	Name   string
	Level  int
	Exp    int
	Title  string
	Status uint8
}

type LevelChange struct {
	Uid      int
	OldTitle string
	NewTitle string
}

type SimpleRoomInfo struct {
	Rid        int
	Name       string
	OnlineNum  int
	OfflineNum int
}

//登录成功后返回对应房间列表信息
type HallInfo struct {
	UserInfo
	Rooms []SimpleRoomInfo
}

//登录
type LoginInfo struct {
	Uid      int
	Rid      int
	Name     string
	Password string
}

//进入房间后，携带房间内用户信息
type RoomInfo struct {
	UserInfo
	Others []UserInfo
	Rid    int
}

//聊天信息
type Chat struct {
	Uid     int
	Name    string
	Type    int8
	Content []byte
}

///////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////

func SplitMsg(msg []byte) (cmd uint8, data []byte) {
	k := msg[0:1]
	data = msg[1:]
	cmd = uint8(k[0])
	return
}

func Serialize(cmd uint8, v interface{}) ([]byte, error) {
	s, err := json.Marshal(v)
	if err == nil {
		s = append([]byte(string(cmd)), s...)
		return s, nil
	}
	return s, err
}

func Deserialize(data []byte) (cmd uint8, v interface{}, err error) {
	var msg []byte
	cmd, msg = SplitMsg(data)
	switch cmd {
	case S2C_BROADCAST_LEVELCHANGED:
		t := LevelChange{}
		err = json.Unmarshal(msg, &t)
		v = t
	case S2C_LOGIN_SUCCESS:
		t := HallInfo{}
		err = json.Unmarshal(msg, &t)
		v = t
	case S2C_ENTER_ROOM_SUCCESS:
		t := RoomInfo{}
		err = json.Unmarshal(msg, &t)
		v = t
	case S2C_CHAT:
		t := Chat{}
		err = json.Unmarshal(msg, &t)
		v = t
		//case S2C_LOGIN_FAIL:
		//case S2C_LOGOUT_SUCCESS:
		//case S2C_ENTER_ROOM_FAIL:
		//case S2C_OUT_ROOM_SUCCESS:
		//case S2C_REGISTER_SUCCESS:

	case C2S_LOGIN:
		t := LoginInfo{}
		json.Unmarshal(msg, &t)
		v = t
	case C2S_LOGOUT:
	case C2S_ENTER_ROOM:
		t := LoginInfo{}
		err = json.Unmarshal(msg, &t)
		v = t
	case C2S_REGISTER:
		t := LoginInfo{}
		err = json.Unmarshal(msg, &t)
		v = t
	case C2S_OUT_ROOM:
	case C2S_CHAT:
		t := Chat{}
		err = json.Unmarshal(msg, &t)
		v = t
	default:
		err = errors.New("error cmd ")
	}
	return
}

func main() {

	login := LoginInfo{Uid: 123, Password: "123", Name: "jkl"}
	b, err := Serialize(C2S_LOGIN, login)
	if err == nil {
		fmt.Println("序列化", b)
	}

	cmd, v, err := Deserialize(b)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("反序列化", cmd, v)
	}

}
