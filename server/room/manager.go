package room

import (
	. "../config"
	"../msg"
	"../mysql"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Manager struct {
	Mu    sync.Mutex
	Rooms []*ChatRoom
}

var id int

func (m *Manager) GetUniqueId() int {
	id++
	return id
}

func (m *Manager) CreateChatRooms(r map[int]string) {
	for id, name := range r {
		CreateAndStartChatRoom(m, id, name)
	}
}

func (m *Manager) StartHeartBeat(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(HEART_TIME))
}

func (m *Manager) StopHeartBeat(conn net.Conn) {
	conn.SetDeadline(time.Time{})
}

func (m *Manager) SendMsgByConn(conn net.Conn, cmd uint8, v interface{}) {
	t, err := msg.Serialize(cmd, v)
	if err != nil {
		log.Print(err)
	}
	_, err = conn.Write(t)
	if err != nil {
		log.Print(err)
	} else {
		fmt.Println("SendMsgByConn : ", cmd, v)
	}
}

//开启服务， 监听连接
func (m *Manager) Start() {
	log.Print("server start :", IP)
	listen, err := net.Listen("tcp", IP)
	if err != nil {
		log.Print(err)
	} else {
		for {
			conn, _ := listen.Accept()
			m.StartHeartBeat(conn)
			go OnLogin(m, conn)
		}
	}
}

//添加一个聊天室
func (m *Manager) AddRoom(r *ChatRoom) {
	m.Rooms = append(m.Rooms, r)
	mysql.AddRoomInfoByRoomId(r.Rid, &mysql.RoomInfo{})
}

// 根据ID获取聊天室
func (m *Manager) GetRoomById(id int) *ChatRoom {
	for i := 0; i < len(m.Rooms); i++ {
		if m.Rooms[i].Rid == id {
			return m.Rooms[i]
		}
	}
	return nil
}

// 获取大厅的简略信息，，包含所有房间信息等
func (m *Manager) GetHallInfo() []msg.SimpleRoomInfo {
	rooms := make([]msg.SimpleRoomInfo, 0)
	for _, room := range m.Rooms {
		s := msg.SimpleRoomInfo{Rid: room.Rid, Name: room.Name, OnlineNum: room.OnlineNum, OfflineNum: room.OfflineNum}
		rooms = append(rooms, s)
	}
	return rooms
}

//检查登录信息，，封装user对象
func (m *Manager) CheckLoginInfo(info msg.LoginInfo) (bool, *User) {
	fmt.Println("checkLoginInfo",info)
	a, b := mysql.CheckUserInfoByNameAndPassword(info.Name, info.Password)
	if a {
		fmt.Println("CheckLoginInfo offlineTime", b[0].OfflineTime)
		offlineTime, err := time.Parse("2006-01-02 15:04:05", b[0].OfflineTime)
		fmt.Println(err)
		fmt.Println("CheckLoginInfo offlineTime", offlineTime)
		return true, &User{Uid: b[0].Uid,
			Name:       b[0].UserName,
			Url:        b[0].HeadPhoto,
			TotalTime:  b[0].OnlineTime,
			Room:       m.GetRoomById(b[0].DefaultRoomId),
			LogoutTime: offlineTime,
		}
	}
	return false, &User{}
}

//检查是否在
func (m *Manager) CheckRoomInfo(info msg.LoginInfo) (bool, *ChatRoom) {
	for _, room := range m.Rooms {
		if room.Rid == info.Rid {
			return true, room
		}
	}
	return false, nil
}

//玩家注册信息，加载到数据库
func (m *Manager) CheckRegisterInfo(info msg.LoginInfo,notExist bool) (bool, *User) {
	if notExist{
		info.Uid = m.GetUniqueId()
		u := mysql.UserInfo{Uid: info.Uid, PassWord: info.Password, UserName: info.Name}
		mysql.AddUserInfo(&u)
		fmt.Println("add user to sql", info)
		return true, &User{Uid: info.Uid, Name: info.Name}
	}
	return false, &User{Uid: info.Uid, Name: info.Name}
}

func (m *Manager) GetUserByIdFromSql(uid int) (result *User) {
	t := mysql.GetSingleUserInfoByUid(uid)
	tt := t[0]
	result = new(User)
	result = &User{Uid: tt.Uid, Name: tt.UserName, Url: tt.HeadPhoto, TotalTime: tt.OnlineTime, Room: m.GetRoomById(tt.DefaultRoomId)}
	return
}
