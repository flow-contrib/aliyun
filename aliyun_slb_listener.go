package aliyun

import (
	"encoding/json"
	"fmt"

	"github.com/denverdino/aliyungo/slb"
	"github.com/gogap/logrus"
)

type CreateLoadBalancerHTTPListenerArgs struct {
	slbClient *slb.Client
	slb.HTTPListenerType
	serverCertificateId string
}

func (p *CreateLoadBalancerHTTPListenerArgs) IsHTTPs() bool {
	return len(p.serverCertificateId) != 0
}

func (p *CreateLoadBalancerHTTPListenerArgs) HTTPCreationArg() (args *slb.CreateLoadBalancerHTTPListenerArgs) {
	ret := slb.CreateLoadBalancerHTTPListenerArgs(p.HTTPListenerType)
	return &ret
}

func (p *CreateLoadBalancerHTTPListenerArgs) HTTPSCreationArg() (args *slb.CreateLoadBalancerHTTPSListenerArgs) {
	return &slb.CreateLoadBalancerHTTPSListenerArgs{
		HTTPListenerType:    p.HTTPListenerType,
		ServerCertificateId: p.serverCertificateId,
	}
}

func (p *CreateLoadBalancerHTTPListenerArgs) Create() (err error) {
	if p.IsHTTPs() {
		err = p.slbClient.CreateLoadBalancerHTTPSListener(p.HTTPSCreationArg())
		if err != nil {
			return
		}
	} else {
		err = p.slbClient.CreateLoadBalancerHTTPListener(p.HTTPCreationArg())
		if err != nil {
			return
		}
	}

	return
}

func (p *CreateLoadBalancerHTTPListenerArgs) Wait(toStatus slb.ListenerStatus) (err error) {
	if p.IsHTTPs() {
		err = p.slbClient.WaitForListenerAsyn(p.LoadBalancerId, p.ListenerPort, slb.ListenerType("HTTPS"), toStatus, 10)
	} else {
		err = p.slbClient.WaitForListenerAsyn(p.LoadBalancerId, p.ListenerPort, slb.ListenerType("HTTP"), toStatus, 10)
	}
	return
}

func (p *CreateLoadBalancerHTTPListenerArgs) CreateAndWait() (err error) {
	err = p.Create()
	if err != nil {
		return
	}

	err = p.slbClient.StartLoadBalancerListener(p.LoadBalancerId, p.ListenerPort)
	if err != nil {
		return
	}

	err = p.Wait(slb.Running)
	if err != nil {
		return
	}

	return
}

type CreateLoadBalancerSocketListenerArgs struct {
	slbClient *slb.Client
	isTCP     bool

	// Common part
	LoadBalancerId            string
	ListenerPort              int
	BackendServerPort         int
	Bandwidth                 int
	Scheduler                 slb.SchedulerType
	PersistenceTimeout        int
	HealthCheck               slb.FlagType
	HealthCheckConnectPort    int
	HealthyThreshold          int
	UnhealthyThreshold        int
	HealthCheckConnectTimeout int
	HealthCheckInterval       int
	VServerGroup              slb.FlagType
	VServerGroupId            string

	// TCP Part
	HealthCheckType     slb.HealthCheckType
	HealthCheckDomain   string
	HealthCheckURI      string
	HealthCheckHttpCode slb.HealthCheckHttpCodeType
}

func (p *CreateLoadBalancerSocketListenerArgs) IsTCP() bool {
	return p.isTCP
}

