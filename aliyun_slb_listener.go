package aliyun

import (
	"fmt"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/sirupsen/logrus"
)

func (p *Aliyun) startSLBListener(loadBalancerId string, port requests.Integer) (err error) {

	req := slb.CreateStartLoadBalancerListenerRequest()

	req.LoadBalancerId = loadBalancerId
	req.RegionId = p.Region
	req.ListenerPort = port

	_, err = p.SLBClient().StartLoadBalancerListener(req)
	if err != nil {
		return
	}

	logrus.WithField("CODE", p.Code).
		WithField("SLB-ID", loadBalancerId).
		WithField("PORT", port).Infoln("Listener started")

	return
}

func (p *Aliyun) CreateLoadBalancerHTTPListener() (err error) {
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

	var reqs []*slb.CreateLoadBalancerHTTPListenerRequest

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			err = fmt.Errorf("slb of %s not exist", slbName)
			return
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.http")

		if listenersConfig.IsEmpty() {
			continue
		}

		var alreadyListendPorts = make(map[string]bool)

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

			if alreadyListendPorts[strconv.Itoa(int(listenPort))] {

				logrus.WithField("CODE", p.Code).
					WithField("SLB-NAME", slbName).
					WithField("SLB-ID", slbInstance.LoadBalancerId).
					WithField("SLB-LISTENER", listenerName).
					WithField("PORT", listenPort).Infoln("Listener already created")

				continue
			}

			req := slb.CreateCreateLoadBalancerHTTPListenerRequest()

			// serverCertificateId: listenerConfig.GetString("server-certificate-id"),

			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.ListenerPort = requests.NewInteger(int(listenPort))
			req.BackendServerPort = requests.NewInteger(int(listenerConfig.GetInt64("server-port")))
			req.Bandwidth = requests.NewInteger(int(listenerConfig.GetInt64("band-width")))
			req.Scheduler = listenerConfig.GetString("scheduler", "wrr")
			req.Gzip = listenerConfig.GetString("gzip", "on")
			req.StickySession = listenerConfig.GetString("sticky-session", "off")
			req.StickySessionType = listenerConfig.GetString("sticky-session-type", "insert")
			req.CookieTimeout = requests.NewInteger(int(listenerConfig.GetInt64("cookie-timeout", 86400)))
			req.Cookie = listenerConfig.GetString("cookie")
			req.HealthCheck = listenerConfig.GetString("health-check.check", "on")
			req.HealthCheckDomain = listenerConfig.GetString("health-check.domain")
			req.HealthCheckURI = listenerConfig.GetString("health-check.url")
			req.HealthCheckConnectPort = requests.NewInteger(int(listenerConfig.GetInt32("health-check.connect-port", listenPort)))
			req.HealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.threshold", 3)))
			req.UnhealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.unhealthy-threshold", 3)))
			req.HealthCheckTimeout = requests.NewInteger(int(listenerConfig.GetInt64("health-check.timeout", 5)))
			req.HealthCheckInterval = requests.NewInteger(int(listenerConfig.GetInt64("health-check.interval", 2)))
			req.HealthCheckHttpCode = listenerConfig.GetString("health-check.http-code", "http_2xx")
			req.VServerGroupId = listenerConfig.GetString("vserver-group-id")
			req.XForwardedForSLBID = listenerConfig.GetString("x-forward-for-slb-id", "on")
			req.XForwardedForSLBIP = listenerConfig.GetString("x-forward-for-slb-ip", "on")
			req.XForwardedForProto = listenerConfig.GetString("x-forward-for-proto", "on")

			reqs = append(reqs, req)
		}
	}

	for i := 0; i < len(reqs); i++ {

		_, err = p.SLBClient().CreateLoadBalancerHTTPListener(reqs[i])
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", reqs[i].LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", reqs[i].ListenerPort).
			Infoln("SLB http listener created")

		err = p.startSLBListener(reqs[i].LoadBalancerId, reqs[i].ListenerPort)
		if err != nil {
			return
		}
	}

	return
}

