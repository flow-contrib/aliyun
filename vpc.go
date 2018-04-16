package aliyun

import (
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
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

	args, err := aliyun.CreateVPCArgs()
	if err != nil {
		return
	}

	for _, arg := range args {

		resp, e := aliyun.ECSClient().CreateVpc(arg)
		if e != nil {
			return e
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("ECS-VPC-NAME", arg.VpcName).
			WithField("ECS-VPC-ID", resp.VpcId).
			WithField("ECS-VPC-REGION", arg.RegionId).
			Infoln("VPC created")
	}

	return
}

func DeleteVPC(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.DeleteVPCArgs()
	if err != nil {
		return
	}

	for _, arg := range args {

		err = aliyun.ECSClient().DeleteVpc(arg.VpcId)
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("ECS-VPC-ID", arg.VpcId).
			Infoln("VPC deleted")
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

	args, err := aliyun.CreateVSwitch()
	if err != nil {
		return
	}

	for _, arg := range args {

		switchId, e := aliyun.ECSClient().CreateVSwitch(arg)
		if e != nil {
			return e
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("ECS-VSWITCH-NAME", arg.VSwitchName).
			WithField("ECS-VSWITCH-ID", switchId).
			Infoln("VSwitch created")
	}

	return
}

func DeleteVSwitch(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.DeleteVSwitchArgs()
	if err != nil {
		return
	}

	for _, arg := range args {

		err = aliyun.ECSClient().DeleteVSwitch(arg.VSwitchId)
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("ECS-VSWITCH-ID", arg.VSwitchId).
			Infoln("VSwitch deleted")
	}

	return
}
