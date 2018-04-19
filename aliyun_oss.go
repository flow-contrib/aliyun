package aliyun

import (
	"github.com/denverdino/aliyungo/oss"
)

type OSSBucketCreateArgs struct {
	Name string
	Perm oss.ACL
}

func (p *Aliyun) CreateOSSBucketArgs() (createArgs []*OSSBucketCreateArgs, err error) {

	resp, err := p.OSSClient().GetService()

	if err != nil {
		return
	}

	mapBuckets := map[string]*oss.BucketInfo{}
	for i, bucketInfo := range resp.Buckets {
		mapBuckets[bucketInfo.Name] = &resp.Buckets[i]
	}

	ossConf := p.Config.GetConfig("aliyun.oss")

	var args []*OSSBucketCreateArgs
	for _, key := range ossConf.Keys() {

		bucketName := ossConf.GetString(key+".name", key)

		_, exist := mapBuckets[bucketName]

		if exist {
			continue
		}

		arg := &OSSBucketCreateArgs{
			Name: bucketName,
			Perm: oss.ACL(ossConf.GetString(bucketName+".perm", "private")),
		}

		args = append(args, arg)
	}

	createArgs = args

	return
}

func (p *Aliyun) OSSBuckets() []*oss.Bucket {

	ossConf := p.Config.GetConfig("aliyun.oss")

	var buckets []*oss.Bucket
	for _, key := range ossConf.Keys() {

		bucketName := ossConf.GetString(key+".name", key)

		bucket := p.OSSClient().Bucket(bucketName)

		buckets = append(buckets, bucket)
	}

	return buckets
}
