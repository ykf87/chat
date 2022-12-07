package ws

import (
	"errors"
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

type User struct {
	Id       int64        `json:"id" gorm:"primaryKey"` //用户id
	Name     string       `json:"name"`                 //用户昵称
	Avatar   string       `json:"avatar"`               //头像地址
	Ipaddr   string       `json:"ipaddr"`               //IP地址
	ConnTime int64        `json:"conn_time" gorm:"-"`   //连接时间
	PingTime int64        `json:"ping_time" gorm:"-"`   //最后一次ping的时间
	Conn     *Connections `json:"-" gorm:"-"`           //用户ws
}
type Room struct {
	Id       int64  `gorm:"primaryKey;autoIncrement" json:"id"` //房间id
	Addtime  int64  `json:"addtime" gorm:"not null"`            //创建时间
	Name     string `json:"name"`                               //房间名称
	Created  int64  `json:"created" gorm:"index"`               //创建人id
	LastTime int64  `json:"last_time"`                          //最后一次消息的时间
	Total    int    `gorm:"default:2" json:"total"`             //群用户总数,如果是2说明是私聊
	Avatar   string `json:"avatar"`                             //群头像
}
type RoomUser struct {
	Uid      int64 `json:"uid" gorm:"primaryKey;not null"`     //用户id
	RoomId   int64 `json:"room_id" gorm:"primaryKey;not null"` //房间id
	Addtime  int64 `json:"addtime" gorm:"not null"`            //添加时间
	To       int64 `json:"to" gorm:"default:0;index"`          //私聊的对象uid,群聊为0
	ReadTime int64 `json:"read_time" gorm:"default:0"`         //消息读取时间
}
type Messages struct { //消息列表
	Id      int64  `gorm:"primaryKey;autoIncrement" json:"id"` //消息id
	RoomId  int64  `gorm:"not null;index"`                     //房间id
	From    int64  `json:"from" gorm:"index"`                  //发消息的用户
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
		db.AutoMigrate(&User{})
		db.AutoMigrate(&Room{})
		db.AutoMigrate(&RoomUser{})
		db.AutoMigrate(&Messages{})
	}
}

type ListRes struct {
	RoomId   int64  `json:"room_id"`
	To       int64  `json:"user"`
	Name     string `json:"name"`
	Online   int    `json:"online"`
	Avatar   string `json:"avatar"`
	NoRead   int64  `json:"no_read"`
	Total    int64  `json:"total"`
	LastTime int64  `json:"last_time"`
	ReadTime int64  `json:"read_time"`
}

//用户的聊天消息列表
func (this *User) GetRooms() interface{} {
	var result []*ListRes
	// rs := DB.Model(&ChatList{}).Select("chat_lists.room_id", "chat_lists.user", "case chat_lists.room_id when 0 then chat_lists.name else rooms.name end as name", "case chat_lists.room_id when 0 then chat_lists.last_time else rooms.last_time end as last_time", "case chat_lists.room_id when 0 then chat_lists.avatar else rooms.avatar end as avatar", "case chat_lists.room_id when 0 then chat_lists.last_time else rooms.last_time end as last_time").Joins("left join rooms on chat_lists.room_id = rooms.id").Where("chat_lists.uid = ?", this.Id).Order("last_time DESC").Scan(&result)
	// rs := DB.Model(&RoomUser{}).Select("room_users.uid", "room_users.room_id", "room_users.readtime", "rooms.total", "rooms.name", "rooms.avatar", "messages.count(id) as noread").Joins("left join rooms on room_users.room_id = rooms.id").Joins("left join messages on messages.room_id = room_users.room_id").Where("room_users.id=?", this.Id).Order("rooms.last_time DESC").Scan(&result)
	rs := DB.Model(&RoomUser{}).Select("room_users.uid", "room_users.read_time", "room_users.room_id", "room_users.`to`", "rooms.total", "rooms.total", "rooms.name", "rooms.avatar", "rooms.last_time").Joins("left join rooms on room_users.room_id = rooms.id").Where("room_users.uid=?", this.Id).Order("rooms.last_time DESC").Scan(&result)
	if rs.Error != nil {
		logs.Logger.Error(rs.Error)
		return nil
	}

	needuser := make(map[int64]*ListRes)
	var needuserid []int64
	for _, v := range result {
		if v.To > 0 {
			if uu := GetUser(v.To); uu != nil {
				v.Online = 1
				v.Avatar = uu.Avatar
				v.Name = uu.Name
			} else {
				needuser[v.To] = v
				needuserid = append(needuserid, v.To)
			}
			DB.Model(&Messages{}).Where("room_id=?", v.RoomId).Where("addtime>?", v.ReadTime).Count(&v.NoRead)
		}
	}
	if len(needuserid) > 0 {
		var uuuus []*User
		rs := DB.Model(&User{}).Where("id in ?", needuserid).Scan(&uuuus)
		if rs.Error == nil {
			for _, v := range uuuus {
				if rsv, ok := needuser[v.Id]; ok {
					rsv.Avatar = v.Avatar
					rsv.Name = v.Name
				}
			}
		}
	}
	return result
}

//读取聊天消息
func (this *User) ReadMsg(roomid int64) {

}

