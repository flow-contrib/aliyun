package aliyun

import (
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

func init() {
	flow.RegisterHandler("devops.aliyun.oss.bucket.create", CreateOSSBucket)
	flow.RegisterHandler("devops.aliyun.oss.bucket.delete", DeleteOSSBucket)
}

func CreateOSSBucket(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateOSSBucket()

	if err != nil {
		return
	}

	return
}

func DeleteOSSBucket(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.DeleteOSSBucket()

	if err != nil {
		return
	}

	return
}
