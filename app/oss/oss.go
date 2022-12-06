package oss

import (
	"errors"
	"ws-chat/app/oss/minio"
)

type Oss interface {
	Upload(string, string) (string, error)
	UploadBase64(string, string) (string, error)
	Url(string) string
	Remove(string) error
	BluckName() string
}

func GetOss(s string) (Oss, error) {
	switch s {
	case "minio":
		return minio.GetClient()
	default:
		return nil, errors.New("找不到oss方法:" + s)
	}
}
