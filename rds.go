package aliyun

import (
	"sync"

	"github.com/denverdino/aliyungo/rds"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("devops.aliyun.rds.instance.create", CreateRDSInstance)
	flow.RegisterHandler("devops.aliyun.rds.instance.delete", DeleteRDSInstance)
	flow.RegisterHandler("devops.aliyun.rds.instance.running.wait", WaitForAllRDSRunning)
}

func CreateRDSInstance(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateRDSInstanceArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		var resp rds.CreateDBInstanceResponse
		resp, err = aliyun.RDSClient().CreateDBInstance(arg)

		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("RDS-DBINSTANCE-ID", resp.DBInstanceId).
			WithField("RDS-ENGINE", string(arg.Engine)+" "+arg.EngineVersion).
			WithField("RDS-CONN-STR", resp.ConnectionString).
			WithField("RDS-REGION", arg.RegionId).
			WithField("RDS-VSWITCH-ID", arg.VSwitchId).
			Infoln("Db instance created")
	}

	return
}

func DeleteRDSInstance(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.DeleteRDSInstanceArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		err = aliyun.RDSClient().DeleteInstance(arg.DBInstanceId)

		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).WithField("RDS-DBINSTANCE-ID", arg.DBInstanceId).Infoln("Db instance deleted")
	}

	return
}

func WaitForAllRDSRunning(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	inst, err := aliyun.ListRDSInstance()

	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}

	for _, v := range inst.Items.DBInstance {
		wg.Add(1)
		go func(instId, name string) {
			defer wg.Done()
			logrus.WithField("CODE", aliyun.Code).WithField("RDS-DBINSTANCE-ID", instId).WithField("RDS-DBINSTANCE-NAME", name).Infoln("Waiting db instance")
			aliyun.RDSClient().WaitForInstance(instId, rds.Running, 60*20)
		}(v.DBInstanceId, v.DBInstanceDescription)
	}

	wg.Wait()

	return
}