func (p *CreateLoadBalancerSocketListenerArgs) Create() (err error) {
	if p.IsTCP() {
		err = p.slbClient.CreateLoadBalancerTCPListener(
			&slb.CreateLoadBalancerTCPListenerArgs{
				LoadBalancerId:            p.LoadBalancerId,
				ListenerPort:              p.ListenerPort,
				BackendServerPort:         p.BackendServerPort,
				Bandwidth:                 p.Bandwidth,
				Scheduler:                 p.Scheduler,
				PersistenceTimeout:        p.PersistenceTimeout,
				HealthCheck:               p.HealthCheck,
				HealthCheckConnectPort:    p.HealthCheckConnectPort,
				HealthyThreshold:          p.HealthyThreshold,
				UnhealthyThreshold:        p.UnhealthyThreshold,
				HealthCheckConnectTimeout: p.HealthCheckConnectTimeout,
				HealthCheckInterval:       p.HealthCheckInterval,
				VServerGroup:              p.VServerGroup,
				VServerGroupId:            p.VServerGroupId,

				HealthCheckType:     p.HealthCheckType,
				HealthCheckDomain:   p.HealthCheckDomain,
				HealthCheckURI:      p.HealthCheckURI,
				HealthCheckHttpCode: p.HealthCheckHttpCode,
			},
		)
		if err != nil {
			return
		}
	} else {
		err = p.slbClient.CreateLoadBalancerUDPListener(
			&slb.CreateLoadBalancerUDPListenerArgs{
				LoadBalancerId:            p.LoadBalancerId,
				ListenerPort:              p.ListenerPort,
				BackendServerPort:         p.BackendServerPort,
				Bandwidth:                 p.Bandwidth,
				Scheduler:                 p.Scheduler,
				PersistenceTimeout:        p.PersistenceTimeout,
				HealthCheck:               p.HealthCheck,
				HealthCheckConnectPort:    p.HealthCheckConnectPort,
				HealthyThreshold:          p.HealthyThreshold,
				UnhealthyThreshold:        p.UnhealthyThreshold,
				HealthCheckConnectTimeout: p.HealthCheckConnectTimeout,
				HealthCheckInterval:       p.HealthCheckInterval,
				VServerGroup:              p.VServerGroup,
				VServerGroupId:            p.VServerGroupId,
			},
		)
		if err != nil {
			return
		}
	}

	return
}

func (p *CreateLoadBalancerSocketListenerArgs) Wait(toStatus slb.ListenerStatus) (err error) {
	if p.IsTCP() {
		err = p.slbClient.WaitForListenerAsyn(p.LoadBalancerId, p.ListenerPort, slb.ListenerType("TCP"), toStatus, 10)
	} else {
		err = p.slbClient.WaitForListenerAsyn(p.LoadBalancerId, p.ListenerPort, slb.ListenerType("UDP"), toStatus, 10)
	}
	return
}

func (p *CreateLoadBalancerSocketListenerArgs) CreateAndWait() (err error) {
	err = p.Create()
	if err != nil {
		return
	}

	err = p.slbClient.StartLoadBalancerListener(p.LoadBalancerId, p.ListenerPort)
	if err != nil {
		return
	}

	err = p.Wait(slb.Running)
	if err != nil {
		return
	}

	return
}

