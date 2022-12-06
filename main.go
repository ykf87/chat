package main

import (
	"ws-chat/app/ws"
)

func main() {
	// contnte := "data:image/jpeg;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAFCAYAAABirU3bAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAQSURBVBhXY/iPBigX+P8fADvQT7F5bXkYAAAAAElFTkSuQmCC"
	// s, c := funcs.ParseBase64(contnte)
	// o, _ := oss.GetOss("minio")
	// sss, eee := o.UploadBase64(s, c, "test/chat")
	// fmt.Println(sss, eee)

	// str, _ := o.Upload("E:/360Downloads/support.png", "2222222222.png")
	// time.Sleep(time.Second * 30)
	// o.Remove(str)

	// u := new(ws.User)
	// u.Id = 1
	// u.GetRooms()

	// s, c := funcs.ParseBase64(contnte)
	// fmt.Println("----", s, c)
	ws.Start()
}
