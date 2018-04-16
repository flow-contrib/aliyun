package aliyun

import (
	"fmt"
	"strings"

	"github.com/denverdino/aliyungo/slb"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("devops.aliyun.slb.balancer.create", CreateSLBBalancer)
	flow.RegisterHandler("devops.aliyun.slb.balancer.delete", DeleteSLBBalancer)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.http.create", CreateSLBHTTPBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.tcp.create", CreateSLBTCPBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.udp.create", CreateSLBUDPBanlancerListener)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.vserver-group.create", CreateVServerGroup)
	flow.RegisterHandler("devops.aliyun.slb.balancer.listener.rules.create", CreateSLBHTTPListenerRule)
}

func CreateSLBBalancer(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateLoadBalancerArgs()
	if err != nil {
		return
	}

	for _, arg := range args {

		var resp *slb.CreateLoadBalancerResponse
		resp, err = aliyun.SLBClient().CreateLoadBalancer(arg)
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-NAME", resp.LoadBalancerName).
			WithField("SLB-BANLANCER-ID", resp.LoadBalancerId).
			WithField("SLB-REGION", arg.RegionId).
			Infoln("SLB banlancer created")
	}

	return
}

func DeleteSLBBalancer(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.DeleteLoadBalancerArgs()
	if err != nil {
		return
	}

	for _, arg := range args {

		err = aliyun.SLBClient().DeleteLoadBalancer(arg.LoadBalancerId)
		if err != nil {
			if strings.Contains(err.Error(), "InvalidLoadBalancerId.NotFound") {
				err = nil
				continue
			}

			err = fmt.Errorf("delete balancer failure, balancer id : %s, error: %s", arg.LoadBalancerId, err.Error())
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-ID", arg.LoadBalancerId).
			Infoln("SLB banlancer deleted")
	}

	return
}

func CreateSLBHTTPBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateLoadBalancerHTTPListenerArgs()
	if err != nil {
		return
	}

	for _, arg := range args {

		err = arg.CreateAndWait()
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-ID", arg.LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", arg.ListenerPort).
			Infoln("SLB http listener created")
	}

	return
}

func CreateSLBTCPBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateLoadBalancerSocketListenerArgs(true)
	if err != nil {
		return
	}

	for _, arg := range args {

		err = arg.CreateAndWait()
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-ID", arg.LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", arg.ListenerPort).
			Infoln("SLB TCP listener created")
	}

	return
}

func CreateSLBUDPBanlancerListener(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateLoadBalancerSocketListenerArgs(false)
	if err != nil {
		return
	}

	for _, arg := range args {

		err = arg.CreateAndWait()
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-ID", arg.LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", arg.ListenerPort).
			Infoln("SLB UDP listener created")
	}

	return
}

func CreateVServerGroup(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateVServerGroupArgs()
	if err != nil {
		return
	}

	for _, arg := range args {
		var resp *slb.CreateVServerGroupResponse
		resp, err = aliyun.SLBClient().CreateVServerGroup(arg)
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-ID", arg.LoadBalancerId).
			WithField("SLB-BANLANCER-VGROUP-NAME", arg.VServerGroupName).
			WithField("SLB-BANLANCER-VGROUP-ID", resp.VServerGroupId).
			Infoln("SLB VGroup created")
	}

	return
}

func CreateSLBHTTPListenerRule(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateSLBHTTPListenerRuleArgs()
	if err != nil {
		return
	}

	for _, arg := range args {
		err = aliyun.SLBClient().CreateRules(arg)
		if err != nil {

			if IsAliErrCode(err, "DomainExist") {
				err = nil
				continue
			}

			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("SLB-BANLANCER-ID", arg.LoadBalancerId).
			WithField("SLB-BANLANCER-LISTENER-PORT", arg.ListenerPort).
			Infoln("SLB listener rules created")
	}

	return
}
