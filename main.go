package main

import (
	"ws-chat/app/db"
	// "time"
	// "wx-chat/app/oss"
	"ws-chat/app/ws"
)

func main() {
	// o, _ := oss.GetOss("minio")
	// str, _ := o.Upload("E:/360Downloads/support.png", "2222222222.png")
	// time.Sleep(time.Second * 30)
	// o.Remove(str)

	db.Create("erwe.db")

	ws.Start()
}
