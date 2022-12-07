package ws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"ws-chat/app/configs"
	"ws-chat/app/funcs"
	"ws-chat/app/logs"
	"ws-chat/app/oss"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

// type User struct {
// 	Id       int64        `json:"id"`        //用户id
// 	Name     string       `json:"name"`      //用户昵称
// 	Avatar   string       `json:"avatar"`    //头像地址
// 	Conn     *Connections `json:"-"`         //用户ws
// 	ConnTime int64        `json:"conn_time"` //连接时间
// 	PingTime int64        `json:"ping_time"`
// }

var Users sync.Map
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}
var OssClient oss.Oss

func init() {
	var err error
	OssClient, err = oss.GetOss("minio")
	if err != nil {
		panic(err)
	}
}

func checkOrigin(r *http.Request) bool {
	return true
}

func Start() {
	if configs.Conf.AuthUrl == "" {
		panic("请配置 AuthUrl")
	}

	r := gin.Default()
	r.GET("/conn", Connect)
	r.GET("/api/chatlist", func(c *gin.Context) {
		uid, _ := strconv.Atoi(c.Query("id"))
		user := GetUser(int64(uid))
		if user == nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 404,
				"msg":  "not found",
				"data": nil,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "",
			"data": user.GetRooms(),
		})
	})
	r.GET("/api/readmsg")

	port := fmt.Sprintf(":%d", configs.Conf.Port)
	fmt.Println("监听端口:", port)

	go func() {
		for {
			time.Sleep(time.Second * 40)
			now := time.Now().Unix()
			Users.Range(func(key, val interface{}) bool {
				if val == nil {
					return false
				}
				u := val.(*User)
				if u.PingTime+60 < now {
					logs.Logger.Infoln(u.Id, ": 连接超时!")
					u.Remove()
				}
				return true
			})
		}
	}()

	err := r.Run(port)
	if err != nil {
		fmt.Println(err)
	}
}

func Connect(c *gin.Context) {
	u, e := checkToken(c.Writer, c.Request)
	if e != nil {
		return
	}

	var (
		err    error
		WsConn *websocket.Conn
		conn   *Connections
	)
	if WsConn, err = upgrader.Upgrade(c.Writer, c.Request, nil); err != nil {
		logs.Logger.Error(err)
		return
	}
	defer WsConn.Close()
	if conn, err = InitConnect(WsConn); err != nil {
		logs.Logger.Error(err)
		return
	}

	logs.Logger.Debug(fmt.Sprintf("%d: 登录", u.Id))
	u.Conn = conn
	u.ConnTime = time.Now().Unix()
	u.PingTime = time.Now().Unix()
	u.Ipaddr = c.ClientIP()
	defer u.Remove()
	u.Listen()
}

//有连接请求时,检查请求的合法性
func checkToken(w http.ResponseWriter, r *http.Request) (*User, error) {
	token := getToken(r)
	if token == "" {
		w.WriteHeader(404)
		return nil, errors.New("Token 未设置")
	}
	platform := getPlatform(r, "3")

	w.Header().Set("content-type", "text/json")
	headers := make(map[string]string)
	headers["authorization"] = token
	headers["platform"] = platform
	url := configs.Conf.AuthUrl
	if strings.Contains(url, "?") == true {
		url = url + "&_from=ws"
	} else {
		url = url + "?_from=ws"
	}

	rs, err := funcs.Request("GET", url, nil, headers, "")
	if err != nil {
		logs.Logger.Error(err)
		w.Write(rs)
		return nil, err
	}

	obj := gjson.ParseBytes(rs).Map()
	if obj["code"].Exists() && obj["code"].Int() == 200 {
		u := new(User)
		err := json.Unmarshal([]byte(obj["data"].String()), u)
		if err != nil {
			logs.Logger.Error(err)
			w.Write([]byte(`{"data": null, "msg": "", "code": 500}`))
			return nil, err
		}
		return u, nil
	}
	w.Write(rs)
	return nil, errors.New("返回的参数不正确")
}

//获取连接请求的token参数
func getToken(r *http.Request) string {
	var token string

	for k, v := range r.Header {
		if k == "Authorization" || k == "Token" {
			token = v[0]
			break
		}
	}
	if token == "" {
		if tks, ok := r.URL.Query()["token"]; ok {
			token = tks[0]
		}
	}
	if token != "" {
		if strings.Contains(token, "Bearer ") == false {
			token = "Bearer " + token
		}
	}
	return token
}

//获取连接请求的platform参数
func getPlatform(r *http.Request, df string) string {
	var platform string

	for k, v := range r.Header {
		if k == "Platform" {
			platform = v[0]
			break
		}
	}
	if platform == "" {
		if tks, ok := r.URL.Query()["platform"]; ok {
			platform = tks[0]
		}
	}
	if platform == "" {
		platform = df
	}
	return platform
}

func SetUser(id int64, name, avatar string) *User {
	ou := GetUser(id)
	if ou != nil {
		ou.Remove()
	}
	u := new(User)
	u.Id = id
	u.Name = name
	u.Avatar = avatar
	Users.Store(id, u)
	return u
}

func GetUser(id int64) *User {
	u, ok := Users.Load(id)
	if !ok {
		return nil
	}
	return u.(*User)
}

var pingbyte = []byte("ping")

func (this *User) Listen() {
	Users.Store(this.Id, this)
	for {
		msg, err := this.Conn.ReadMessage()
		if err != nil {
			// logs.Logger.Error(err)
			break
		}
		this.PingTime = time.Now().Unix()
		if bytes.Equal(msg, pingbyte) {
			this.Conn.WriteMessage([]byte("pong"))
		} else {
			this.GetChat(msg)
		}
	}
}

func (this *User) Remove() {
	this.Conn.Close()
	Users.Delete(this.Id)
	logs.Logger.Debug(fmt.Sprintf("%d: 退出", this.Id))
}
