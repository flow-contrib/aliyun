package aliyun

import (
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

func init() {
	flow.RegisterHandler("devops.aliyun.dns.domain.record.add", AddDomainRecord)
	flow.RegisterHandler("devops.aliyun.dns.domain.record.update", UpdateDomainRecord)
	flow.RegisterHandler("devops.aliyun.dns.domain.record.delete", DeleteDomainRecord)
}

func AddDomainRecord(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.AddDomainRecord()
	if err != nil {
		return
	}

	return
}

func UpdateDomainRecord(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.UpdateDomainRecord()
	if err != nil {
		return
	}

	return
}

func DeleteDomainRecord(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.DeleteDomainRecord()
	if err != nil {
		return
	}

	return
}