func (p *Aliyun) CreateLoadBalancerHTTPListenerArgs() (createArgs []*CreateLoadBalancerHTTPListenerArgs, err error) {
	currentLBSs, err := p.ListLoadBalancers(true)
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

	var args []*CreateLoadBalancerHTTPListenerArgs

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			err = fmt.Errorf("slb of %s not exist", slbName)
			return
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("http-listener")

		if listenersConfig.IsEmpty() {
			continue
		}

		var alreadyListendPorts = make(map[int]bool)

		for _, port := range slbInstance.ListenerPorts.ListenerPort {
			alreadyListendPorts[port] = true
		}

		for _, listenerName := range listenersConfig.Keys() {

			listenerConfig := listenersConfig.GetConfig(listenerName)

			if listenerConfig.IsEmpty() {
				err = fmt.Errorf("listen config is empty: %s.%s", slbName, listenerName)
				return
			}

			listenPort := listenerConfig.GetInt32("listen-port")

			if listenPort <= 0 {
				err = fmt.Errorf("listen port is not correct, listener: %s.%s", slbName, listenerName)
				return
			}

			if alreadyListendPorts[int(listenPort)] {

				logrus.WithField("CODE", p.Code).
					WithField("SLB-NAME", slbName).
					WithField("SLB-ID", slbInstance.LoadBalancerId).
					WithField("SLB-LISTENER", listenerName).
					WithField("PORT", listenPort).Infoln("Listener already created")

				continue
			}

			arg := &CreateLoadBalancerHTTPListenerArgs{

				slbClient: p.SLBClient(),

				serverCertificateId: listenerConfig.GetString("server-certificate-id"),

				HTTPListenerType: slb.HTTPListenerType{
					LoadBalancerId:         slbInstance.LoadBalancerId,
					ListenerPort:           int(listenPort),
					BackendServerPort:      int(listenerConfig.GetInt64("server-port")),
					Bandwidth:              int(listenerConfig.GetInt64("band-width")),
					Scheduler:              slb.SchedulerType(listenerConfig.GetString("scheduler", "wrr")),
					Gzip:                   slb.FlagType(listenerConfig.GetString("gzip", "on")),
					StickySession:          slb.FlagType(listenerConfig.GetString("sticky-session", "off")),
					StickySessionType:      slb.StickySessionType(listenerConfig.GetString("sticky-session-type", "insert")),
					CookieTimeout:          int(listenerConfig.GetInt64("cookie-timeout", 86400)),
					Cookie:                 listenerConfig.GetString("cookie"),
					HealthCheck:            slb.FlagType(listenerConfig.GetString("health-check.check", "on")),
					HealthCheckDomain:      listenerConfig.GetString("health-check.domain"),
					HealthCheckURI:         listenerConfig.GetString("health-check.url"),
					HealthCheckConnectPort: int(listenerConfig.GetInt64("health-check.connect-port")),
					HealthyThreshold:       int(listenerConfig.GetInt64("health-check.threshold", 3)),
					UnhealthyThreshold:     int(listenerConfig.GetInt64("health-check.unhealthy-threshold", 3)),
					HealthCheckTimeout:     int(listenerConfig.GetInt64("health-check.timeout", 5)),
					HealthCheckInterval:    int(listenerConfig.GetInt64("health-check.interval", 2)),
					HealthCheckHttpCode:    slb.HealthCheckHttpCodeType(listenerConfig.GetString("health-check.http-code", "http_2xx")),
					VServerGroup:           slb.FlagType(listenerConfig.GetString("vserver-group", "on")),
					VServerGroupId:         listenerConfig.GetString("vserver-group-id"),
					XForwardedFor_SLBID:    slb.FlagType(listenerConfig.GetString("x-forward-for-slb-id", "on")),
					XForwardedFor_SLBIP:    slb.FlagType(listenerConfig.GetString("x-forward-for-slb-ip", "on")),
					XForwardedFor_proto:    slb.FlagType(listenerConfig.GetString("x-forward-for-proto", "on")),
				},
			}

			args = append(args, arg)
		}
	}

	createArgs = args

	return
}

func (p *Aliyun) CreateLoadBalancerSocketListenerArgs(isTCP bool) (createArgs []*CreateLoadBalancerSocketListenerArgs, err error) {
	currentLBSs, err := p.ListLoadBalancers(true)
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

	var args []*CreateLoadBalancerSocketListenerArgs

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			err = fmt.Errorf("slb of %s not exist", slbName)
			return
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenerConfName := "tcp-listener"

		if !isTCP {
			listenerConfName = "udp-listener"
		}

		listenersConfig := lbConfig.GetConfig(listenerConfName)

		if listenersConfig.IsEmpty() {
			continue
		}

		var alreadyListendPorts = make(map[int]bool)

		for _, port := range slbInstance.ListenerPorts.ListenerPort {
			alreadyListendPorts[port] = true
		}

		for _, listenerName := range listenersConfig.Keys() {

			listenerConfig := listenersConfig.GetConfig(listenerName)

			if listenerConfig.IsEmpty() {
				err = fmt.Errorf("listen config is empty: %s.%s", slbName, listenerName)
				return
			}

			listenPort := listenerConfig.GetInt32("listen-port")

			if listenPort <= 0 {
				err = fmt.Errorf("listen port is not correct, listener: %s.%s", slbName, listenerName)
				return
			}

			if alreadyListendPorts[int(listenPort)] {

				logrus.WithField("CODE", p.Code).
					WithField("SLB-NAME", slbName).
					WithField("SLB-ID", slbInstance.LoadBalancerId).
					WithField("SLB-LISTENER", listenerName).
					WithField("IS-TCP", isTCP).
					WithField("PORT", listenPort).Infoln("Listener already created")

				continue
			}

			arg := &CreateLoadBalancerSocketListenerArgs{

				slbClient: p.SLBClient(),

				isTCP: isTCP,

				// Common part
				LoadBalancerId:            slbInstance.LoadBalancerId,
				ListenerPort:              int(listenPort),
				BackendServerPort:         int(listenerConfig.GetInt64("server-port")),
				Bandwidth:                 int(listenerConfig.GetInt64("band-width")),
				Scheduler:                 slb.SchedulerType(listenerConfig.GetString("scheduler", "wrr")),
				PersistenceTimeout:        int(listenerConfig.GetInt64("persistence-timeout")),
				HealthCheck:               slb.FlagType(listenerConfig.GetString("health-check.check", "on")),
				HealthCheckConnectPort:    int(listenerConfig.GetInt64("health-check.connect-port")),
				HealthyThreshold:          int(listenerConfig.GetInt64("health-check.threshold", 3)),
				UnhealthyThreshold:        int(listenerConfig.GetInt64("health-check.unhealthy-threshold", 3)),
				HealthCheckConnectTimeout: int(listenerConfig.GetInt64("health-check.timeout", 5)),
				HealthCheckInterval:       int(listenerConfig.GetInt64("health-check.interval", 2)),
				VServerGroup:              slb.FlagType(listenerConfig.GetString("vserver-group", "on")),
				VServerGroupId:            listenerConfig.GetString("vserver-group-id"),

				// TCP Part
				HealthCheckType:     slb.HealthCheckType(listenerConfig.GetString("health-check.type", "tcp")),
				HealthCheckDomain:   listenerConfig.GetString("health-check.domain"),
				HealthCheckURI:      listenerConfig.GetString("health-check.url"),
				HealthCheckHttpCode: slb.HealthCheckHttpCodeType(listenerConfig.GetString("health-check.http-code", "http_2xx")),
			}

			args = append(args, arg)
		}
	}

	createArgs = args

	return
}

