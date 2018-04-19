package aliyun

import (
	"fmt"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/rds"
	"github.com/sirupsen/logrus"
)

func (p *Aliyun) ListRDSInstance() (resp *rds.DescribeDBInstancesResponse, err error) {

	dbInsResp, err := p.RDSClient().DescribeDBInstances(
		&rds.DescribeDBInstancesArgs{
			RegionId: p.Region,
		},
	)

	if err != nil {
		return
	}

	resp = dbInsResp

	return
}

func (p *Aliyun) FindRDSInstance(engine rds.Engine, vpcName, vSwitchName, rdsName, description string) (attrs *rds.DBInstanceAttribute, err error) {

	dbInsResp, err := p.ListRDSInstance()

	if err != nil {
		return
	}

	vswitch, err := p.FindVSwitch(vpcName, vSwitchName)
	if err != nil {
		return
	}

	if vswitch == nil {
		err = fmt.Errorf("vswitch not found: %s in vpc %s is not found", vSwitchName, vpcName)
		return
	}

	for i, v := range dbInsResp.Items.DBInstance {

		if vswitch.VpcId == v.VpcId &&
			vswitch.VSwitchId == v.VSwitchId &&
			v.DBInstanceDescription == p.signWithCode(rdsName+" "+description) &&
			v.Engine == engine {
			attrs = &dbInsResp.Items.DBInstance[i]
			return
		}
	}

	return
}

func (p *Aliyun) CreateRDSInstanceArgs() (createArgs []*rds.CreateDBInstanceArgs, err error) {
	rdssConf := p.Config.GetConfig("aliyun.rds")

	if rdssConf.IsEmpty() {
		return
	}

	var args []*rds.CreateDBInstanceArgs

	for _, rdsName := range rdssConf.Keys() {
		rdsConf := rdssConf.GetConfig(rdsName)

		vpcName := rdsConf.GetString("vpc-name")
		vSwitchName := rdsConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("rds config of %s's vpc-name or vswitch-name is empty", rdsName)
			return
		}

		engine := rds.Engine(rdsConf.GetString("engine", "MySQL"))
		desc := rdsConf.GetString("instance-description")

		var dbIns *rds.DBInstanceAttribute
		dbIns, err = p.FindRDSInstance(rds.Engine(engine), vpcName, vSwitchName, rdsName, desc)
		if err != nil {
			return
		}

		if dbIns != nil {
			logrus.WithField("CODE", p.Code).WithField("RDS", dbIns.DBInstanceId).WithField("DBINSTANCE-NAME", rdsName).Infoln("RDS Instance already created")
			continue
		}

		var vSwitch *ecs.VSwitchSetType
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)

		if err != nil {
			return
		}

		if vSwitch == nil {
			err = fmt.Errorf("rds instance of %s vsiwtch is not found", rdsName)
			return
		}

		arg := &rds.CreateDBInstanceArgs{

			RegionId: common.Region(p.Region),
			ZoneId:   rdsConf.GetString("zone-id", p.ZoneId),

			Engine:        rds.Engine(engine),
			EngineVersion: rdsConf.GetString("engine-version", "5.6"),
			PayType:       rds.DBPayType(rdsConf.GetString("pay-type", "Postpaid")),

			DBInstanceClass:       rdsConf.GetString("instance-class", "rds.mys2.small"),
			DBInstanceStorage:     int(rdsConf.GetInt64("instance-storage", 5)),
			DBInstanceNetType:     common.NetType(rdsConf.GetString("instance-net-type", "Internet")),
			DBInstanceDescription: p.signWithCode(rdsName + " " + desc),
			InstanceNetworkType:   common.NetworkType(rdsConf.GetString("instance-network-type", "VPC")),

			VPCId:     vSwitch.VpcId,
			VSwitchId: vSwitch.VSwitchId,

			UsedTime: rdsConf.GetString("used-time"),
			Period:   common.TimeType(rdsConf.GetString("period")),

			ConnectionMode:   rds.ConnectionMode(rdsConf.GetString("connection-mode", "Performance")),
			SecurityIPList:   rdsConf.GetString("security-ip-list", "172.18.0.0/24"),
			PrivateIpAddress: rdsConf.GetString("private-ip-address", ""),
		}

		args = append(args, arg)
	}

	createArgs = args

	return
}

func (p *Aliyun) DeleteRDSInstanceArgs() (deleteArgs []*rds.DeleteDBInstanceArgs, err error) {
	rdssConf := p.Config.GetConfig("aliyun.rds")

	if rdssConf.IsEmpty() {
		return
	}

	var args []*rds.DeleteDBInstanceArgs

	for _, rdsName := range rdssConf.Keys() {

		rdsConf := rdssConf.GetConfig(rdsName)

		vpcName := rdsConf.GetString("vpc-name")
		vSwitchName := rdsConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("rds config of %s's vpc-name or vswitch-name is empty", rdsName)
			return
		}

		engine := rds.Engine(rdsConf.GetString("engine", "MySQL"))
		desc := rdsConf.GetString("instance-description")

		var dbIns *rds.DBInstanceAttribute
		dbIns, err = p.FindRDSInstance(rds.Engine(engine), vpcName, vSwitchName, rdsName, desc)
		if err != nil {
			if strings.Contains(err.Error(), "vswitch not found") {
				err = nil
			}
			return
		}

		if dbIns == nil {
			continue
		}

		var vSwitch *ecs.VSwitchSetType
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)

		if err != nil {
			return
		}

		if vSwitch == nil {
			err = fmt.Errorf("rds instance of %s vsiwtch is not found", rdsName)
			return
		}

		arg := &rds.DeleteDBInstanceArgs{
			DBInstanceId: dbIns.DBInstanceId,
		}

		args = append(args, arg)

	}

	deleteArgs = args

	return
}
