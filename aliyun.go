package aliyun

import (
	"fmt"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/cs"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/oss"
	"github.com/denverdino/aliyungo/rds"
	"github.com/denverdino/aliyungo/slb"

	"github.com/gogap/config"
	"github.com/gogap/context"
)

type Aliyun struct {
	Config config.Configuration

	AccessKeyId     string
	AccessKeySecret string
	Region          common.Region
	Code            string
	ZoneId          string

	ecsClient *ecs.Client
	ossClient *oss.Client
	rdsClient *rds.Client
	csClient  *cs.Client
	slbClient *slb.Client
}

func NewAliyun(ctx context.Context, conf config.Configuration) *Aliyun {

	code := conf.GetString("aliyun.code")

	if len(code) == 0 {
		panic(fmt.Errorf("the config of aliyun.code is empty"))
	}

	region := conf.GetString("aliyun.region", "cn-beijing")
	zoneId := conf.GetString("aliyun.zone-id")
	akId := conf.GetString("aliyun.access-key-id")
	akSecret := conf.GetString("aliyun.access-key-secret")

	ali := &Aliyun{
		Config: conf,

		AccessKeyId:     akId,
		AccessKeySecret: akSecret,
		Region:          common.Region(region),
		ZoneId:          zoneId,
		Code:            code,

		ecsClient: ecs.NewClient(akId, akSecret),
		ossClient: oss.NewOSSClient(oss.Region("oss-"+region), false, akId, akSecret, true),
		rdsClient: rds.NewRDSClient(akId, akSecret, common.Region(region)),
		csClient:  cs.NewClient(akId, akSecret),
		slbClient: slb.NewSLBClient(akId, akSecret, common.Region(region)),
	}

	return ali
}

func (p *Aliyun) ECSClient() *ecs.Client {
	return p.ecsClient
}

func (p *Aliyun) OSSClient() *oss.Client {
	return p.ossClient
}

func (p *Aliyun) RDSClient() *rds.Client {
	return p.rdsClient
}

func (p *Aliyun) CSClient() *cs.Client {
	return p.csClient
}

func (p *Aliyun) SLBClient() *slb.Client {
	return p.slbClient
}

func (p *Aliyun) signWithCode(str string) string {
	return fmt.Sprintf("%s [%s]", str, p.Code)
}

func (p *Aliyun) isSignd(str string) bool {
	return strings.Contains(str, fmt.Sprintf("[%s]", p.Code))
}

func IsAliErrCode(err error, code string) bool {
	if aliyunErr, ok := err.(*common.Error); ok {
		if aliyunErr.Code == code {
			return true
		}
	}

	return false
}
