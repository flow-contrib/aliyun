package aliyun

import (
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("devops.aliyun.oss.bucket.create", CreateOSSBucket)
	flow.RegisterHandler("devops.aliyun.oss.bucket.delete", DeleteOSSBucket)
}

func CreateOSSBucket(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateOSSBucketArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		bucket := aliyun.OSSClient().Bucket(arg.Name)
		err = bucket.PutBucket(arg.Perm)
		if err != nil {
			return
		}
	}

	return
}

func DeleteOSSBucket(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	buckets := aliyun.OSSBuckets()

	for _, bucket := range buckets {
		e := bucket.DelBucket()

		if e != nil {
			if strings.Contains(e.Error(), "NoSuchBucket") {
				continue
			}

			err = e
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("OSS-BUCKET", bucket.Name).
			WithField("OSS-BUCKET-REGION", bucket.Region).
			Infoln("Bucket deleted")
	}

	return
}
