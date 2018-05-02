package aliyun

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
)

type SearchECSInstanceArgs struct {
	InstanceId   string
	InstanceName string
	VPCName      string
	VSwitchName  string
	ZoneId       string
	NetworkType  string

	Tags []Tag

	vpcId     string
	vswitchId string
}

func (p *Aliyun) FindECSInstance(arg *SearchECSInstanceArgs) (inst *ecs.Instance, err error) {

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

	req := ecs.CreateDescribeInstancesRequest()

	req.RegionId = p.Region
	req.InstanceIds = arg.InstanceId
	req.InstanceName = arg.InstanceName
	req.InstanceNetworkType = arg.NetworkType
	req.ZoneId = arg.ZoneId
	req.VSwitchId = arg.vswitchId
	req.VpcId = arg.vpcId

	for i, tag := range arg.Tags {
		switch i {
		case 0:
			req.Tag1Key = tag.Key
			req.Tag1Value = tag.Value
		case 1:
			req.Tag2Key = tag.Key
			req.Tag2Value = tag.Value
		case 2:
			req.Tag3Key = tag.Key
			req.Tag3Value = tag.Value
		case 3:
			req.Tag4Key = tag.Key
			req.Tag4Value = tag.Value
		case 4:
			req.Tag5Key = tag.Key
			req.Tag5Value = tag.Value
		}
	}

	resp, err := p.ECSClient().DescribeInstances(req)
	if err != nil {
		return
	}

	if len(resp.Instances.Instance) == 0 {
		return
	}

	if len(resp.Instances.Instance) > 1 {
		err = fmt.Errorf("find more than one instance")
		return
	}

	inst = &resp.Instances.Instance[0]

	return
}
