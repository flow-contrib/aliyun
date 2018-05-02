package aliyun

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

	"github.com/sirupsen/logrus"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"github.com/denverdino/aliyungo/common"
)

type SLBLoadBalancer struct {
	*slb.LoadBalancer

	AutoReleaseTime          int
	Bandwidth                int
	LoadBalancerSpec         string
	EndTime                  string
	EndTimeStamp             int
	ListenerPorts            slb.ListenerPorts
	ListenerPortsAndProtocol slb.ListenerPortsAndProtocol
}

func (p *Aliyun) ListLoadBalancers(details bool) (val map[string]*SLBLoadBalancer, err error) {

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

	req := slb.CreateDescribeLoadBalancersRequest()

	req.RegionId = p.Region

	resp, err := p.SLBClient().DescribeLoadBalancers(req)

	if err != nil {
		return
	}

	if len(resp.LoadBalancers.LoadBalancer) == 0 {
		return
	}

	ret := map[string]*SLBLoadBalancer{}

	for i, lb := range resp.LoadBalancers.LoadBalancer {
		if mapSLBNames[lb.LoadBalancerName] {

			if details {
				var lbDetails *slb.DescribeLoadBalancerAttributeResponse
				attrReq := slb.CreateDescribeLoadBalancerAttributeRequest()
				attrReq.LoadBalancerId = lb.LoadBalancerId

				lbDetails, err = p.SLBClient().DescribeLoadBalancerAttribute(attrReq)

				if err != nil {
					return
				}

				lbWithDetails := &SLBLoadBalancer{
					LoadBalancer:             &resp.LoadBalancers.LoadBalancer[i],
					AutoReleaseTime:          lbDetails.AutoReleaseTime,
					Bandwidth:                lbDetails.Bandwidth,
					LoadBalancerSpec:         lbDetails.LoadBalancerSpec,
					EndTime:                  lbDetails.EndTime,
					EndTimeStamp:             lbDetails.EndTimeStamp,
					ListenerPorts:            lbDetails.ListenerPorts,
					ListenerPortsAndProtocol: lbDetails.ListenerPortsAndProtocol,
				}

				lbWithDetails.BackendServers.BackendServer = lbDetails.BackendServers.BackendServer

				ret[lb.LoadBalancerName] = lbWithDetails
			} else {
				ret[lb.LoadBalancerName] = &SLBLoadBalancer{LoadBalancer: &resp.LoadBalancers.LoadBalancer[i]}
			}
		}
	}

	val = ret

	return
}

func (p *Aliyun) CreateLoadBalancer() (err error) {

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

	var reqs []*slb.CreateLoadBalancerRequest

	for _, needCreateSLBName := range slbNames {

		// already exist
		if _, exist := currentLBSs[needCreateSLBName]; exist {

			logrus.WithField("CODE", p.Code).
				WithField("SLB-NAME", needCreateSLBName).
				WithField("REGION", common.Region(p.Region)).
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
				var vSwitch *vpc.VSwitch
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

		req := slb.CreateCreateLoadBalancerRequest()

		req.RegionId = p.Region
		req.LoadBalancerName = needCreateSLBName
		req.AddressType = addressType
		req.VSwitchId = vSwitchId
		req.InternetChargeType = chargeType
		req.Bandwidth = requests.NewInteger(int(bandWidth))

		reqs = append(reqs, req)
	}

	for i := 0; i < len(reqs); i++ {
		var resp *slb.CreateLoadBalancerResponse
		resp, err = p.SLBClient().CreateLoadBalancer(reqs[i])
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-NAME", resp.LoadBalancerName).
			WithField("SLB-BANLANCER-ID", resp.LoadBalancerId).
			WithField("SLB-REGION", reqs[i].RegionId).
			Infoln("SLB banlancer created")
	}

	return
}

func (p *Aliyun) DeleteLoadBalancer() (err error) {
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

	var reqs []*slb.DeleteLoadBalancerRequest

	for _, needDeleteSLBName := range slbNames {

		if len(needDeleteSLBName) == 0 {
			continue
		}

		lbInstancd, exist := currentLBSs[needDeleteSLBName]

		if !exist {
			continue
		}

		req := slb.CreateDeleteLoadBalancerRequest()

		req.LoadBalancerId = lbInstancd.LoadBalancerId

		reqs = append(reqs, req)
	}

	for i := 0; i < len(reqs); i++ {
		_, err = p.SLBClient().DeleteLoadBalancer(reqs[i])

		if IsAliErrCode(err, "InvalidLoadBalancerId.NotFound") {
			err = nil
			continue
		}

		if err != nil {
			err = fmt.Errorf("delete balancer failure, balancer id : %s, error: %s", reqs[i].LoadBalancerId, err.Error())
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", reqs[i].LoadBalancerId).
			Infoln("SLB banlancer deleted")
	}

	return
}
