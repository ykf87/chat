package ws

import (
	"fmt"
	"os"
	"time"
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
	Avatar   string `json:"avatar"`                             //群头像
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
	Avatar   string `json:"avatar"`                             //对方头像
}
type Messages struct { //消息列表
	Id      int64  `gorm:"primaryKey;autoIncrement" json:"id"` //消息id
	RoomId  int64  `gorm:""`                                   //房间id
	From    int64  `json:"from" gorm:"index"`                  //发消息的用户
	To      int64  `json:"to" gorm:""`                         //接收消息的用户
	Addtime int64  `json:"addtime" gorm:"not null"`            //添加时间
	Content string `json:"content"`                            //消息内容
	Size    int64  `json:"size"`                               //消息体大小
	Type    int32  `json:"type"`                               //消息类型
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
		RoomId   int64  `json:"room_id"`
		User     int64  `json:"user"`
		Name     string `json:"name"`
		Online   int    `json:"online"`
		Avatar   string `json:"avatar"`
		NoRead   int    `json:"no_read"`
		LastTime int64  `json:"last_time"`
	}{}
	rs := DB.Model(&ChatList{}).Select("chat_lists.room_id", "chat_lists.user", "case chat_lists.room_id when 0 then chat_lists.name else rooms.name end as name", "case chat_lists.room_id when 0 then chat_lists.last_time else rooms.last_time end as last_time", "case chat_lists.room_id when 0 then chat_lists.avatar else rooms.avatar end as avatar", "case chat_lists.room_id when 0 then chat_lists.last_time else rooms.last_time end as last_time").Joins("left join rooms on chat_lists.room_id = rooms.id").Where("chat_lists.uid = ?", this.Id).Order("last_time DESC").Scan(&result)
	if rs.Error != nil {
		logs.Logger.Error(rs.Error)
		return nil
	}
	for k, v := range result {
		if v.User > 0 {
			if uu := GetUser(v.User); uu != nil {
				result[k].Online = 1
				result[k].Avatar = uu.Avatar
				result[k].Name = uu.Name
			}
		}
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
	fmt.Println(msgobj)

	mmm := new(Messages)
	mmm.Addtime = time.Now().Unix()
	mmm.From = this.Id
	mmm.RoomId = msgobj.GetRoom()
	mmm.To = msgobj.GetTo()
	mmm.Size = msgobj.GetData().GetSize()
	mmm.Type = msgobj.GetData().GetType()

	isurl := false
	switch msgobj.GetData().GetType() {
	case 1: //文本消息
		mmm.Content = msgobj.GetData().Data
	case 3: //表情
		mmm.Content = msgobj.GetData().Data
	case 2: //图片消息
		isurl = true
		str, err := OssClient.UploadBase64(msgobj.GetData().Data, "chat/")
		if err != nil {
			logs.Logger.Error(err)
		} else {
			mmm.Content = str
		}
	case 4: //语音消息
		isurl = true
		str, err := OssClient.UploadBase64(msgobj.GetData().Data, "chat/")
		if err != nil {
			logs.Logger.Error(err)
		} else {
			mmm.Content = str
		}
	}
	DB.Create(mmm)
	toUserId := msgobj.GetTo()
	roomId := msgobj.GetRoom()
	var sendTo []*User
	var noticeTo []int64
	if toUserId > 0 {
		toUser := GetUser(msgobj.GetTo())
		if toUser == nil { //聊天对象不在线
			if configs.Conf.OfflineNotice != "" {

			}
			noticeTo = append(noticeTo, toUserId)
		} else {
			sendTo = append(sendTo, toUser)
		}
	} else if roomId > 0 {
		var res []int64
		rs := DB.Model(&ChatList{}).Select("uid").Where("room_id=?", roomId).Scan(&res)
		if rs.Error != nil {
			logs.Logger.Error(rs.Error)
			return
		} else {

		}
	}

	body := new(chat.Body)
	if isurl == true {
		body.Data = OssClient.Url(mmm.Content)
	} else {
		body.Data = mmm.Content
	}

	body.Size = msgobj.GetData().GetSize()
	body.Type = msgobj.GetData().GetType()
	chatobj := new(chat.Chat)
	chatobj.Data = body
	chatobj.From = msgobj.GetFrom()
	chatobj.To = msgobj.GetTo()
	chatobj.Id = msgobj.GetId()
	chatobj.Room = msgobj.GetRoom()

	bt, err := proto.Marshal(chatobj)
	fmt.Println("========================", string(bt), err)
	if err != nil {
		logs.Logger.Error("消息发送失败!", err)
		return
	}
	for _, v := range sendTo {
		v.Conn.WriteMessage(bt)
	}
}
