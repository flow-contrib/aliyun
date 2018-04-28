package aliyun

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"
	"strings"
)

type OSSBucketCreateArgs struct {
	Name string
	Perm string
}

func (p *Aliyun) CreateOSSBucket() (err error) {

	resp, err := p.OSSClient().ListBuckets()

	if err != nil {
		return
	}

	mapBuckets := map[string]*oss.BucketProperties{}
	for i, bucketInfo := range resp.Buckets {
		mapBuckets[bucketInfo.Name] = &resp.Buckets[i]
	}

	ossConf := p.Config.GetConfig("aliyun.oss.bucket")

	var args []*OSSBucketCreateArgs
	for _, key := range ossConf.Keys() {

		bucketName := ossConf.GetString(key+".name", key)

		_, exist := mapBuckets[bucketName]

		if exist {
			continue
		}

		arg := &OSSBucketCreateArgs{
			Name: bucketName,
			Perm: ossConf.GetString(bucketName+".perm", "private"),
		}

		args = append(args, arg)
	}

	for _, arg := range args {
		err = p.OSSClient().CreateBucket(arg.Name)
		if err != nil {
			return
		}

		logrus.WithField("code", p.Code).WithField("bucket", arg.Name).Infoln("bucket created")
	}

	return
}

func (p *Aliyun) DeleteOSSBucket() (err error) {

	ossConf := p.Config.GetConfig("aliyun.oss.bucket")

	for _, key := range ossConf.Keys() {

		bucketName := ossConf.GetString(key+".name", key)

		err = p.OSSClient().DeleteBucket(bucketName)

		if err != nil {

			if strings.Contains(err.Error(), "ErrorCode=NoSuchBucket") {
				logrus.WithField("code", p.Code).WithField("bucket", bucketName).Debugln("bucket not exist, ignore to delete")
				err = nil
				continue
			}

			err = fmt.Errorf("delete bucket '%s' failure: %s", bucketName, err.Error())
			return
		}

		logrus.WithField("code", p.Code).WithField("bucket", bucketName).Infoln("bucket deleted")
	}

	return
}
