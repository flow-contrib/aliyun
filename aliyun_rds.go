package aliyun

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"github.com/sirupsen/logrus"
)

func (p *Aliyun) listRDSInstance(tags map[string]string) (resp *rds.DescribeDBInstancesResponse, err error) {

	describReq := rds.CreateDescribeDBInstancesRequest()

	describReq.RegionId = string(p.Region)

	if tags == nil {
		tags = make(map[string]string)
	}

	tags["creator"] = "go-flow"
	tags["code"] = p.Code

	var tagData []byte
	tagData, err = json.Marshal(tags)
	if err != nil {
		return
	}
	describReq.Tags = string(tagData)

	dbInsResp, err := p.RDSClient().DescribeDBInstances(describReq)

	if err != nil {
		return
	}

	resp = dbInsResp

	return
}

func (p *Aliyun) DescribeRDSInstancesAttr() (attr []rds.DBInstanceAttribute, err error) {

	resp, err := p.listRDSInstance(nil)
	if err != nil {
		return
	}

	var ids []string

	for _, inst := range resp.Items.DBInstance {
		ids = append(ids, inst.DBInstanceId)
	}

	strIds := strings.Join(ids, ",")

	descReq := rds.CreateDescribeDBInstanceAttributeRequest()

	descReq.DBInstanceId = strIds

	attrResp, err := p.RDSClient().DescribeDBInstanceAttribute(descReq)

	if err != nil {
		return
	}

	return attrResp.Items.DBInstanceAttribute, nil
}

func (p *Aliyun) FindRDSInstance(engine, vpcName, vSwitchName, rdsName string) (attrs *rds.DBInstance, err error) {

	dbInsResp, err := p.listRDSInstance(map[string]string{"name": rdsName})

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
			v.Engine == engine {
			attrs = &dbInsResp.Items.DBInstance[i]
			return
		}
	}

	return
}

type CreateRDSInstancesArgs struct {
	*rds.CreateDBInstanceRequest
	Name string
}

