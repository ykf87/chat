package minio

import (
	"context"
	"os"
	"strings"
	"wx-chat/app/configs"
	"wx-chat/app/funcs"
	"wx-chat/app/logs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	Conf  *configs.Oss
	Minio *minio.Client
}

var MailObj *Client

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

	return info.Bucket + "/" + info.Key, nil
}

//上传base64格式文件到服务器
//content base64的内容
//filename 保存的路径或文件名
func (this *Client) UploadBase64(content, filename string) (string, error) {
	// this.Minio.put
	return "", nil
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
	return strings.TrimRight(url, "/") + "/" + filename
}