func (p *Aliyun) CreateLoadBalancerHTTPSListener() (err error) {
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

	var reqs []*slb.CreateLoadBalancerHTTPSListenerRequest

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			err = fmt.Errorf("slb of %s not exist", slbName)
			return
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.https")

		if listenersConfig.IsEmpty() {
			continue
		}

		var alreadyListendPorts = make(map[string]bool)

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

			if alreadyListendPorts[strconv.Itoa(int(listenPort))] {

				logrus.WithField("CODE", p.Code).
					WithField("SLB-NAME", slbName).
					WithField("SLB-ID", slbInstance.LoadBalancerId).
					WithField("SLB-LISTENER", listenerName).
					WithField("PORT", listenPort).Infoln("Listener already created")

				continue
			}

			req := slb.CreateCreateLoadBalancerHTTPSListenerRequest()

			req.ServerCertificateId = listenerConfig.GetString("server-certificate-id")
			req.CACertificateId = listenerConfig.GetString("ca-certificate-id")
			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.ListenerPort = requests.NewInteger(int(listenPort))
			req.BackendServerPort = requests.NewInteger(int(listenerConfig.GetInt64("server-port")))
			req.Bandwidth = requests.NewInteger(int(listenerConfig.GetInt64("band-width")))
			req.Scheduler = listenerConfig.GetString("scheduler", "wrr")
			req.Gzip = listenerConfig.GetString("gzip", "on")
			req.StickySession = listenerConfig.GetString("sticky-session", "off")
			req.StickySessionType = listenerConfig.GetString("sticky-session-type", "insert")
			req.CookieTimeout = requests.NewInteger(int(listenerConfig.GetInt64("cookie-timeout", 86400)))
			req.Cookie = listenerConfig.GetString("cookie")
			req.HealthCheck = listenerConfig.GetString("health-check.check", "on")
			req.HealthCheckDomain = listenerConfig.GetString("health-check.domain")
			req.HealthCheckURI = listenerConfig.GetString("health-check.url")
			req.HealthCheckConnectPort = requests.NewInteger(int(listenerConfig.GetInt32("health-check.connect-port", listenPort)))
			req.HealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.threshold", 3)))
			req.UnhealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.unhealthy-threshold", 3)))
			req.HealthCheckTimeout = requests.NewInteger(int(listenerConfig.GetInt64("health-check.timeout", 5)))
			req.HealthCheckInterval = requests.NewInteger(int(listenerConfig.GetInt64("health-check.interval", 2)))
			req.HealthCheckHttpCode = listenerConfig.GetString("health-check.http-code", "http_2xx")
			req.VServerGroupId = listenerConfig.GetString("vserver-group-id")
			req.XForwardedForSLBID = listenerConfig.GetString("x-forward-for-slb-id", "on")
			req.XForwardedForSLBIP = listenerConfig.GetString("x-forward-for-slb-ip", "on")
			req.XForwardedForProto = listenerConfig.GetString("x-forward-for-proto", "on")

			reqs = append(reqs, req)
		}
	}

	for i := 0; i < len(reqs); i++ {

		_, err = p.SLBClient().CreateLoadBalancerHTTPSListener(reqs[i])
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", reqs[i].LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", reqs[i].ListenerPort).
			Infoln("SLB https listener created")

		err = p.startSLBListener(reqs[i].LoadBalancerId, reqs[i].ListenerPort)
		if err != nil {
			return
		}
	}

	return
}

