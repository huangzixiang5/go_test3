package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"time"
)

var user = "root"
var pwd = "root123"
var dbName = "test1"
var userInfoTab = "userinfo"  //用户信息
var hallInfoTab = "hallinfo"  //大厅信息
var roomInfoTab = "roominfo_" //房间信息

type UserInfo struct {
	Uid           int
	UserName      string
	PassWord      string
	HeadPhoto     string
	Status        string
	OnlineTime    int
	OfflineTime   string
	DefaultRoomId int
}

type RoomInfo struct {
	Uid      int
	SendTime string
	SendMsg  string
	RoomName string
}

type HallInfo struct {
	RoomId    int
	RoomName  string
	TotalNum  int
	OnlineNum int
}

//mysql建立连接
func GetDBConnect() *sql.DB {
	db, err := sql.Open("mysql", user+":"+pwd+"@/"+dbName+"?charset=utf8")
	checkErr(err)
	return db
}

//创建信息表(
func CreateBasalTab(userInfoTab string, hallInfoTab string) {
	str1 := "CREATE TABLE " + hallInfoTab + " (`room_id` int(11) NOT NULL,`room_name` varchar(255) NOT NULL,`total_num` int(11) NOT NULL,`online_num` int(11) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	str2 := "CREATE TABLE " + userInfoTab + " (`uid` int(11) NOT NULL,`user_name` varchar(255) NOT NULL,`password` varchar(255) NOT NULL, `head_photo` varchar(255) DEFAULT NULL,`status` varchar(255) DEFAULT NULL,`online_time` int(11) DEFAULT NULL,`offline_time` datetime DEFAULT NULL,`default_roomid` int(11) DEFAULT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	tabSli := []string{str1, str2}
	db := GetDBConnect()
	defer db.Close()
	for i := 0; i < 2; i++ {
		stmt, err := db.Prepare(tabSli[i])
		checkErr(err)
		stmt.Exec()
	}
}
func CreateRoomTabByRoomId(roomId int, roomInfoTab string) {
	str := "CREATE TABLE " + roomInfoTab + strconv.Itoa(roomId) + " (`uid` int(11) NOT NULL,`send_time` datetime NOT NULL,`send_msg` varchar(255) NOT NULL,`room_name` varchar(255) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;"
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(str)
	checkErr(err)
	stmt.Exec()
}

func ExecuteStr(sqlStr string, args []interface{}) {
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(sqlStr)
	checkErr(err)
	_, err = stmt.Exec(args...)
	checkErr(err)
}

//-------------------------------------------------------用户信息表 userInfoTab 的增删改查---------------
//新增用户
func AddUserInfo(userInfo *UserInfo) {
	var ttime time.Time
	if userInfo.OfflineTime != "" {
		ttime, _ = time.Parse("2006-01-02 15:04:05", userInfo.OfflineTime)
		//checkErr(err)
	} else {
		ttime, _ = time.Parse("2006-01-02 15:04:05", "0000-00-00 00:00:00")
		//checkErr(err)
	}

	str := "INSERT " + userInfoTab + " set uid = ?, user_name = ?,password = ?, head_photo = ?, status = ?, online_time = ?, offline_time = ?,default_roomid = ?"
	args := []interface{}{userInfo.Uid, userInfo.UserName, userInfo.PassWord, userInfo.HeadPhoto, userInfo.Status, userInfo.OnlineTime, ttime, userInfo.DefaultRoomId}
	ExecuteStr(str, args)
	log.Println("insert id", userInfo.Uid)
}

//更新用户信息
func UpDateUserInfoByUid(uid int, userInfo *UserInfo) {
	ttime, err := time.Parse("2006-01-02 15:04:05", userInfo.OfflineTime)
	checkErr(err)
	str := "UPDATE " + userInfoTab + " SET `user_name`= ?, `password`=?, `head_photo`=?, `status`=?, `online_time`=?, `offline_time`=?, `default_roomid`=? WHERE `uid`=? "
	args := []interface{}{userInfo.UserName, userInfo.PassWord, userInfo.HeadPhoto, userInfo.Status, userInfo.OnlineTime, ttime, userInfo.DefaultRoomId, uid}
	ExecuteStr(str, args)
}

//
func UpDateUserTimeInfo(uid int, userInfo *UserInfo) {
	ttime, err := time.Parse("2006-01-02 15:04:05", userInfo.OfflineTime)
	checkErr(err)
	str := "UPDATE " + userInfoTab + " set `status`=?, `online_time`=?, `offline_time`=?, `default_roomid`=? WHERE `uid`=? "
	args := []interface{}{userInfo.Status, userInfo.OnlineTime, ttime, userInfo.DefaultRoomId, uid}
	ExecuteStr(str, args)
}

//删除单个用户信息
func DeleteSingleUserInfoByUid(uid int) {
	str := "DELETE FROM " + userInfoTab + " WHERE `uid` =?"
	args := []interface{}{}
	args = append(args, uid)
	ExecuteStr(str, args)
}

//查询单个用户信息
func GetSingleUserInfoByUid(uid int) (result []UserInfo) {
	result = []UserInfo{}
	str := "SELECT * FROM " + userInfoTab + " WHERE `uid`=? LIMIT 1"

	db := GetDBConnect()
	defer db.Close()

	row := db.QueryRow(str, uid)
	var Uid int
	var UserName string
	var PassWord string
	var HeadPhoto string
	var Status string
	var OnlineTime int
	var OfflineTime string
	var DefaultRoomId int
	err := row.Scan(&Uid, &UserName, &PassWord, &HeadPhoto, &Status, &OnlineTime, &OfflineTime, &DefaultRoomId)
	checkErr(err)
	result = append(result, UserInfo{Uid, UserName, PassWord, HeadPhoto, Status, OnlineTime, OfflineTime, DefaultRoomId})
	return result
}

//查询所有用户信息
func GetAllUserInfo() (result []UserInfo) {
	result = []UserInfo{}
	str := "SELECT * FROM " + userInfoTab
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(str)
	checkErr(err)
	rows, err := stmt.Query()
	checkErr(err)
	for rows.Next() {
		var Uid int
		var UserName string
		var PassWord string
		var HeadPhoto string
		var Status string
		var OnlineTime int
		var OfflineTime string
		var DefaultRoomId int
		rows.Scan(&Uid, &UserName, &PassWord, &HeadPhoto, &Status, &OnlineTime, &OfflineTime, &DefaultRoomId)
		result = append(result, UserInfo{Uid, UserName, PassWord, HeadPhoto, Status, OnlineTime, OfflineTime, DefaultRoomId})
	}
	return result
}

//根据房间ID查找所有玩家
func GetAllUserInfoByRoomId(roomId int) (result []UserInfo) {
	result = []UserInfo{}
	str := "SELECT * FROM " + userInfoTab + " WHERE `default_roomid`=? "
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(str)
	checkErr(err)
	rows, err := stmt.Query(roomId)
	checkErr(err)
	for rows.Next() {
		var Uid int
		var UserName string
		var PassWord string
		var HeadPhoto string
		var Status string
		var OnlineTime int
		var OfflineTime string
		var DefaultRoomId int
		rows.Scan(&Uid, &UserName, &PassWord, &HeadPhoto, &Status, &OnlineTime, &OfflineTime, &DefaultRoomId)
		result = append(result, UserInfo{Uid, UserName, PassWord, HeadPhoto, Status, OnlineTime, OfflineTime, DefaultRoomId})
	}
	return result
}

//根据用户的id和密码查询用户信息
func CheckUserInfoByUidAndPassword(uid int, password string) (bool bool, result []UserInfo) {
	result = []UserInfo{}
	str := "SELECT * FROM " + userInfoTab + " WHERE `uid`=? and `password`=? "

	db := GetDBConnect()
	defer db.Close()

	row := db.QueryRow(str, uid, password)
	var Uid int
	var UserName string
	var PassWord string
	var HeadPhoto string
	var Status string
	var OnlineTime int
	var OfflineTime string
	var DefaultRoomId int
	err := row.Scan(&Uid, &UserName, &PassWord, &HeadPhoto, &Status, &OnlineTime, &OfflineTime, &DefaultRoomId)
	if err == nil {
		result = append(result, UserInfo{Uid, UserName, PassWord, HeadPhoto, Status, OnlineTime, OfflineTime, DefaultRoomId})
		return true, result
	}
	return false, result
	//checkErr(err)
}

func CheckUserInfoExistByName(name string) (b bool) {
	str := "SELECT * FROM " + userInfoTab + " WHERE `user_name`=? "
	db := GetDBConnect()
	defer db.Close()
	row := db.QueryRow(str, name)
	var Uid int
	var UserName string
	var PassWord string
	var HeadPhoto string
	var Status string
	var OnlineTime int
	var OfflineTime string
	var DefaultRoomId int
	err := row.Scan(&Uid, &UserName, &PassWord, &HeadPhoto, &Status, &OnlineTime, &OfflineTime, &DefaultRoomId)
	if err == nil {
		return false
	}
	return true
}

func CheckUserInfoByNameAndPassword(name, password string) (b bool, result []UserInfo) {
	result = []UserInfo{}
	str := "SELECT * FROM " + userInfoTab + " WHERE `user_name`=? and `password`=? "

	db := GetDBConnect()
	defer db.Close()

	row := db.QueryRow(str, name, password)
	var Uid int
	var UserName string
	var PassWord string
	var HeadPhoto string
	var Status string
	var OnlineTime int
	var OfflineTime string
	var DefaultRoomId int
	err := row.Scan(&Uid, &UserName, &PassWord, &HeadPhoto, &Status, &OnlineTime, &OfflineTime, &DefaultRoomId)
	if err == nil {
		result = append(result, UserInfo{Uid, UserName, PassWord, HeadPhoto, Status, OnlineTime, OfflineTime, DefaultRoomId})
		return true, result
	}
	return false, result
}

//------------------------------------------------------------------------------------------------

//-------------------------------------------------------大厅信息表 hallInfoTab 的增删改查---------------
func AddHallInfo(hallInfo *HallInfo) {
	str := "INSERT INTO " + hallInfoTab + "(`room_id`, `room_name`,`total_num`, `online_num`) VALUES ( ?,?,?,?)"
	args := []interface{}{hallInfo.RoomId, hallInfo.RoomName, hallInfo.TotalNum, hallInfo.OnlineNum}
	ExecuteStr(str, args)
	log.Println("insert roomid", hallInfo.RoomId)
}
func UpDataHallInfoByRoomId(roomId int, hallInfo *HallInfo) {
	str := "UPDATE " + hallInfoTab + " SET`room_name`=?, `total_num`=?, `online_num`=? WHERE `room_id`=? "
	args := []interface{}{hallInfo.RoomName, hallInfo.TotalNum, hallInfo.OnlineNum, roomId}
	ExecuteStr(str, args)
}
func DeleteSingleHallInfoByRoomId(roomId int) {
	str := "DELETE FROM " + hallInfoTab + " WHERE `room_id` =?"
	args := []interface{}{}
	args = append(args, roomId)
	ExecuteStr(str, args)
}
func GetSingleHallInfoByRoomId(roomId int) (result []HallInfo) {
	result = []HallInfo{}
	str := "SELECT * FROM " + hallInfoTab + " WHERE `room_id`=? LIMIT 1"
	db := GetDBConnect()
	defer db.Close()
	row := db.QueryRow(str, roomId)
	var RoomId int
	var RoomName string
	var TotalNum int
	var OnlineNum int
	err := row.Scan(&RoomId, &RoomName, &TotalNum, &OnlineNum)
	checkErr(err)
	result = append(result, HallInfo{RoomId, RoomName, TotalNum, OnlineNum})
	return result
}
func GetAllHallInfo() (result []HallInfo) {
	result = []HallInfo{}
	str := "SELECT * FROM " + hallInfoTab
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(str)
	checkErr(err)
	rows, err := stmt.Query()
	checkErr(err)
	for rows.Next() {
		var RoomId int
		var RoomName string
		var TotalNum int
		var OnlineNum int
		rows.Scan(&RoomId, &RoomName, &TotalNum, &OnlineNum)
		result = append(result, HallInfo{RoomId, RoomName, TotalNum, OnlineNum})
	}
	return result
}

//------------------------------------------------------------------------------------------------

//-------------------------------------------------------房间信息表 roomInfoTab 的增删改查------------

func AddRoomInfoByRoomId(roomId int, roomInfo *RoomInfo) {
	var tabname string
	var tab = roomInfoTab + strconv.Itoa(roomId)
	db := GetDBConnect()
	defer db.Close()
	err := db.QueryRow("SELECT table_name FROM information_schema.TABLES WHERE table_name = " + tab).Scan(&tabname)
	if err != nil {
		log.Print("该房间之前是空的")
		CreateRoomTabByRoomId(roomId, roomInfoTab)
		return
	} else {
	}

	ttime, err := time.Parse("2006-01-02 15:04:05", roomInfo.SendTime)
	checkErr(err)
	str := "INSERT INTO " + roomInfoTab + strconv.Itoa(roomId) + "(`uid`, `send_time`, `send_msg`, `room_name`) VALUES ( ?,?,?,?)"
	args := []interface{}{roomInfo.Uid, ttime, roomInfo.SendMsg, roomInfo.RoomName}
	ExecuteStr(str, args)
	log.Println("insert roommsg", roomInfo.SendMsg)
}

func UpdataRoomInfoBySendTime(roomId int, sendTime string, roomInfo *RoomInfo) {
	ttime, err := time.Parse("2006-01-02 15:04:05", sendTime)
	checkErr(err)
	str := "UPDATE " + roomInfoTab + strconv.Itoa(roomId) + " SET `uid`= ?, `room_name`=?, `send_msg`=? WHERE `send_time`=? "
	args := []interface{}{roomInfo.Uid, roomInfo.RoomName, roomInfo.SendMsg, ttime}
	ExecuteStr(str, args)
}
func DeleteSingleRoomInfoBySendTime(roomId int, sendTime string) {
	ttime, err := time.Parse("2006-01-02 15:04:05", sendTime)
	checkErr(err)
	str := "DELETE FROM " + roomInfoTab + strconv.Itoa(roomId) + " WHERE `send_time` =?"
	args := []interface{}{}
	args = append(args, ttime)
	ExecuteStr(str, args)
}
func GetSingleRoomInfoBySendTime(roomId int, sendTime string) (result []RoomInfo) {
	result = []RoomInfo{}
	ttime, err := time.Parse("2006-01-02 15:04:05", sendTime)
	checkErr(err)
	str := "SELECT * FROM " + roomInfoTab + strconv.Itoa(roomId) + " WHERE `send_time`=? LIMIT 1"
	db := GetDBConnect()
	defer db.Close()
	row := db.QueryRow(str, ttime)
	var Uid int
	var SendTime string
	var SendMsg string
	var RoomName string
	row.Scan(&Uid, &SendTime, &SendMsg, &RoomName)
	result = append(result, RoomInfo{Uid, SendTime, SendMsg, RoomName})
	return result
}
func GetAllRoomInfo(roomId int) (result []RoomInfo) {
	result = []RoomInfo{}
	str := "SELECT * FROM " + roomInfoTab + strconv.Itoa(roomId)
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(str)
	checkErr(err)
	rows, err := stmt.Query()
	checkErr(err)
	for rows.Next() {
		var Uid int
		var SendTime string
		var SendMsg string
		var RoomName string
		rows.Scan(&Uid, &SendTime, &SendMsg, &RoomName)
		result = append(result, RoomInfo{Uid, SendTime, SendMsg, RoomName})
	}
	return result
}
func GetRoomInfoByTimes(time1 string, time2 string, roomId int) (result []RoomInfo) {
	result = []RoomInfo{}
	t1, err := time.Parse("2006-01-02 15:04:05", time1)
	checkErr(err)
	t2, err := time.Parse("2006-01-02 15:04:05", time2)
	checkErr(err)

	id := strconv.Itoa(roomId)
	str := "SELECT * FROM " + roomInfoTab + id + " WHERE send_time BETWEEN ? AND ?"
	db := GetDBConnect()
	defer db.Close()
	stmt, err := db.Prepare(str)
	checkErr(err)
	rows, err := stmt.Query(t1, t2)
	checkErr(err)
	for rows.Next() {
		var Uid int
		var SendTime string
		var SendMsg string
		var RoomName string
		rows.Scan(&Uid, &SendTime, &SendMsg, &RoomName)
		result = append(result, RoomInfo{Uid, SendTime, SendMsg, RoomName})
	}
	return result
}

//------------------------------------------------------------------------------------------------

func checkErr(err error) {
	if err != nil {
		log.Println("mysql", err)
	}
}

func main() {
	CreateBasalTab(userInfoTab, hallInfoTab) //创建用户信息表和大厅信息表，需要初始创建
	//CreateRoomTabByRoomId(2, roomInfoTab) // 创建具体聊天室表，不用单独创建，新增房间信息的时候会创建

	//用户信息操作接口：
	//AddUserInfo(userinfo)
	//UpDateUserInfoByUid(id, userinfo)
	//DeleteSingleUserInfoByUid(id)
	//GetSingleUserInfoByUid(id)
	//GetAllUserInfo()

	//大厅信息操作接口：
	//AddHallInfo(hallInfo)
	//UpDataHallInfoByRoomId(roomid hallinfo)
	//DeleteSingleHallInfoByRoomId(roomid)
	//GetSingleHallInfoByRoomId(roomid)
	//GetAllHallInfo()

	//具体房间信息操作接口：
	//AddRoomInfoByRoomId(id.roominfo)
	//UpdataRoomInfoBySendTime(roomid, sendtime, roominfo)
	//DeleteSingleRoomInfoBySendTime(roomid,sendtime)
	//getRoomInfoByTimes(time1, time2) //时间段内的房间信息
	//GetSingleRoomInfoBySendTime(roomid,sentime)
	//GetAllRoomInfo(roomid)

	//room1 := new(RoomInfo)
	//room1.RoomName = "qweqwe"
	//room1.Uid = 1
	//room1.SendTime = time.Now().Format("2006-01-02 15:04:05")
	//room1.SendMsg = "hello world"
	//AddRoomInfoByRoomId(5, room1)
	//a, b := CheckUserInfoByUidAndPassword(2, "789789")
	//log.Println(a, b[0].UserName)
	//log.Println(a.FieldByName("Uid"))
}
