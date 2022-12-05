package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"ws-chat/app/configs"
	"ws-chat/app/funcs"
	"ws-chat/app/logs"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type User struct {
	Id       int64        `json:"id"`        //用户id
	Name     string       `json:"name"`      //用户昵称
	Avatar   string       `json:"avatar"`    //头像地址
	Conn     *Connections `json:"-"`         //用户ws
	OutConn  chan byte    `json:"-"`         //退出协程
	ConnTime int64        `json:"conn_time"` //连接时间
}

var Users sync.Map
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(r *http.Request) bool {
	return true
}

func Start() {
	if configs.Conf.AuthUrl == "" {
		panic("请配置 AuthUrl")
	}

	http.HandleFunc("/conn", Connect)

	port := fmt.Sprintf(":%d", configs.Conf.Port)
	fmt.Println("监听端口:", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func Connect(w http.ResponseWriter, r *http.Request) {
	u, e := checkToken(w, r)
	if e != nil {
		return
	}

	var (
		err    error
		WsConn *websocket.Conn
		conn   *Connections
	)
	if WsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
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
	u.OutConn = make(chan byte)
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

func (this *User) Listen() {
	Users.Store(this.Id, this)
	go func() {
		for {
			time.Sleep(time.Second * 30)
			if time.Now().Unix()-30 > this.ConnTime {
				this.Conn.Close()
				break
			}
		}
	}()
	for {
		msg, err := this.Conn.ReadMessage()
		if err != nil {
			logs.Logger.Error(err)
			break
		}
		msgstr := string(msg)
		if msgstr == "ping" {
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