func (p *Aliyun) ListSLBHTTPListeners() (listerners map[string][]*slb.DescribeLoadBalancerHTTPListenerAttributeResponse, err error) {

	currentLBSs, err := p.ListLoadBalancers(true)
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

	slbListeners := make(map[string][]*slb.DescribeLoadBalancerHTTPListenerAttributeResponse)

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			continue
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("http-listener")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, port := range slbInstance.ListenerPorts.ListenerPort {
			var resp *slb.DescribeLoadBalancerHTTPListenerAttributeResponse
			resp, err = p.SLBClient().DescribeLoadBalancerHTTPListenerAttribute(slbInstance.LoadBalancerId, port)
			if err != nil {
				return
			}

			slbListeners[slbName] = append(slbListeners[slbName], resp)
		}
	}

	listerners = slbListeners

	return
}

type Rule struct {
	RuleName       string
	Domain         string
	Url            string `json:",omitempty"`
	VServerGroupId string
}

func (p *Aliyun) CreateSLBHTTPListenerRuleArgs() (createArgs []*slb.CreateRulesArgs, err error) {

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

	var args []*slb.CreateRulesArgs

	for _, slbName := range slbNames {

		slbInstance, exist := balancers[slbName]

		if !exist {
			err = fmt.Errorf("slb instance not exist: %s", slbName)
			return
		}

		var alreadyListendPorts = make(map[int]bool)

		for _, port := range slbInstance.ListenerPorts.ListenerPort {
			alreadyListendPorts[port] = true
		}

		lbConfig := balancersConfig.GetConfig(slbName)
		listenersConfig := lbConfig.GetConfig("http-listener")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, listenerName := range listenersConfig.Keys() {
			listenerConf := listenersConfig.GetConfig(listenerName)

			if listenerConf.IsEmpty() {
				continue
			}

			port := int(listenerConf.GetInt32("listen-port"))

			if !alreadyListendPorts[port] {
				err = fmt.Errorf("port %d not listened in balance %s", port, slbName)
				return
			}

			var vSrvGroupsResp *slb.DescribeVServerGroupsResponse
			vSrvGroupsResp, err = p.SLBClient().DescribeVServerGroups(&slb.DescribeVServerGroupsArgs{slbInstance.LoadBalancerId, p.Region})
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

			var ruleDescribRep *slb.DescribeRulesResponse
			ruleDescribRep, err = p.SLBClient().DescribeRules(&slb.DescribeRulesArgs{
				RegionId:       p.Region,
				LoadBalancerId: slbInstance.LoadBalancerId,
				ListenerPort:   port,
			})

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

			arg := &slb.CreateRulesArgs{
				RegionId:       p.Region,
				LoadBalancerId: slbInstance.LoadBalancerId,
				ListenerPort:   port,
				RuleList:       string(ruleData),
			}

			args = append(args, arg)
		}
	}

	createArgs = args

	return
}
