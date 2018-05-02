package aliyun

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/sirupsen/logrus"
)

type Rule struct {
	RuleName       string
	Domain         string
	Url            string `json:",omitempty"`
	VServerGroupId string
}

func (p *Aliyun) CreateSLBHTTPListenerRule() (err error) {

	balancersConfig := p.Config.GetConfig("aliyun.slb.balancer")

	if balancersConfig.IsEmpty() {
		return
	}

	balancers, err := p.ListLoadBalancers(true)
	if err != nil {
		return
	}

	if len(balancers) == 0 {
		return
	}

	slbNames := balancersConfig.Keys()

	if len(slbNames) == 0 {
		return
	}

	var reqs []*slb.CreateRulesRequest

	for _, slbName := range slbNames {

		slbInstance, exist := balancers[slbName]

		if !exist {
			err = fmt.Errorf("slb instance not exist: %s", slbName)
			return
		}

		var alreadyListendPorts = make(map[string]bool)

		for _, port := range slbInstance.ListenerPorts.ListenerPort {
			alreadyListendPorts[port] = true
		}

		lbConfig := balancersConfig.GetConfig(slbName)
		listenersConfig := lbConfig.GetConfig("listener.http")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, listenerName := range listenersConfig.Keys() {
			listenerConf := listenersConfig.GetConfig(listenerName)

			if listenerConf.IsEmpty() {
				continue
			}

			port := int(listenerConf.GetInt32("listen-port"))

			if !alreadyListendPorts[strconv.Itoa(port)] {
				err = fmt.Errorf("port %d not listened in balance %s", port, slbName)
				return
			}

			describeVgroupReq := slb.CreateDescribeVServerGroupsRequest()

			describeVgroupReq.LoadBalancerId = slbInstance.LoadBalancerId
			describeVgroupReq.RegionId = slbInstance.RegionId

			var vSrvGroupsResp *slb.DescribeVServerGroupsResponse
			vSrvGroupsResp, err = p.SLBClient().DescribeVServerGroups(describeVgroupReq)
			if err != nil {
				return
			}

			if vSrvGroupsResp == nil || len(vSrvGroupsResp.VServerGroups.VServerGroup) == 0 {
				err = fmt.Errorf("no vserver group exist, lb: %s, listener: %s", slbName, listenerName)
				return
			}

			var mapSrvGroups = make(map[string]string) //name:id

			for _, vg := range vSrvGroupsResp.VServerGroups.VServerGroup {
				mapSrvGroups[vg.VServerGroupName] = vg.VServerGroupId
			}

			describeRulReq := slb.CreateDescribeRulesRequest()

			describeRulReq.RegionId = p.Region
			describeRulReq.LoadBalancerId = slbInstance.LoadBalancerId
			describeRulReq.ListenerPort = requests.NewInteger(port)

			var ruleDescribRep *slb.DescribeRulesResponse
			ruleDescribRep, err = p.SLBClient().DescribeRules(describeRulReq)

			mapExistsRules := map[string]slb.Rule{}

			for i, rule := range ruleDescribRep.Rules.Rule {
				mapExistsRules[rule.RuleName] = ruleDescribRep.Rules.Rule[i]
			}

			rulesConf := listenerConf.GetConfig("rules")

			var rules []Rule

			for _, ruleName := range rulesConf.Keys() {

				if _, ruleExist := mapExistsRules[ruleName]; ruleExist {
					logrus.WithField("CODE", p.Code).
						WithField("SLB-NAME", slbName).
						WithField("SLB-ID", slbInstance.LoadBalancerId).
						WithField("SLB-LISTENER", listenerName).
						WithField("LSB-LISTENER-RULE", ruleName).Infoln("Listener rule already created")

					continue
				}

				ruleConf := rulesConf.GetConfig(ruleName)

				vGroupName := ruleConf.GetString("vserver-group-name")

				vGroupId, exist := mapSrvGroups[vGroupName]
				if !exist {
					err = fmt.Errorf("vgroup of %s in lb %s not created.", vGroupName, slbName)
					return
				}

				domain := ruleConf.GetString("domain")
				url := ruleConf.GetString("url")

				r := Rule{
					RuleName:       ruleName,
					Domain:         domain,
					Url:            url,
					VServerGroupId: vGroupId,
				}

				rules = append(rules, r)
			}

			var ruleData []byte
			ruleData, err = json.Marshal(rules)
			if err != nil {
				err = fmt.Errorf("marshal rule list error, slb instance: %s, listener: %s", slbName, listenerName)
				return
			}

			req := slb.CreateCreateRulesRequest()

			req.RegionId = p.Region
			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.ListenerPort = requests.NewInteger(port)
			req.RuleList = string(ruleData)

			reqs = append(reqs, req)
		}
	}

	for _, req := range reqs {
		_, err = p.SLBClient().CreateRules(req)
		if err != nil {

			if IsAliErrCode(err, "DomainExist") {
				err = nil
				continue
			}

			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", req.LoadBalancerId).
			WithField("SLB-BANLANCER-LISTENER-PORT", req.ListenerPort).
			Infoln("SLB listener rules created")
	}

	return
}