func (p *Aliyun) CreateRDSInstances() (createResps []*rds.CreateDBInstanceResponse, err error) {
	rdssConf := p.Config.GetConfig("aliyun.rds")

	if rdssConf.IsEmpty() {
		return
	}

	var args []*CreateRDSInstancesArgs

	for _, rdsName := range rdssConf.Keys() {
		rdsConf := rdssConf.GetConfig(rdsName)

		vpcName := rdsConf.GetString("vpc-name")
		vSwitchName := rdsConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("rds config of %s's vpc-name or vswitch-name is empty", rdsName)
			return
		}

		engine := rdsConf.GetString("engine", "MySQL")

		var dbIns *rds.DBInstance
		dbIns, err = p.FindRDSInstance(engine, vpcName, vSwitchName, rdsName)
		if err != nil {
			return
		}

		if dbIns != nil {
			logrus.WithField("CODE", p.Code).WithField("RDS", dbIns.DBInstanceId).WithField("DBINSTANCE-NAME", rdsName).Infoln("RDS Instance already created")
			continue
		}

		var vSwitch *vpc.VSwitch
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)

		if err != nil {
			return
		}

		if vSwitch == nil {
			err = fmt.Errorf("rds instance of %s vsiwtch is not found", rdsName)
			return
		}

		arg := rds.CreateCreateDBInstanceRequest()

		arg.RegionId = string(p.Region)
		arg.ZoneId = rdsConf.GetString("zone-id", p.ZoneId)

		arg.Engine = engine
		arg.EngineVersion = rdsConf.GetString("engine-version", "5.6")
		arg.PayType = rdsConf.GetString("pay-type", "Postpaid")

		arg.DBInstanceClass = rdsConf.GetString("instance-class", "rds.mys2.small")
		arg.DBInstanceStorage = requests.Integer(rdsConf.GetString("instance-storage", "5"))
		arg.DBInstanceNetType = rdsConf.GetString("instance-net-type", "Internet")
		arg.DBInstanceDescription = rdsName
		arg.InstanceNetworkType = rdsConf.GetString("instance-network-type", "VPC")

		arg.VPCId = vSwitch.VpcId
		arg.VSwitchId = vSwitch.VSwitchId

		arg.UsedTime = rdsConf.GetString("used-time")
		arg.Period = rdsConf.GetString("period")

		arg.ConnectionMode = rdsConf.GetString("connection-mode", "Performance")
		arg.SecurityIPList = rdsConf.GetString("security-ip-list", "172.18.0.0/24")
		arg.PrivateIpAddress = rdsConf.GetString("private-ip-address", "")

		args = append(args, &CreateRDSInstancesArgs{
			CreateDBInstanceRequest: arg,
			Name: rdsName,
		})
	}

	var ret []*rds.CreateDBInstanceResponse

	for _, arg := range args {

		var resp *rds.CreateDBInstanceResponse

		resp, err = p.RDSClient().CreateDBInstance(arg.CreateDBInstanceRequest)

		if err != nil {
			return
		}

		addTagsReq := rds.CreateAddTagsToResourceRequest()
		addTagsReq.RegionId = string(p.Region)

		addTagsReq.DBInstanceId = resp.DBInstanceId

		addTagsReq.Tag1Key = "code"
		addTagsReq.Tag1Value = p.Code

		addTagsReq.Tag2Key = "creator"
		addTagsReq.Tag2Value = "go-flow"

		addTagsReq.Tag3Key = "name"
		addTagsReq.Tag3Value = arg.Name

		var oRdsClient *rds.Client
		oRdsClient, err = rds.NewClientWithAccessKey(string(p.Region), p.AccessKeyId, p.AccessKeySecret)
		if err != nil {
			return
		}

		_, err = oRdsClient.AddTagsToResource(addTagsReq)

		if err != nil {
			return
		}

		ret = append(ret, resp)

		logrus.WithField("CODE", p.Code).
			WithField("RDS-DBINSTANCE-ID", resp.DBInstanceId).
			WithField("RDS-ENGINE", string(arg.Engine)+" "+arg.EngineVersion).
			WithField("RDS-CONN-STR", resp.ConnectionString).
			WithField("RDS-REGION", arg.RegionId).
			WithField("RDS-VSWITCH-ID", arg.VSwitchId).
			Infoln("Db instance created")
	}

	createResps = ret

	return
}

