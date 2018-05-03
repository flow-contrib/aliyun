package aliyun

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"github.com/sirupsen/logrus"
)

func (p *Aliyun) describeVPCs(vpcIds ...string) (resp *vpc.DescribeVpcsResponse, err error) {
	describeReq := vpc.CreateDescribeVpcsRequest()

	describeReq.RegionId = p.Region
	describeReq.VpcId = strings.Join(vpcIds, ",")

	resp, err = p.VPCClient().DescribeVpcs(describeReq)

	if err != nil {
		return
	}

	return
}

func (p *Aliyun) describeVSwitches(vpcId string, switchIds ...string) (resp *vpc.DescribeVSwitchesResponse, err error) {

	describeReq := vpc.CreateDescribeVSwitchesRequest()

	describeReq.RegionId = p.Region
	describeReq.VpcId = vpcId

	describeReq.VSwitchId = strings.Join(switchIds, ",")

	resp, err = p.VPCClient().DescribeVSwitches(describeReq)
	if err != nil {
		return
	}

	return
}

func (p *Aliyun) CreateVPCs() (err error) {

	vpcsConf := p.Config.GetConfig("aliyun.vpc.vpc")

	if vpcsConf.IsEmpty() {
		return
	}

	vpcDescribeResp, err := p.describeVPCs()

	if err != nil {
		return
	}

	var createReqList []*vpc.CreateVpcRequest

	for _, vpcName := range vpcsConf.Keys() {

		vpcConf := vpcsConf.GetConfig(vpcName)

		vpcId := vpcConf.GetString("id")

		if len(vpcId) > 0 {
			continue
		}

		created := false

		desc := vpcConf.GetString("description")

		req := vpc.CreateCreateVpcRequest()

		req.RegionId = p.Region
		req.VpcName = vpcName
		req.CidrBlock = vpcConf.GetString("cidr-block", "172.16.0.0/16")
		req.Description = p.signWithCode(desc)

		for _, s := range vpcDescribeResp.Vpcs.Vpc {
			if s.VpcName == req.VpcName &&
				s.CidrBlock == req.CidrBlock &&
				s.RegionId == req.RegionId &&
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

		createReqList = append(createReqList, req)
	}

	for _, arg := range createReqList {

		resp, e := p.VPCClient().CreateVpc(arg)
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

func (p *Aliyun) DeleteVPC() (err error) {
	vpcsConf := p.Config.GetConfig("aliyun.vpc.vpc")

	if vpcsConf.IsEmpty() {
		return
	}

	vpcDescribeResp, err := p.describeVPCs()

	if err != nil {
		return
	}

	var deleteReqList []*vpc.DeleteVpcRequest

	for _, vpcName := range vpcsConf.Keys() {

		vpcConf := vpcsConf.GetConfig(vpcName)

		vpcId := vpcConf.GetString("id")

		if len(vpcId) == 0 {
			for _, s := range vpcDescribeResp.Vpcs.Vpc {
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
			req := vpc.CreateDeleteVpcRequest()

			req.VpcId = vpcId

			deleteReqList = append(deleteReqList, req)
		}
	}

	for _, req := range deleteReqList {

		_, err = p.VPCClient().DeleteVpc(req)
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VPC-ID", req.VpcId).
			Infoln("VPC deleted")
	}

	return
}

func (p *Aliyun) WaitForAllVpcRunning(timeout int) (err error) {

	vpcsConf := p.Config.GetConfig("aliyun.vpc.vpc")

	if vpcsConf.IsEmpty() {
		return
	}

	vpcDescribeResp, err := p.describeVPCs()

	if err != nil {
		return
	}

	mapVpcs := map[string]string{}

	for _, s := range vpcDescribeResp.Vpcs.Vpc {
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
			p.WaitForVpcAvailable(vpcId, timeout)
		}(vpcIds[i])
	}

	logrus.WithField("CODE", p.Code).Infoln("Wait for all VPC available")

	wg.Wait()

	return
}

func (p *Aliyun) FindVPC(vpcName string) (ret *vpc.Vpc, err error) {
	vpcDescribeResp, err := p.describeVPCs()

	for i, vpc := range vpcDescribeResp.Vpcs.Vpc {
		if vpcName == vpc.VpcName &&
			p.isSignd(vpc.Description) {

			ret = &vpcDescribeResp.Vpcs.Vpc[i]
			return
		}
	}

	return
}

func (p *Aliyun) FindVSwitch(vpcName, vSwitchName string) (ret *vpc.VSwitch, err error) {

	vpcInst, err := p.FindVPC(vpcName)

	if err != nil {
		return
	}

	if vpcInst == nil {
		return
	}

	vswitchesDescribe, err := p.describeVSwitches(vpcInst.VpcId)
	if err != nil {
		return
	}

	for i, vswitch := range vswitchesDescribe.VSwitches.VSwitch {
		if vswitch.VSwitchName == vSwitchName &&
			p.isSignd(vswitch.Description) {

			ret = &vswitchesDescribe.VSwitches.VSwitch[i]
			return
		}
	}

	return
}

func (p *Aliyun) CreateVSwitch() (err error) {
	vSwitchesConf := p.Config.GetConfig("aliyun.vpc.vswitch")

	if vSwitchesConf.IsEmpty() {
		return
	}

	var createReqList []*vpc.CreateVSwitchRequest

	for _, vSwitchName := range vSwitchesConf.Keys() {

		vSwitchConf := vSwitchesConf.GetConfig(vSwitchName)
		vpcName := vSwitchConf.GetString("vpc-name")

		if len(vpcName) == 0 {
			err = fmt.Errorf("vswitch config of %s's vpc-name is not set", vSwitchName)
			return
		}

		vpcId := ""

		var vpcInst *vpc.Vpc
		vpcInst, err = p.FindVPC(vpcName)

		if err != nil {
			return
		}

		if vpcInst == nil {
			err = fmt.Errorf("vswitch config of %s's vpc-name: %s is not found at aliyun", vSwitchName, vpcName)
			return
		}

		vpcId = vpcInst.VpcId

		logrus.WithField("CODE", p.Code).
			WithField("VPCID", vpcId).
			WithField("VSWITCH", vSwitchName).Infof("Found vswitch @ %s", vpcInst.VpcId)

		if len(vpcId) == 0 {
			err = fmt.Errorf("vswitch config of %s's vpc-name: %s is not found at aliyun", vSwitchName, vpcName)
			return
		}

		var vSwitch *vpc.VSwitch
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

		zoneId := vSwitchConf.GetString("zone-id")

		if len(zoneId) == 0 {
			err = fmt.Errorf("the config of zone-id is empty in vswitch of %s", vSwitchName)
			return
		}

		if len(zoneId) > 0 {
			if !strings.HasPrefix(zoneId, string(p.Region)) {
				err = fmt.Errorf("zone-id is illegal, zone-id's prefix should be region")
				return
			}
		}

		cidr := vSwitchConf.GetString("cidr-block", "172.16.0.0/24")
		desc := vSwitchConf.GetString("description")

		req := vpc.CreateCreateVSwitchRequest()

		req.VpcId = vpcId
		req.ZoneId = zoneId
		req.CidrBlock = cidr
		req.VSwitchName = vSwitchName
		req.Description = p.signWithCode(desc)

		createReqList = append(createReqList, req)
	}

	for _, req := range createReqList {

		var resp *vpc.CreateVSwitchResponse
		resp, err = p.VPCClient().CreateVSwitch(req)
		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VSWITCH-NAME", req.VSwitchName).
			WithField("ECS-VSWITCH-ID", resp.VSwitchId).
			Infoln("VSwitch created")
	}

	return
}
func (p *Aliyun) DeleteVSwitch() (err error) {
	vSwitchsConf := p.Config.GetConfig("aliyun.vpc.vswitch")

	if vSwitchsConf.IsEmpty() {
		return
	}

	var deleteReqList []*vpc.DeleteVSwitchRequest

	for _, vSwitchName := range vSwitchsConf.Keys() {

		vSwitchConf := vSwitchsConf.GetConfig(vSwitchName)
		vpcName := vSwitchConf.GetString("vpc-name")

		if len(vpcName) == 0 {
			err = fmt.Errorf("vswitch config of %s's vpc-name is not set", vSwitchName)
			return
		}

		var vSwtich *vpc.VSwitch
		vSwtich, err = p.FindVSwitch(vpcName, vSwitchName)
		if err != nil {
			return
		}

		if vSwtich == nil {
			continue
		}

		req := vpc.CreateDeleteVSwitchRequest()

		req.RegionId = p.Region
		req.VSwitchId = vSwtich.VSwitchId

		deleteReqList = append(deleteReqList, req)
	}

	for _, req := range deleteReqList {

		_, err = p.VPCClient().DeleteVSwitch(req)
		if err != nil {
			return
		}

		time.Sleep(time.Second * 2) // waitting for router list to be deleted, or else, it will error while delete vpc

		logrus.WithField("CODE", p.Code).
			WithField("ECS-VSWITCH-ID", req.VSwitchId).
			Infoln("VSwitch deleted")
	}

	return
}

func (p *Aliyun) WaitForVpcAvailable(vpcId string, timeout int) (err error) {
	if timeout <= 0 {
		timeout = 60
	}

	for {
		var resp *vpc.DescribeVpcsResponse
		resp, err = p.describeVPCs(vpcId)
		if err != nil {
			return err
		}

		if len(resp.Vpcs.Vpc) > 0 && resp.Vpcs.Vpc[0].Status == "Available" {
			break
		}

		timeout = timeout - 5

		if timeout <= 0 {
			err = fmt.Errorf("wait for vpc '%s' available timeout", vpcId)
			return
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}

func (p *Aliyun) WaitForVSwitchAvailable(vpcId string, vswitchId string, timeout int) (err error) {
	if timeout <= 0 {
		timeout = 60
	}

	for {

		var resp *vpc.DescribeVSwitchesResponse
		resp, err = p.describeVSwitches(vpcId, vswitchId)
		if err != nil {
			return
		}

		if len(resp.VSwitches.VSwitch) == 0 {
			err = fmt.Errorf("vswitch %s not found", vswitchId)
			return
		}

		if resp.VSwitches.VSwitch[0].Status == "Available" {
			break
		}

		timeout = timeout - 5
		if timeout <= 0 {
			err = fmt.Errorf("wait for vsiwtch '%s' available timeout", vswitchId)
			return
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}
