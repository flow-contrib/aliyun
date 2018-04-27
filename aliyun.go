package aliyun

import (
	"fmt"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/cs"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/oss"
	"github.com/denverdino/aliyungo/slb"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

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

	code, _ := ctx.Value("code").(string)

	if len(code) == 0 {
		code = conf.GetString("code")
	}

	if len(code) == 0 {
		panic(fmt.Errorf("the context of code is empty"))
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
	}

	return ali
}

func (p *Aliyun) ECSClient() *ecs.Client {
	if p.ecsClient == nil {
		p.ecsClient = ecs.NewClient(p.AccessKeyId, p.AccessKeySecret)
	}

	return p.ecsClient
}

func (p *Aliyun) OSSClient() *oss.Client {
	if p.ossClient == nil {
		p.ossClient = oss.NewOSSClient(oss.Region("oss-"+string(p.Region)), false, p.AccessKeyId, p.AccessKeySecret, true)
	}

	return p.ossClient
}

func (p *Aliyun) RDSClient() *rds.Client {
	if p.rdsClient == nil {
		var err error
		p.rdsClient, err = rds.NewClientWithAccessKey(string(p.Region), p.AccessKeyId, p.AccessKeySecret)
		if err != nil {
			panic(err)
		}
	}

	return p.rdsClient
}

func (p *Aliyun) CSClient() *cs.Client {
	if p.csClient == nil {
		p.csClient = cs.NewClient(p.AccessKeyId, p.AccessKeySecret)
	}

	return p.csClient
}

func (p *Aliyun) SLBClient() *slb.Client {
	if p.slbClient == nil {
		p.slbClient = slb.NewSLBClient(p.AccessKeyId, p.AccessKeySecret, p.Region)
	}
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
