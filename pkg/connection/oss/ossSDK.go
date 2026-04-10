package oss

import (
	"Rshell/pkg/encrypt"
	"Rshell/pkg/logger"
	"io/ioutil"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Client struct {
	Cli             *oss.Client
	Bucket          *oss.Bucket
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string
}

var Service *Client

// var c *oss.Bucket

func InitClient(endPoint, accessKeyId, accessKeySecret, bucketName string) error {
	var ossClient *oss.Client
	var err error

	ossClient, err = oss.New(endPoint, accessKeyId, accessKeySecret)
	if err != nil {
		return err
	}

	var ossBucket *oss.Bucket
	ossBucket, err = ossClient.Bucket(bucketName)
	if err != nil {
		return err
	}

	Service = &Client{
		Cli:             ossClient,
		Bucket:          ossBucket,
		Endpoint:        endPoint,
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		BucketName:      bucketName,
	}
	return nil
}
func List(c *Client) ([]oss.ObjectProperties, error) {
	lsRes, err := c.Bucket.ListObjects(oss.MaxKeys(3), oss.Prefix(""))
	if err != nil {
		logger.Error("Error:", err)
		return nil, err
	}
	return lsRes.Objects, nil
}
func Send(c *Client, name string, content []byte) {
	encodeData, err := encrypt.EncodeBase64(content)
	// 1.通过字符串上传对象
	f := strings.NewReader(string(encodeData))
	// var err error
	err = c.Bucket.PutObject(name, f)
	if err != nil {
		logger.Error("[-]", "上传失败")
		return
	}

}
func Get(c *Client, name string) []byte {
	body, err := c.Bucket.GetObject(name)
	if err != nil {
		return nil
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄漏，导致请求无连接可用，程序无法正常工作。
	defer body.Close()
	data, err := ioutil.ReadAll(body)
	if err != nil {
		logger.Error("Error:", err)
		return nil
	}
	decodeData, err := encrypt.DecodeBase64(data)
	return decodeData
}

func Del(c *Client, name string) error {
	err := c.Bucket.DeleteObject(name)
	if err != nil {
		logger.Error("Error deleting object:", name, err)
		return err
	}
	return nil
}