func (this *User) GetChat(msg []byte) {
	fmt.Println("from: ", msg)
	// this.Conn.WriteMessage([]byte("123"))
	// return
	msgobj := &chat.Chat{}
	err := proto.Unmarshal(msg, msgobj)
	if err != nil {
		logs.Logger.Error(err)
		return
	}

	if msgobj.GetData().Data == "" {
		if msgobj.GetRoom() > 0 {
			DB.Model(&RoomUser{}).Where("uid=?", this.Id).Where("room_id=?", msgobj.GetRoom()).UpdateColumn("read_time", time.Now().Unix())
			return
		}
	}

	content := this.FmtChat(msgobj)
	roomid := msgobj.GetRoom()

	if msgobj.GetTo() > 0 { //私聊
		roomid = this.SingleChat(msgobj, content)
	} else if msgobj.GetRoom() > 0 { //群聊

	}

	mmm := new(Messages)
	mmm.Addtime = time.Now().Unix()
	mmm.From = this.Id
	mmm.RoomId = roomid
	mmm.Size = msgobj.GetData().GetSize()
	mmm.Type = msgobj.GetData().GetType()
	mmm.Content = content
	DB.Create(mmm)

	// toUserId := msgobj.GetTo()
	// roomId := msgobj.GetRoom()
	// var sendTo []*User
	// var noticeTo []int64
	// if toUserId > 0 {
	// 	toUser := GetUser(msgobj.GetTo())
	// 	if toUser == nil { //聊天对象不在线
	// 		if configs.Conf.OfflineNotice != "" {

	// 		}
	// 		noticeTo = append(noticeTo, toUserId)
	// 	} else {
	// 		sendTo = append(sendTo, toUser)
	// 	}
	// } else if roomId > 0 {
	// 	var res []int64
	// 	rs := DB.Model(&ChatList{}).Select("uid").Where("room_id=?", roomId).Scan(&res)
	// 	if rs.Error != nil {
	// 		logs.Logger.Error(rs.Error)
	// 		return
	// 	} else {

	// 	}
	// }

	// body := new(chat.Body)
	// if isurl == true {
	// 	body.Data = OssClient.Url(mmm.Content)
	// } else {
	// 	body.Data = mmm.Content
	// }

	// body.Size = msgobj.GetData().GetSize()
	// body.Type = msgobj.GetData().GetType()
	// chatobj := new(chat.Chat)
	// chatobj.Data = body
	// chatobj.From = msgobj.GetFrom()
	// chatobj.To = msgobj.GetTo()
	// chatobj.Id = msgobj.GetId()
	// chatobj.Room = msgobj.GetRoom()

	// bt, err := proto.Marshal(chatobj)
	// fmt.Println("========================", string(bt), err)
	// if err != nil {
	// 	logs.Logger.Error("消息发送失败!", err)
	// 	return
	// }
	// for _, v := range sendTo {
	// 	v.Conn.WriteMessage(bt)
	// }
}

func (this *User) IsUploaded(tp int32) bool {
	switch tp {
	case 1: //文本消息
		return false
	case 2: //图片消息
		return true
	case 3: //表情
		return false
	case 4: //语音消息
		return true
	}
	return false
}

func (this *User) FmtChat(msgobj *chat.Chat) (content string) {
	if this.IsUploaded(msgobj.GetData().GetType()) == true {
		str, err := OssClient.UploadBase64(msgobj.GetData().Data, "chat/")
		if err != nil {
			logs.Logger.Error(err)
		} else {
			content = str
		}
	} else {
		content = msgobj.GetData().Data
	}
	return
}

//私聊
func (this *User) SingleChat(msg *chat.Chat, content string) (roomid int64) {
	to := msg.GetTo()
	if to < 1 {
		return
	}
	result := new(RoomUser)
	rs := DB.Model(&RoomUser{}).Where("uid = ?", this.Id).Where("`to` = ?", to).Scan(result)
	if rs.Error != nil {
		roomObj := new(Room)
		roomObj.Addtime = time.Now().Unix()
		roomObj.Created = this.Id
		rs := DB.Create(roomObj)
		if rs.Error != nil {
			return
		}
		ru := new(RoomUser)
		ru.Uid = to
		ru.RoomId = roomObj.Id
		ru.Addtime = time.Now().Unix()
		ru.To = this.Id

		ru2 := new(RoomUser)
		ru2.Uid = this.Id
		ru2.RoomId = roomObj.Id
		ru2.Addtime = time.Now().Unix()
		ru2.To = to

		var ccccc []*RoomUser
		ccccc = append(ccccc, ru2)
		ccccc = append(ccccc, ru)

		DB.Create(ccccc)
		roomid = roomObj.Id
	} else {
		roomid = result.RoomId
	}

	if user := GetUser(to); user != nil {
		user.Response(msg, content)
	}
	return
}

func (this *User) Response(msg *chat.Chat, content string) error {
	if this.Conn == nil {
		return errors.New(fmt.Sprintf("%d - User not connect!", this.Id))
	}
	body := new(chat.Body)
	if this.IsUploaded(msg.GetData().GetType()) == true {
		body.Data = OssClient.Url(content)
	} else {
		body.Data = content
	}

	body.Size = msg.GetData().GetSize()
	body.Type = msg.GetData().GetType()
	chatobj := new(chat.Chat)
	chatobj.Data = body
	chatobj.From = msg.GetFrom()
	chatobj.To = msg.GetTo()
	chatobj.Id = msg.GetId()
	chatobj.Room = msg.GetRoom()

	bt, err := proto.Marshal(chatobj)
	fmt.Println("tooo: ", bt)
	fmt.Println(msg.String())
	fmt.Println(chatobj.String())
	if err != nil {
		return err
	}
	return this.Conn.WriteMessage(bt)
}
