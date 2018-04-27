package aliyun

import (
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

func init() {
	flow.RegisterHandler("devops.aliyun.ecs.vpc.create", CreateVPC)
	flow.RegisterHandler("devops.aliyun.ecs.vpc.delete", DeleteVPC)
	flow.RegisterHandler("devops.aliyun.ecs.vpc.running.wait", WaitForAllVpcRunning)
	flow.RegisterHandler("devops.aliyun.ecs.vswitch.create", CreateVSwitch)
	flow.RegisterHandler("devops.aliyun.ecs.vswitch.delete", DeleteVSwitch)
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

	err = aliyun.DeleteVPCArgs()
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

	err = aliyun.DeleteVSwitchArgs()
	if err != nil {
		return
	}

	return
}
