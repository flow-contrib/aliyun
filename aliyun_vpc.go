package aliyun

import (
	"fmt"
	"strings"
	"sync"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/sirupsen/logrus"
)

func (p *Aliyun) CreateVPCs() (err error) {

	vpcsConf := p.Config.GetConfig("aliyun.ecs.vpc")

	if vpcsConf.IsEmpty() {
		return
	}

	var sets []ecs.VpcSetType
	sets, _, err = p.ECSClient().DescribeVpcs(
		&ecs.DescribeVpcsArgs{
			RegionId: common.Region(p.Region),
		})

	if err != nil {
		return
	}

	var args []*ecs.CreateVpcArgs

	for _, vpcName := range vpcsConf.Keys() {

		vpcConf := vpcsConf.GetConfig(vpcName)

		vpcId := vpcConf.GetString("id")

		if len(vpcId) > 0 {
			continue
		}

		created := false

		desc := vpcConf.GetString("description")

		arg := &ecs.CreateVpcArgs{
			RegionId:    common.Region(p.Region),
			VpcName:     vpcName,
			CidrBlock:   vpcConf.GetString("cidr-block", "172.16.0.0/16"),
			Description: p.signWithCode(desc),
		}

		for _, s := range sets {
			if s.VpcName == arg.VpcName &&
				s.CidrBlock == arg.CidrBlock &&
				s.RegionId == arg.RegionId &&
				p.isSignd(s.Description) {

				created = true
				vpcId = s.VpcId
				break
			}
		}

		if created == true {
			logrus.WithField("CODE", p.Code).WithField("VPCID", vpcId).Infoln("VPC already created")
			continue
		}

		args = append(args, arg)
	}

	for _, arg := range args {

		resp, e := p.ECSClient().CreateVpc(arg)
		if e != nil {
			return e
		}

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VPC-NAME", arg.VpcName).
			WithField("ECS-VPC-ID", resp.VpcId).
			WithField("ECS-VPC-REGION", arg.RegionId).
			Infoln("VPC created")
	}

	return
}

func (p *Aliyun) DeleteVPCArgs() (err error) {
	vpcsConf := p.Config.GetConfig("aliyun.ecs.vpc")

	if vpcsConf.IsEmpty() {
		return
	}

	var args []*ecs.DeleteVpcArgs

	var sets []ecs.VpcSetType
	sets, _, err = p.ECSClient().DescribeVpcs(&ecs.DescribeVpcsArgs{
		RegionId: common.Region(p.Region),
	})

	if err != nil {
		return
	}

	for _, vpcName := range vpcsConf.Keys() {

		vpcConf := vpcsConf.GetConfig(vpcName)

		vpcId := vpcConf.GetString("id")

		if len(vpcId) == 0 {
			for _, s := range sets {
				if s.CidrBlock == vpcConf.GetString("cidr-block", "172.16.0.0/16") &&
					s.RegionId == p.Region &&
					s.VpcName == vpcName &&
					p.isSignd(s.Description) {

					vpcId = s.VpcId
					logrus.WithField("CODE", p.Code).WithField("VPCID", vpcId).WithField("NAME", s.VpcName).Infoln("VPC found at aliyun")
					break
				}
			}
		}

		if len(vpcId) > 0 {
			arg := &ecs.DeleteVpcArgs{
				VpcId: vpcId,
			}

			args = append(args, arg)
		}
	}

	for _, arg := range args {

		err = p.ECSClient().DeleteVpc(arg.VpcId)
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VPC-ID", arg.VpcId).
			Infoln("VPC deleted")
	}

	return
}

