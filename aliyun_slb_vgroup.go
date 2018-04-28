package aliyun

import (
	"encoding/json"
	"fmt"

	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"
	"github.com/sirupsen/logrus"

	"github.com/denverdino/aliyungo/common"
)

func (p *Aliyun) CreateVServerGroupArgs() (createArgs []*slb.CreateVServerGroupArgs, err error) {
	balancersConfig := p.Config.GetConfig("aliyun.slb.balancer")

	if balancersConfig.IsEmpty() {
		return
	}

	lbs, err := p.ListLoadBalancers(true)
	if err != nil {
		return
	}

	var args []*slb.CreateVServerGroupArgs

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

		var existGroups *slb.DescribeVServerGroupsResponse
		existGroups, err = p.SLBClient().DescribeVServerGroups(&slb.DescribeVServerGroupsArgs{
			LoadBalancerId: lb.LoadBalancerId,
			RegionId:       common.Region(p.Region),
		})

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

				tag := map[string]string{}

				tagConf := serarchConf.GetConfig("tag")

				for _, k := range tagConf.Keys() {
					tag[k] = tagConf.GetString(k)
				}

				var inst *ecs.InstanceAttributesType
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
						Tag:          tag,
					},
				)

				if err != nil {
					return
				}

				if inst == nil {
					err = fmt.Errorf("instance '%s' not found: %s.%s.%s, tags: %#v", serarchConf.GetString("name"), balancerName, groupName, srv, tag)
					return
				}

				var backendServers []slb.VBackendServerType

				portsConf := groupConf.GetConfig(srv + ".ports")

				for _, portName := range portsConf.Keys() {
					portConf := portsConf.GetConfig(portName)
					vSrv := slb.VBackendServerType{
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

				arg := &slb.CreateVServerGroupArgs{
					LoadBalancerId:   lb.LoadBalancerId,
					RegionId:         common.Region(p.Region),
					VServerGroupName: groupName,
					BackendServers:   string(srvData),
				}

				logrus.WithField("SLB-INSTANCE-NAME", balancerName).
					WithField("SLB-VGROUP-NAME", groupName).
					WithField("SLB-VGROUP-VSERVER", string(srvData)).
					Debugln("SLB VServerGroup Info")

				args = append(args, arg)
			}
		}
	}

	createArgs = args

	return
}
