package ws

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type Connections struct {
	Conn      *websocket.Conn
	InChan    chan []byte
	OutChan   chan []byte
	CloseChan chan byte
	Mutex     sync.Mutex
	IsClose   bool
}

// 初始化链接
func InitConnect(WsConn *websocket.Conn) (conn *Connections, err error) {
	conn = &Connections{
		Conn:      WsConn,
		InChan:    make(chan []byte, 1024),
		OutChan:   make(chan []byte, 1024),
		CloseChan: make(chan byte, 1),
	}
	go conn.reader()
	go conn.writer()
	return
}

// 线程安全的消息读取
func (conn *Connections) ReadMessage() (data []byte, err error) {
	select {
	case <-conn.CloseChan:
		err = errors.New("链接被关闭")
	case data = <-conn.InChan:
	}
	return
}

// 线程安全的消息写入
func (conn *Connections) WriteMessage(data []byte) (err error) {
	select {
	case conn.OutChan <- data:
	case <-conn.CloseChan:
		err = errors.New("链接被关闭!")
	}

	return
}

// 线程安全的关闭
func (conn *Connections) Close() {
	conn.Conn.Close()

	conn.Mutex.Lock()
	if !conn.IsClose {
		close(conn.CloseChan)
		conn.IsClose = true
	}
	conn.Mutex.Unlock()

}

// 消息读取
func (conn *Connections) reader() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.Conn.ReadMessage(); err != nil {
			goto ERR
		}
		select {
		case <-conn.CloseChan:
			goto ERR
		case conn.InChan <- data:
		}

	}
ERR:
	conn.Close()
}

// 发送消息
func (conn *Connections) writer() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case <-conn.CloseChan:
			goto ERR
		case data = <-conn.OutChan:
		}

		if conn.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
