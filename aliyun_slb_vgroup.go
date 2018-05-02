package aliyun

import (
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/sirupsen/logrus"
)

func (p *Aliyun) CreateVServerGroup() (err error) {
	balancersConfig := p.Config.GetConfig("aliyun.slb.balancer")

	if balancersConfig.IsEmpty() {
		return
	}

	lbs, err := p.ListLoadBalancers(true)
	if err != nil {
		return
	}

	var reqs []*slb.CreateVServerGroupRequest

	for _, balancerName := range balancersConfig.Keys() {

		lb, exist := lbs[balancerName]
		if !exist {
			err = fmt.Errorf("could not find slb balancer: %s", balancerName)
			return
		}

		vpcName := balancersConfig.GetString(balancerName + ".vpc-name")
		vSwitchName := balancersConfig.GetString(balancerName + ".vswitch-name")

		vServerGroupsConf := balancersConfig.GetConfig(balancerName + ".vserver-group")

		if vServerGroupsConf.IsEmpty() {
			continue
		}

		describVgroupReq := slb.CreateDescribeVServerGroupsRequest()

		describVgroupReq.LoadBalancerId = lb.LoadBalancerId
		describVgroupReq.RegionId = p.Region

		var existGroups *slb.DescribeVServerGroupsResponse
		existGroups, err = p.SLBClient().DescribeVServerGroups(describVgroupReq)

		if err != nil {
			return
		}

		mapExistGroups := map[string]bool{}

		if existGroups != nil {
			for _, group := range existGroups.VServerGroups.VServerGroup {
				mapExistGroups[group.VServerGroupName] = true
			}
		}

		for _, groupName := range vServerGroupsConf.Keys() {

			if mapExistGroups[groupName] {

				logrus.WithField("SLB-INSTANCE-NAME", balancerName).
					WithField("SLB-VGROUP-NAME", groupName).
					Infoln("SLB VServerGroup already exist")

				continue
			}

			groupConf := vServerGroupsConf.GetConfig(groupName)

			for _, srv := range groupConf.Keys() {

				serarchConf := groupConf.GetConfig(srv + ".instance")

				var tags []Tag

				tagConf := serarchConf.GetConfig("tag")

				for _, k := range tagConf.Keys() {
					tags = append(tags, Tag{Key: k, Value: tagConf.GetString(k)})
				}

				var inst *ecs.Instance
				inst, err = p.FindECSInstance(
					&SearchECSInstanceArgs{
						InstanceId:   serarchConf.GetString("id"),
						InstanceName: serarchConf.GetString("name"),
						NetworkType:  serarchConf.GetString("network-type"),
						ZoneId:       serarchConf.GetString("zone-id"),
						VPCName:      vpcName,
						VSwitchName:  vSwitchName,
						vswitchId:    lb.VSwitchId,
						vpcId:        lb.VpcId,
						Tags:         tags,
					},
				)

				if err != nil {
					return
				}

				if inst == nil {
					err = fmt.Errorf("instance '%s' not found: %s.%s.%s, tags: %#v", serarchConf.GetString("name"), balancerName, groupName, srv, tags)
					return
				}

				var backendServers []slb.BackendServer

				portsConf := groupConf.GetConfig(srv + ".ports")

				for _, portName := range portsConf.Keys() {
					portConf := portsConf.GetConfig(portName)

					vSrv := slb.BackendServer{
						ServerId: inst.InstanceId,
						Port:     int(portConf.GetInt32("port")),
						Weight:   int(portConf.GetInt32("weight")),
					}

					backendServers = append(backendServers, vSrv)
				}

				var srvData []byte
				srvData, err = json.Marshal(backendServers)

				if err != nil {
					return
				}

				req := slb.CreateCreateVServerGroupRequest()

				req.LoadBalancerId = lb.LoadBalancerId
				req.RegionId = p.Region
				req.VServerGroupName = groupName
				req.BackendServers = string(srvData)

				logrus.WithField("SLB-INSTANCE-NAME", balancerName).
					WithField("SLB-VGROUP-NAME", groupName).
					WithField("SLB-VGROUP-VSERVER", string(srvData)).
					Debugln("SLB VServerGroup Info")

				reqs = append(reqs, req)
			}
		}
	}

	for i := 0; i < len(reqs); i++ {

		var resp *slb.CreateVServerGroupResponse

		resp, err = p.SLBClient().CreateVServerGroup(reqs[i])
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("SLB-BANLANCER-ID", reqs[i].LoadBalancerId).
			WithField("SLB-BANLANCER-VGROUP-NAME", reqs[i].VServerGroupName).
			WithField("SLB-BANLANCER-VGROUP-ID", resp.VServerGroupId).
			Infoln("SLB VGroup created")
	}

	return
}
