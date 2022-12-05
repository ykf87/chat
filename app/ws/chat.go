package ws

import (
	"os"
	"ws-chat/app/configs"
	"ws-chat/app/db"
	"ws-chat/app/logs"
	"ws-chat/app/msg/chat"

	"gorm.io/gorm"

	"github.com/golang/protobuf/proto"
)

type Room struct { //群聊
	Id       int64  `gorm:"primaryKey;autoIncrement" json:"id"` //房间id
	Addtime  int64  `json:"addtime" gorm:"not null"`            //创建时间
	Name     string `json:"name"`                               //房间名称
	Created  int64  `json:"created" gorm:"index"`               //创建人id
	LastTime int64  `json:"last_time"`                          //最后一次消息的时间
	Total    int    `gorm:"default:0" json:"total"`             //群用户总数
}
type ChatList struct { //用户聊天列表
	Id       int64  `json:"id" gorm:"primaryKey;autoIncrement"` //id
	Uid      int64  `gorm:"index" json:"uid"`                   //用户
	RoomId   int64  `gorm:"index;default:0" json:"room_id"`     //房间id
	User     int64  `json:"user" gorm:"default:0"`              //私聊的对象
	Addtime  int64  `json:"addtime" gorm:"not null"`            //添加时间
	Name     string `json:"name"`                               //聊天的名字,如果是私聊,则使用对方名字
	LastTime int64  `json:"last_tile"`                          //最后一次消息的时间
	LastRead int64  `json:"last_read"`                          //最后一次读这个聊天消息的时间
}
type Messages struct { //消息列表
	Id      int64  `gorm:"primaryKey;autoIncrement" json:"id"` //消息id
	RoomId  int64  `gorm:"not null"`                           //房间id
	From    int64  `json:"from" gorm:"index"`                  //发消息的用户
	Addtime int64  `json:"addtime" gorm:"not null"`            //添加时间
	Content string `json:"content"`                            //消息内容
}

var DB *gorm.DB

func init() {
	filename := configs.Conf.DbPath + "/chat.db"
	isfirst := false
	if _, err := os.Stat(filename); err != nil {
		isfirst = true
	}
	db, err := db.Create(configs.Conf.DbPath + "/chat.db")
	if err != nil {
		panic(err)
	}
	DB = db
	if isfirst == true {
		db.AutoMigrate(&Room{})
		db.AutoMigrate(&ChatList{})
		db.AutoMigrate(&Messages{})
	}
}

//用户的聊天消息列表
func (this *User) GetRooms() interface{} {
	var result = []struct {
		RoomId int64  `json:"room_id"`
		User   int64  `json:"user"`
		Name   string `json:"name"`
	}{}
	rs := DB.Model(&ChatList{}).Select("chat_lists.room_id", "chat_lists.user", "case chat_lists.room_id when 0 then chat_lists.name else rooms.name end as name", "case chat_lists.room_id when 0 then chat_lists.last_time else rooms.last_time end as last_time").Joins("left join rooms on chat_lists.room_id = rooms.id").Where("chat_lists.uid = ?", this.Id).Order("last_time DESC").Scan(&result)
	if rs.Error != nil {
		logs.Logger.Error(rs.Error)
		return nil
	}
	return result
}

func (this *User) GetChat(msg []byte) {
	msgobj := chat.Chat{}
	err := proto.Unmarshal(msg, &msgobj)
	if err != nil {
		logs.Logger.Error(err)
		return
	}
	toUser := GetUser(msgobj.GetTo())
	if toUser == nil { //聊天对象不在线
		if configs.Conf.OfflineNotice != "" {

		}
		return
	}
	switch msgobj.GetData().GetType() {
	case 1: //文本消息
	case 2: //图片消息
	case 3: //表情
	case 4: //语音消息
	}
}