func (p *Aliyun) WaitForAllVpcRunning(timeout int) (err error) {

	vpcsConf := p.Config.GetConfig("aliyun.ecs.vpc")

	if vpcsConf.IsEmpty() {
		return
	}

	var sets []ecs.VpcSetType
	sets, _, err = p.ECSClient().DescribeVpcs(
		&ecs.DescribeVpcsArgs{
			RegionId: common.Region(p.Region),
		})

	if err != nil {
		return
	}

	mapVpcs := map[string]string{}

	for _, s := range sets {
		mapVpcs[s.VpcName] = s.VpcId
	}

	var vpcIds []string

	for _, vpcName := range vpcsConf.Keys() {

		vpcId, exist := mapVpcs[vpcName]

		if !exist {
			continue
		}

		vpcIds = append(vpcIds, vpcId)
	}

	if len(vpcIds) == 0 {
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(len(vpcIds))

	for i := 0; i < len(vpcIds); i++ {
		go func(vpcId string) {
			defer wg.Done()
			p.ECSClient().WaitForVpcAvailable(common.Region(p.Region), vpcId, timeout)
		}(vpcIds[i])
	}

	logrus.WithField("CODE", p.Code).Infoln("Wait for all VPC available")

	wg.Wait()

	return
}

func (p *Aliyun) FindVPC(vpcName string) (ret *ecs.VpcSetType, err error) {
	var vpcSets []ecs.VpcSetType
	vpcSets, _, err = p.ECSClient().DescribeVpcs(
		&ecs.DescribeVpcsArgs{
			RegionId: common.Region(p.Region),
		})

	if err != nil {
		return
	}

	for i, vpc := range vpcSets {
		if vpcName == vpc.VpcName &&
			p.isSignd(vpc.Description) {

			ret = &vpcSets[i]
			return
		}
	}

	return
}

func (p *Aliyun) FindVSwitch(vpcName, vSwitchName string) (ret *ecs.VSwitchSetType, err error) {

	vpc, err := p.FindVPC(vpcName)

	if err != nil {
		return
	}

	if vpc == nil {
		return
	}

	for _, vswitchId := range vpc.VSwitchIds.VSwitchId {
		var vSwitchSets []ecs.VSwitchSetType

		vSwitchSets, _, err = p.ECSClient().DescribeVSwitches(
			&ecs.DescribeVSwitchesArgs{
				RegionId:  common.Region(p.Region),
				VpcId:     vpc.VpcId,
				VSwitchId: vswitchId,
			},
		)

		for i, vswitch := range vSwitchSets {
			if vswitch.VSwitchName == vSwitchName &&
				p.isSignd(vswitch.Description) {

				ret = &vSwitchSets[i]
				return
			}
		}
	}

	return
}

func (p *Aliyun) CreateVSwitch() (err error) {
	vpcsConf := p.Config.GetConfig("aliyun.ecs.vswitch")

	if vpcsConf.IsEmpty() {
		return
	}

	var args []*ecs.CreateVSwitchArgs

	for _, vSwitchName := range vpcsConf.Keys() {

		vSwitchConf := vpcsConf.GetConfig(vSwitchName)
		vpcName := vSwitchConf.GetString("vpc-name")

		if len(vpcName) == 0 {
			err = fmt.Errorf("vswitch config of %s's vpc-name is not set", vSwitchName)
			return
		}

		vpcId := ""

		var vpc *ecs.VpcSetType
		vpc, err = p.FindVPC(vpcName)

		if err != nil {
			return
		}

		if vpc == nil {
			err = fmt.Errorf("vswitch config of %s's vpc-name: %s is not found at aliyun", vpcName)
			return
		}

		vpcId = vpc.VpcId

		logrus.WithField("CODE", p.Code).
			WithField("VPCID", vpcId).
			WithField("VSWITCH", vSwitchName).Infof("Found vswitch @ %s", vpc.VpcId)

		if len(vpcId) == 0 {
			err = fmt.Errorf("vswitch config of %s's vpc-name: %s is not found at aliyun", vSwitchName, vpcName)
			return
		}

		var vSwitch *ecs.VSwitchSetType
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)
		if err != nil {
			return
		}

		// already created, ignore
		if vSwitch != nil {

			logrus.WithField("CODE", p.Code).
				WithField("VPCID", vpcId).
				WithField("VSWITCH", vSwitchName).WithField("VSWITCH-ID", vSwitch.VSwitchId).Infoln("VSwitch already created")

			continue
		}

		zoneId := vSwitchConf.GetString("zone-id", p.ZoneId)

		if len(zoneId) > 0 {
			if !strings.HasPrefix(zoneId, string(p.Region)) {
				err = fmt.Errorf("zone-id is illegal, zone-id's prefix should be region")
				return
			}
		}

		cidr := vSwitchConf.GetString("cidr-block", "172.16.0.0/24")
		desc := vSwitchConf.GetString("description")

		arg := &ecs.CreateVSwitchArgs{
			VpcId:       vpcId,
			ZoneId:      zoneId,
			CidrBlock:   cidr,
			VSwitchName: vSwitchName,
			Description: p.signWithCode(desc),
		}

		args = append(args, arg)
	}

	for _, arg := range args {

		switchId, e := p.ECSClient().CreateVSwitch(arg)
		if e != nil {
			return e
		}

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VSWITCH-NAME", arg.VSwitchName).
			WithField("ECS-VSWITCH-ID", switchId).
			Infoln("VSwitch created")
	}

	return
}
func (p *Aliyun) DeleteVSwitchArgs() (err error) {
	vSwitchsConf := p.Config.GetConfig("aliyun.ecs.vswitch")

	if vSwitchsConf.IsEmpty() {
		return
	}

	var args []*ecs.DeleteVSwitchArgs

	for _, vSwitchName := range vSwitchsConf.Keys() {

		vSwitchConf := vSwitchsConf.GetConfig(vSwitchName)
		vpcName := vSwitchConf.GetString("vpc-name")

		if len(vpcName) == 0 {
			err = fmt.Errorf("vswitch config of %s's vpc-name is not set", vSwitchName)
			return
		}

		var vSwtich *ecs.VSwitchSetType
		vSwtich, err = p.FindVSwitch(vpcName, vSwitchName)
		if err != nil {
			return
		}

		if vSwtich == nil {
			continue
		}

		arg := &ecs.DeleteVSwitchArgs{
			VSwitchId: vSwtich.VSwitchId,
		}

		args = append(args, arg)
	}

	for _, arg := range args {

		err = p.ECSClient().DeleteVSwitch(arg.VSwitchId)
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VSWITCH-ID", arg.VSwitchId).
			Infoln("VSwitch deleted")
	}

	return
}
