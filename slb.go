package aliyun

import (
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

func init() {
	flow.RegisterHandler("devops.aliyun.slb.balancer.create", CreateSLBBalancer)
	flow.RegisterHandler("devops.aliyun.slb.balancer.delete", DeleteSLBBalancer)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.http.create", CreateSLBHTTPBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.https.create", CreateSLBHTTPSBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.tcp.create", CreateSLBTCPBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.udp.create", CreateSLBUDPBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.vserver-group.create", CreateVServerGroup)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.rules.create", CreateSLBHTTPListenerRule)
}

func CreateSLBBalancer(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateLoadBalancer()
	if err != nil {
		return
	}

	return
}

func DeleteSLBBalancer(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.DeleteLoadBalancer()
	if err != nil {
		return
	}

	return
}

func CreateSLBHTTPBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateLoadBalancerHTTPListener()
	if err != nil {
		return
	}

	return
}

func CreateSLBHTTPSBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateLoadBalancerHTTPSListener()
	if err != nil {
		return
	}

	return
}

func CreateSLBTCPBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateLoadBalancerTCPListener()
	if err != nil {
		return
	}

	return
}

func CreateSLBUDPBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateLoadBalancerUDPListener()
	if err != nil {
		return
	}

	return
}

func CreateVServerGroup(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateVServerGroup()
	if err != nil {
		return
	}

	return
}

func CreateSLBHTTPListenerRule(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateSLBHTTPListenerRule()
	if err != nil {
		return
	}

	return
}
