package configs

import (
	"flag"
	"os"

	"github.com/jinzhu/configor"
)

var path = flag.String("conf", "./config.json", "指定配置文件路径")
var Conf = struct {
	Port             int    `default:"19192" json:"port"`                //服务端口
	AuthUrl          string ``                                           //客户端连接时的权限验证地址
	OfflineNotice    string ``                                           //被聊天对象不在线的通知
	DbPath           string `default:"db"`                               //数据默认存储路径
	MessageCacheTime int64  `json:"message_cache_time" default:"604800"` //聊天数据缓存时长,单位 秒(默认7天)
	Oss              struct {
		Endpoint    string `required:"true"`
		Url         string `required:"true"`
		Accesskeyid string `required:"true"`
		Secret      string `required:"true"`
		Bucket      string `required:"true"`
		Ssl         bool   `default:"false"`
	} `required:"true" json:"oss"` //文件存储配置
}{}

func init() {
	flag.Parse()
	if _, err := os.Stat(*path); err != nil {
		panic(err)
	}

	if err := configor.Load(&Conf, *path); err != nil {
		panic(err)
	}
}