func (p *Aliyun) CreateLoadBalancerTCPListener() (err error) {
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

	var reqs []*slb.CreateLoadBalancerTCPListenerRequest

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			err = fmt.Errorf("slb of %s not exist", slbName)
			return
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.tcp")

		if listenersConfig.IsEmpty() {
			continue
		}

		var alreadyListendPorts = make(map[string]bool)

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

			if alreadyListendPorts[strconv.Itoa(int(listenPort))] {

				logrus.WithField("CODE", p.Code).
					WithField("SLB-NAME", slbName).
					WithField("SLB-ID", slbInstance.LoadBalancerId).
					WithField("SLB-LISTENER", listenerName).
					WithField("PORT", listenPort).Infoln("Listener already created")

				continue
			}

			req := slb.CreateCreateLoadBalancerTCPListenerRequest()

			// Common part
			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.ListenerPort = requests.NewInteger(int(listenPort))
			req.BackendServerPort = requests.NewInteger(int(listenerConfig.GetInt64("server-port")))
			req.Bandwidth = requests.NewInteger(int(listenerConfig.GetInt64("band-width")))
			req.Scheduler = listenerConfig.GetString("scheduler", "wrr")
			req.PersistenceTimeout = requests.NewInteger(int(listenerConfig.GetInt64("persistence-timeout")))
			req.HealthCheckConnectPort = requests.NewInteger(int(listenerConfig.GetInt32("health-check.connect-port", listenPort)))
			req.HealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.threshold", 3)))
			req.UnhealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.unhealthy-threshold", 3)))
			req.HealthCheckConnectTimeout = requests.NewInteger(int(listenerConfig.GetInt64("health-check.timeout", 5)))
			req.HealthCheckInterval = requests.NewInteger(int(listenerConfig.GetInt64("health-check.interval", 2)))
			req.VServerGroupId = listenerConfig.GetString("vserver-group-id")

			// TCP Part
			req.HealthCheckType = listenerConfig.GetString("health-check.type", "tcp")
			req.HealthCheckDomain = listenerConfig.GetString("health-check.domain")
			req.HealthCheckURI = listenerConfig.GetString("health-check.url")
			req.HealthCheckHttpCode = listenerConfig.GetString("health-check.http-code", "http_2xx")

			reqs = append(reqs, req)
		}
	}

	for i := 0; i < len(reqs); i++ {

		_, err = p.SLBClient().CreateLoadBalancerTCPListener(reqs[i])
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", reqs[i].LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", reqs[i].ListenerPort).
			Infoln("SLB TCP listener created")

		err = p.startSLBListener(reqs[i].LoadBalancerId, reqs[i].ListenerPort)
		if err != nil {
			return
		}
	}

	return
}

func (p *Aliyun) CreateLoadBalancerUDPListener() (err error) {
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

	var reqs []*slb.CreateLoadBalancerUDPListenerRequest

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			err = fmt.Errorf("slb of %s not exist", slbName)
			return
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.udp")

		if listenersConfig.IsEmpty() {
			continue
		}

		var alreadyListendPorts = make(map[string]bool)

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

			if alreadyListendPorts[strconv.Itoa(int(listenPort))] {

				logrus.WithField("CODE", p.Code).
					WithField("SLB-NAME", slbName).
					WithField("SLB-ID", slbInstance.LoadBalancerId).
					WithField("SLB-LISTENER", listenerName).
					WithField("UDP-PORT", listenPort).Infoln("Listener already created")

				continue
			}

			req := slb.CreateCreateLoadBalancerUDPListenerRequest()

			// Common part
			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.ListenerPort = requests.NewInteger(int(listenPort))
			req.BackendServerPort = requests.NewInteger(int(listenerConfig.GetInt64("server-port")))
			req.Bandwidth = requests.NewInteger(int(listenerConfig.GetInt64("band-width")))
			req.Scheduler = listenerConfig.GetString("scheduler", "wrr")
			req.PersistenceTimeout = requests.NewInteger(int(listenerConfig.GetInt64("persistence-timeout")))
			req.HealthCheckConnectPort = requests.NewInteger(int(listenerConfig.GetInt32("health-check.connect-port", listenPort)))
			req.HealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.threshold", 3)))
			req.UnhealthyThreshold = requests.NewInteger(int(listenerConfig.GetInt64("health-check.unhealthy-threshold", 3)))
			req.HealthCheckConnectTimeout = requests.NewInteger(int(listenerConfig.GetInt64("health-check.timeout", 5)))
			req.HealthCheckInterval = requests.NewInteger(int(listenerConfig.GetInt64("health-check.interval", 2)))
			req.VServerGroupId = listenerConfig.GetString("vserver-group-id")

			reqs = append(reqs, req)
		}
	}

	for i := 0; i < len(reqs); i++ {
		_, err = p.SLBClient().CreateLoadBalancerUDPListener(reqs[i])
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", reqs[i].LoadBalancerId).
			WithField("SLB-BANLANCER-LISTEN-PORT", reqs[i].ListenerPort).
			Infoln("SLB UDP listener created")

		err = p.startSLBListener(reqs[i].LoadBalancerId, reqs[i].ListenerPort)
		if err != nil {
			return
		}
	}

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

		listenersConfig := lbConfig.GetConfig("listener.http")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, port := range slbInstance.ListenerPorts.ListenerPort {

			req := slb.CreateDescribeLoadBalancerHTTPListenerAttributeRequest()

			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.Port = port

			var resp *slb.DescribeLoadBalancerHTTPListenerAttributeResponse
			resp, err = p.SLBClient().DescribeLoadBalancerHTTPListenerAttribute(req)
			if err != nil {
				return
			}

			slbListeners[slbName] = append(slbListeners[slbName], resp)
		}
	}

	listerners = slbListeners

	return
}

