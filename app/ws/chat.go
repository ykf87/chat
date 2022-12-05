package ws

import (
	"ws-chat/app/configs"
	"ws-chat/app/logs"
	"ws-chat/app/msg/chat"

	"github.com/golang/protobuf/proto"
)

type Chat struct {
	Id     int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomId string `gorm:"roomid"`
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
