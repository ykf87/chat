package minio

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"ws-chat/app/configs"
	"ws-chat/app/funcs"
	"ws-chat/app/logs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	Conf  *configs.Oss
	Minio *minio.Client
}

var MailObj *Client

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetClient() (*Client, error) {
	if MailObj != nil {
		return MailObj, nil
	}
	conf := configs.Conf.Oss
	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.Accesskeyid, conf.Secret, ""),
		Secure: conf.Ssl,
	})
	if err != nil {
		return nil, err
	}
	cc := new(Client)
	cc.Conf = conf
	cc.Minio = minioClient
	MailObj = cc
	return cc, nil
}

//获取桶名
func (this *Client) BluckName() string {
	return this.Conf.Bucket
}

//objectName 云端保存的路径和文件名
//ex: o.Upload("E:/360Downloads/support.png", "2222222222.png")
func (this *Client) Upload(filePath, objectName string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	contentType, err := funcs.GetFileContentType(f)
	if err != nil {
		return "", err
	}

	// Upload the zip file with FPutObject
	ctx := context.Background()
	// info, err := this.Minio.PutObject(ctx, this.Conf.Bucket, objectName, f, fileInfo.Size(), minio.PutObjectOptions{ContentType: contentType})
	info, err := this.Minio.FPutObject(ctx, this.Conf.Bucket, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logs.Logger.Error("上传云端失败: ", err.Error(), this.Conf.Bucket)
		return "", err
	}

	return info.Key, nil
}

//上传base64格式文件到服务器
//content base64的内容
//filename 保存的路径或文件名
func (this *Client) UploadBase64(content, path string) (string, error) {
	var contentType string
	content, contentType = funcs.ParseBase64(content)
	if content == "" {
		return "", errors.New("File error")
	}

	if strings.Contains(path, ".") == false {
		tp := funcs.GetBase64Type(contentType)
		if tp == "" {
			logs.Logger.Error("不支持的格式 ", contentType)
			return "", errors.New(contentType + " 不支持")
		}
		path = strings.TrimRight(path, "/") + fmt.Sprintf("/%d-%d%s", time.Now().Unix(), funcs.Random(100, 999), tp)
	}

	contbyte, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	info, err := this.Minio.PutObject(context.Background(), this.Conf.Bucket, path, bytes.NewBuffer(contbyte), int64(len(contbyte)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logs.Logger.Error("上传云端失败: ", err.Error(), this.Conf.Bucket)
		return "", err
	}

	return info.Key, nil
}

func (this *Client) Remove(src string) error {
	opts := minio.RemoveObjectOptions{
		ForceDelete:      true,
		GovernanceBypass: true,
	}
	src = strings.Replace(src, this.Conf.Bucket+"/", "", -1)
	return this.Minio.RemoveObject(context.Background(), this.Conf.Bucket, src, opts)
}

func (this *Client) Url(filename string) string {
	url := this.Conf.Url
	if url == "" {
		sheme := "http://"
		if this.Conf.Ssl == true {
			sheme = "https://"
		}
		url = sheme + this.Conf.Endpoint
	}
	return strings.TrimRight(url, "/") + "/" + this.Conf.Bucket + "/" + filename
}
