package main

import (
	"../room"
	"sync"
)

var rooms = map[int]string{
	1: "room_1",
	2: "room_2",
	//3: "room_3",
	//4: "room_4",
	//5: "room_5",
}

func main() {
	manager := &room.Manager{sync.Mutex{}, make([]*room.ChatRoom, 0)}
	manager.CreateChatRooms(rooms)
	manager.Start()
}
