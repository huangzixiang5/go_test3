package config

import "time"

const (
	IP         = "localhost:50000"
	HEART_TIME = time.Second * 300
)

const (
	C2S_LOGIN      uint8 = 0x001 //请求登录
	C2S_LOGOUT     uint8 = 0x003 //请求登出
	C2S_ENTER_ROOM uint8 = 0x005 //请求进房间
	C2S_OUT_ROOM   uint8 = 0X007 //请求出房间
	C2S_REGISTER   uint8 = 0X009 //注册
	C2S_CHAT       uint8 = 0X00B //聊天消息
	C2S_HEART      uint8 = 0x0d  //心跳包
)

const (
	S2C_LOGIN_SUCCESS          uint8 = 0X002 //登录大厅成功
	S2C_LOGIN_FAIL             uint8 = 0X004 //登录失败
	S2C_LOGOUT_SUCCESS         uint8 = 0X006 //登出成功
	S2C_ENTER_ROOM_SUCCESS     uint8 = 0X008 //进房成功
	S2C_ENTER_ROOM_FAIL        uint8 = 0X00A //进房失败
	S2C_OUT_ROOM_SUCCESS       uint8 = 0X00C //出房间成功
	S2C_REGISTER_SUCCESS       uint8 = 0X00E //注册成功
	S2C_CHAT                   uint8 = 0X010 //聊天消息
	S2C_BROADCAST_ENTER_ROOM   uint8 = 0x012 // 广播玩家进入房间
	S2C_BROADCAST_OUT_ROOM     uint8 = 0x014 // 广播玩家出房间
	S2C_BROADCAST_OFFLINE      uint8 = 0x016 //广播玩家离线
	S2C_BROADCAST_LEVELCHANGED uint8 = 0x018 //广播玩家等级变化

)

const ( //玩家状态
	STATUS_OFFLINE uint8 = 0X01 //离线状态
	STATUS_IN_ROOM uint8 = 0X02 //在房间中
	STATUS_IN_HALL uint8 = 0X03 //在大厅
)

const (
	CHAT_TEXT uint8 = 0x01 //字符串信息
	CHAT_FACE uint8 = 0x02 //表情
	CHAT_PIC  uint8 = 0x03 //图片
)
