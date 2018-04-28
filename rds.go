package aliyun

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("devops.aliyun.rds.db.create", CreateRDSInstance)
	flow.RegisterHandler("devops.aliyun.rds.db.attribute.describe", DescribeRDSInstanceAttr)
	flow.RegisterHandler("devops.aliyun.rds.db.netinfo.describe", DescribeRDSInstanceNetInfo)
	flow.RegisterHandler("devops.aliyun.rds.db.delete", DeleteRDSInstance)
	flow.RegisterHandler("devops.aliyun.rds.db.running.wait", WaitForAllRDSRunning)
	flow.RegisterHandler("devops.aliyun.rds.db.account.create", CreateRDSDbAccounts)
	flow.RegisterHandler("devops.aliyun.rds.db.conn.public.alloc", AllocateInstancePublicConnection)
	flow.RegisterHandler("devops.aliyun.rds.db.conn.public.release", ReleaseInstancePublicConnection)
}

func CreateRDSInstance(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	_, err = aliyun.CreateRDSInstances()

	if err != nil {
		return
	}

	return
}

func CreateRDSDbAccounts(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.CreateRDSDbAccount()

	return
}

func DeleteRDSInstance(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.DeleteRDSInstances()

	return
}

func AllocateInstancePublicConnection(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.AllocateInstancePublicConnection()

	return
}

func ReleaseInstancePublicConnection(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	err = aliyun.ReleaseInstancePublicConnection()

	return
}

func DescribeRDSInstanceAttr(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	insts, err := aliyun.DescribeRDSInstancesAttr()
	if err != nil {
		return
	}

	if len(insts) == 0 {
		return
	}

	ouputData, err := json.Marshal(insts)
	if err != nil {
		return
	}

	var tags []string

	for _, inst := range insts {
		tags = append(tags, inst.Name)
		setENV(fmt.Sprintf("rds_db_%s_host", inst.Name), inst.ConnectionString)
		setENV(fmt.Sprintf("rds_db_%s_port", inst.Name), inst.Port)
	}

	tags = append(tags, "aliyun", "rds", aliyun.Code)

	flow.AppendOutput(ctx, flow.NameValue{Name: "ALIYUN_RDS_INSTANCES", Value: ouputData, Tags: tags})

	return
}

func DescribeRDSInstanceNetInfo(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	insts, err := aliyun.DescribeDBInstanceNetInfo()
	if err != nil {
		return
	}

	if len(insts) == 0 {
		return
	}

	ouputData, err := json.Marshal(insts)
	if err != nil {
		return
	}

	var tags []string

	for _, inst := range insts {
		tags = append(tags, inst.InstanceName)

		for _, netinfo := range inst.NetInfo {
			setENV(fmt.Sprintf("rds_db_%s_%s_host", inst.InstanceName, netinfo.IPType), netinfo.ConnectionString)
			setENV(fmt.Sprintf("rds_db_%s_%s_port", inst.InstanceName, netinfo.IPType), netinfo.Port)
		}
	}

	tags = append(tags, "aliyun", "rds", aliyun.Code)

	flow.AppendOutput(ctx, flow.NameValue{Name: "ALIYUN_RDS_INSTANCES_NET_INFO", Value: ouputData, Tags: tags})

	return
}

func WaitForAllRDSRunning(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	inst, err := aliyun.listRDSInstance(nil)

	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}

	for _, v := range inst.Items.DBInstance {
		wg.Add(1)
		go func(instId, name string) {
			defer wg.Done()
			logrus.WithField("CODE", aliyun.Code).WithField("RDS-DBINSTANCE-ID", instId).WithField("RDS-DBINSTANCE-NAME", name).Infoln("Waiting db instance")
			aliyun.WaitForDBInstance(instId, "Running", 60*20)
		}(v.DBInstanceId, v.DBInstanceDescription)
	}

	wg.Wait()

	return
}
