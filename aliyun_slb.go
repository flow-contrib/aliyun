package aliyun

import (
	"fmt"

	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"github.com/gogap/logrus"
)

func (p *Aliyun) ListLoadBalancers(details bool) (val map[string]*slb.LoadBalancerType, err error) {

	balancersConfig := p.Config.GetConfig("aliyun.slb.balancer")

	if balancersConfig.IsEmpty() {
		return
	}

	slbNames := balancersConfig.Keys()

	if len(slbNames) == 0 {
		return
	}

	mapSLBNames := map[string]bool{}

	for _, name := range slbNames {
		mapSLBNames[name] = true
	}

	lbs, err := p.SLBClient().DescribeLoadBalancers(
		&slb.DescribeLoadBalancersArgs{
			RegionId: p.Region,
		},
	)

	if err != nil {
		return
	}

	if len(lbs) == 0 {
		return
	}

	ret := map[string]*slb.LoadBalancerType{}

	for i, lb := range lbs {
		if mapSLBNames[lb.LoadBalancerName] {

			if details {
				var lbDetails *slb.LoadBalancerType
				lbDetails, err = p.SLBClient().DescribeLoadBalancerAttribute(lb.LoadBalancerId)
				if err != nil {
					return
				}
				ret[lb.LoadBalancerName] = lbDetails
			} else {
				ret[lb.LoadBalancerName] = &lbs[i]
			}
		}
	}

	val = ret

	return
}

func (p *Aliyun) CreateLoadBalancerArgs() (createArgs []*slb.CreateLoadBalancerArgs, err error) {

	currentLBSs, err := p.ListLoadBalancers(false)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	balancersConfig := p.Config.GetConfig("aliyun.slb.balancer")

	slbNames := balancersConfig.Keys()

	if len(slbNames) == 0 {
		return
	}

	var args []*slb.CreateLoadBalancerArgs

	for _, needCreateSLBName := range slbNames {

		// already exist
		if _, exist := currentLBSs[needCreateSLBName]; exist {

			logrus.WithField("CODE", p.Code).
				WithField("SLB-NAME", needCreateSLBName).
				WithField("REGION", p.Region).
				Infoln("SLB already created")

			continue
		}

		slbConf := balancersConfig.GetConfig(needCreateSLBName)

		addressType := slbConf.GetString("address-type", "internet")

		vSwitchId := ""

		vpcName := slbConf.GetString("vpc-name")
		vSwitchName := slbConf.GetString("vswitch-name")

		if len(vpcName)+len(vSwitchName) > 0 {
			if len(vpcName) == 0 || len(vSwitchName) == 0 {
				err = fmt.Errorf("slb config of %s's vpc-name or vswitch-name is empty", needCreateSLBName)
				return
			} else {
				var vSwitch *ecs.VSwitchSetType
				vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)

				if err != nil {
					return
				}

				if vSwitch == nil {
					err = fmt.Errorf("slb instance of %s vsiwtch is not found,VPC:%s VSwtich: %s", needCreateSLBName, vpcName, vSwitchName)
					return
				}
				vSwitchId = vSwitch.VSwitchId
			}
		}

		chargeType := slbConf.GetString("charge-type", "paybytraffic")
		bandWidth := slbConf.GetInt64("band-width", 100)

		arg := &slb.CreateLoadBalancerArgs{
			RegionId:           p.Region,
			LoadBalancerName:   needCreateSLBName,
			AddressType:        slb.AddressType(addressType),
			VSwitchId:          vSwitchId,
			InternetChargeType: slb.InternetChargeType(chargeType),
			Bandwidth:          int(bandWidth),
		}

		args = append(args, arg)
	}

	createArgs = args

	return
}

func (p *Aliyun) DeleteLoadBalancerArgs() (deleteArgs []*slb.DeleteLoadBalancerArgs, err error) {
	currentLBSs, err := p.ListLoadBalancers(false)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	slbConfig := p.Config.GetConfig("aliyun.slb.balancer")

	slbNames := slbConfig.Keys()

	if len(slbNames) == 0 {
		return
	}

	var args []*slb.DeleteLoadBalancerArgs

	for _, needDeleteSLBName := range slbNames {

		if len(needDeleteSLBName) == 0 {
			continue
		}

		lbInstancd, exist := currentLBSs[needDeleteSLBName]

		if !exist {
			continue
		}

		arg := &slb.DeleteLoadBalancerArgs{
			LoadBalancerId: lbInstancd.LoadBalancerId,
		}

		args = append(args, arg)
	}

	deleteArgs = args

	return
}
