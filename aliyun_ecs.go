package aliyun

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
)

type SearchECSInstanceArgs struct {
	InstanceId   string
	InstanceName string
	VPCName      string
	VSwitchName  string
	ZoneId       string
	NetworkType  string
	Tag          map[string]string

	vpcId     string
	vswitchId string
}

func (p *Aliyun) FindECSInstance(arg *SearchECSInstanceArgs) (inst *ecs.InstanceAttributesType, err error) {

	if len(arg.vpcId)+len(arg.vswitchId) == 0 {
		if len(arg.VPCName) > 0 && len(arg.VSwitchName) > 0 {
			var vSwitch *vpc.VSwitch
			vSwitch, err = p.FindVSwitch(arg.VPCName, arg.VSwitchName)
			if err != nil {
				return
			}

			if vSwitch != nil {
				arg.vpcId = vSwitch.VpcId
				arg.vswitchId = vSwitch.VSwitchId
			}
		}
	}

	instances, _, err := p.ECSClient().DescribeInstances(
		&ecs.DescribeInstancesArgs{
			RegionId:            common.Region(p.Region),
			InstanceIds:         arg.InstanceId,
			InstanceName:        arg.InstanceName,
			InstanceNetworkType: arg.NetworkType,
			ZoneId:              arg.ZoneId,
			Tag:                 arg.Tag,
			VSwitchId:           arg.vswitchId,
			VpcId:               arg.vpcId,
		},
	)

	if err != nil {
		return
	}

	if len(instances) == 0 {
		return
	}

	if len(instances) > 1 {
		err = fmt.Errorf("find more than one instance")
		return
	}

	inst = &instances[0]

	return
}
