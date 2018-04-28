package aliyun

import (
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

func init() {
	flow.RegisterHandler("devops.aliyun.vpc.vpc.create", CreateVPC)
	flow.RegisterHandler("devops.aliyun.vpc.vpc.delete", DeleteVPC)
	flow.RegisterHandler("devops.aliyun.vpc.running.wait", WaitForAllVpcRunning)
	flow.RegisterHandler("devops.aliyun.vpc.vswitch.create", CreateVSwitch)
	flow.RegisterHandler("devops.aliyun.vpc.vswitch.delete", DeleteVSwitch)
}

func CreateVPC(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateVPCs()
	if err != nil {
		return
	}

	return
}

func DeleteVPC(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.DeleteVPC()
	if err != nil {
		return
	}

	return
}

func WaitForAllVpcRunning(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	err = aliyun.WaitForAllVpcRunning(30)

	if err != nil {
		return
	}

	return
}

func CreateVSwitch(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateVSwitch()
	if err != nil {
		return
	}

	return
}

func DeleteVSwitch(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.DeleteVSwitch()
	if err != nil {
		return
	}

	return
}