func (p *Aliyun) ListSLBHTTPSListeners() (listerners map[string][]*slb.DescribeLoadBalancerHTTPSListenerAttributeResponse, err error) {

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

	slbListeners := make(map[string][]*slb.DescribeLoadBalancerHTTPSListenerAttributeResponse)

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			continue
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.https")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, port := range slbInstance.ListenerPorts.ListenerPort {

			req := slb.CreateDescribeLoadBalancerHTTPSListenerAttributeRequest()

			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.Port = port

			var resp *slb.DescribeLoadBalancerHTTPSListenerAttributeResponse
			resp, err = p.SLBClient().DescribeLoadBalancerHTTPSListenerAttribute(req)
			if err != nil {
				return
			}

			slbListeners[slbName] = append(slbListeners[slbName], resp)
		}
	}

	listerners = slbListeners

	return
}

func (p *Aliyun) ListSLBTCPListeners() (listerners map[string][]*slb.DescribeLoadBalancerTCPListenerAttributeResponse, err error) {

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

	slbListeners := make(map[string][]*slb.DescribeLoadBalancerTCPListenerAttributeResponse)

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			continue
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.tcp")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, port := range slbInstance.ListenerPorts.ListenerPort {

			req := slb.CreateDescribeLoadBalancerTCPListenerAttributeRequest()

			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.Port = port

			var resp *slb.DescribeLoadBalancerTCPListenerAttributeResponse
			resp, err = p.SLBClient().DescribeLoadBalancerTCPListenerAttribute(req)
			if err != nil {
				return
			}

			slbListeners[slbName] = append(slbListeners[slbName], resp)
		}
	}

	listerners = slbListeners

	return
}

func (p *Aliyun) ListSLBUDPListeners() (listerners map[string][]*slb.DescribeLoadBalancerUDPListenerAttributeResponse, err error) {

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

	slbListeners := make(map[string][]*slb.DescribeLoadBalancerUDPListenerAttributeResponse)

	for _, slbName := range slbNames {

		slbInstance, exist := currentLBSs[slbName]

		if !exist {
			continue
		}

		lbConfig := balancersConfig.GetConfig(slbName)

		listenersConfig := lbConfig.GetConfig("listener.tcp")

		if listenersConfig.IsEmpty() {
			continue
		}

		for _, port := range slbInstance.ListenerPorts.ListenerPort {

			req := slb.CreateDescribeLoadBalancerUDPListenerAttributeRequest()

			req.LoadBalancerId = slbInstance.LoadBalancerId
			req.Port = port

			var resp *slb.DescribeLoadBalancerUDPListenerAttributeResponse
			resp, err = p.SLBClient().DescribeLoadBalancerUDPListenerAttribute(req)
			if err != nil {
				return
			}

			slbListeners[slbName] = append(slbListeners[slbName], resp)
		}
	}

	listerners = slbListeners

	return
}