func (p *Aliyun) CreateRDSDbAccount() (err error) {

	rdssConf := p.Config.GetConfig("aliyun.rds")

	if rdssConf.IsEmpty() {
		return
	}

	for _, rdsName := range rdssConf.Keys() {
		rdsConf := rdssConf.GetConfig(rdsName)

		vpcName := rdsConf.GetString("vpc-name")
		vSwitchName := rdsConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("rds config of %s's vpc-name or vswitch-name is empty", rdsName)
			return
		}

		engine := rdsConf.GetString("engine", "MySQL")

		var dbIns *rds.DBInstance
		dbIns, err = p.FindRDSInstance(engine, vpcName, vSwitchName, rdsName)
		if err != nil {
			return
		}

		if dbIns == nil {
			logrus.WithField("CODE", p.Code).WithField("DBINSTANCE-NAME", rdsName).Infoln("RDS Instance not exist")
			continue
		}

		accountsConf := rdsConf.GetConfig("accounts")

		if accountsConf.IsEmpty() {
			return
		}

		var accountsResp *rds.DescribeAccountsResponse

		describeAccReq := rds.CreateDescribeAccountsRequest()
		describeAccReq.DBInstanceId = dbIns.DBInstanceId

		accountsResp, err = p.RDSClient().DescribeAccounts(describeAccReq)

		if err != nil {
			return
		}

		existAccounts := map[string]rds.DBInstanceAccount{}

		for _, account := range accountsResp.Accounts.DBInstanceAccount {
			existAccounts[account.AccountName] = account
		}

		if engine == "PostgreSQL" {
			if len(accountsConf.Keys()) > 1 {
				err = fmt.Errorf("the db of [%s]'s instance type is PostgreSQL, it only can create one admin account", rdsName)
				return
			}
		}

		for _, accountName := range accountsConf.Keys() {

			accountConf := accountsConf.GetConfig(accountName)

			accountName = accountsConf.GetString("account", accountName)

			_, exist := existAccounts[accountName]

			if !exist {

				createAccountArgs := rds.CreateCreateAccountRequest()

				createAccountArgs.DBInstanceId = dbIns.DBInstanceId
				createAccountArgs.AccountName = accountName
				createAccountArgs.AccountPassword = accountConf.GetString("password")
				createAccountArgs.AccountDescription = accountConf.GetString("description")
				createAccountArgs.AccountType = accountConf.GetString("type", "Normal")

				_, err = p.RDSClient().CreateAccount(createAccountArgs)

				if err != nil {
					return
				}
			}

			privilegeConf := accountConf.GetConfig("databases")

			if engine == "PostgreSQL" {
				if len(privilegeConf.Keys()) > 0 {
					err = fmt.Errorf("the db of [%s]'s instance type is PostgreSQL,it could not grant privilege", rdsName)
					return
				}
			}

			if privilegeConf.IsEmpty() {
				continue
			}

			for _, dbName := range privilegeConf.Keys() {

				grantArgs := rds.CreateGrantAccountPrivilegeRequest()

				grantArgs.DBInstanceId = dbIns.DBInstanceId
				grantArgs.AccountName = accountName
				grantArgs.DBName = dbName
				grantArgs.AccountPrivilege = privilegeConf.GetString(dbName+".privilege", "ReadWrite")

				_, err = p.RDSClient().GrantAccountPrivilege(grantArgs)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (p *Aliyun) DeleteRDSInstances() (err error) {
	rdssConf := p.Config.GetConfig("aliyun.rds")

	if rdssConf.IsEmpty() {
		return
	}

	var args []*rds.DeleteDBInstanceRequest

	for _, rdsName := range rdssConf.Keys() {

		rdsConf := rdssConf.GetConfig(rdsName)

		vpcName := rdsConf.GetString("vpc-name")
		vSwitchName := rdsConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("rds config of %s's vpc-name or vswitch-name is empty", rdsName)
			return
		}

		engine := rdsConf.GetString("engine", "MySQL")

		var dbIns *rds.DBInstance
		dbIns, err = p.FindRDSInstance(engine, vpcName, vSwitchName, rdsName)
		if err != nil {
			if strings.Contains(err.Error(), "vswitch not found") {
				err = nil
			}
			return
		}

		if dbIns == nil {
			continue
		}

		var vSwitch *vpc.VSwitch
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)

		if err != nil {
			return
		}

		if vSwitch == nil {
			err = fmt.Errorf("rds instance of %s vsiwtch is not found", rdsName)
			return
		}

		arg := rds.CreateDeleteDBInstanceRequest()

		arg.DBInstanceId = dbIns.DBInstanceId

		args = append(args, arg)

	}

	for _, arg := range args {

		_, err = p.RDSClient().DeleteDBInstance(arg)

		if err != nil {
			return
		}

		logrus.WithField("CODE", p.Code).WithField("RDS-DBINSTANCE-ID", arg.DBInstanceId).Infoln("Db instance deleted")
	}

	return
}

// Default timeout value for WaitForInstance method
const InstanceDefaultTimeout = 120
const DefaultWaitForInterval = 10

// WaitForInstance waits for instance to given status
func (p *Aliyun) WaitForDBInstance(instanceId string, status string, timeout int) error {
	if timeout <= 0 {
		timeout = InstanceDefaultTimeout
	}
	for {

		args := rds.CreateDescribeDBInstancesRequest()

		args.DBInstanceId = instanceId

		resp, err := p.RDSClient().DescribeDBInstances(args)

		if err != nil {
			return nil
		}

		if !resp.IsSuccess() {
			err = fmt.Errorf("describe db instances failure")
			return err
		}

		timeout = timeout - DefaultWaitForInterval
		time.Sleep(DefaultWaitForInterval * time.Second)

		if len(resp.Items.DBInstance) < 1 {
			continue
		}

		instance := resp.Items.DBInstance[0]
		if instance.DBInstanceStatus == status {
			break
		}
	}
	return nil
}
