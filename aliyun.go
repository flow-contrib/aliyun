package aliyun

import (
	"fmt"
	"os"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/cs"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/slb"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/gogap/config"
	"github.com/gogap/context"

	alierrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
)

type Aliyun struct {
	Config config.Configuration

	AccessKeyId     string
	AccessKeySecret string
	Region          string
	Code            string
	// ZoneId          string

	vpcClient *vpc.Client
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

	envRegion := os.Getenv("ENV_ALIYUN_REGION")
	envAkId := os.Getenv("ENV_ALIYUN_ACCESS_KEY_ID")
	envAkSecret := os.Getenv("ENV_ALIYUN_ACCESS_KEY_SECRET")

	akId := conf.GetString("aliyun.access-key-id", envAkId)
	akSecret := conf.GetString("aliyun.access-key-secret", envAkSecret)
	region := conf.GetString("aliyun.region", envRegion)

	if len(akId) == 0 || len(akSecret) == 0 {
		panic(fmt.Errorf("please set aliyun AccessKeyId into config or set env to ${ENV_ALIYUN_ACCESS_KEY_ID} and accessKeySecret into ${ENV_ALIYUN_ACCESS_KEY_SECRET}"))
	}

	if len(region) == 0 {
		panic("region is empty, please set it to config file or env ${ENV_ALIYUN_REGION}")
	}

	ali := &Aliyun{
		Config: conf,

		AccessKeyId:     akId,
		AccessKeySecret: akSecret,
		Region:          region,
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
		endpoint := fmt.Sprintf("oss-%s.aliyuncs.com", p.Region)
		var err error
		p.ossClient, err = oss.New(endpoint, p.AccessKeyId, p.AccessKeySecret)
		if err != nil {
			panic(err)
		}
	}

	return p.ossClient
}

func (p *Aliyun) RDSClient() *rds.Client {
	if p.rdsClient == nil {
		var err error
		p.rdsClient, err = rds.NewClientWithAccessKey(p.Region, p.AccessKeyId, p.AccessKeySecret)
		if err != nil {
			panic(err)
		}
	}

	return p.rdsClient
}

func (p *Aliyun) VPCClient() *vpc.Client {
	if p.vpcClient == nil {
		var err error
		p.vpcClient, err = vpc.NewClientWithAccessKey(p.Region, p.AccessKeyId, p.AccessKeySecret)
		if err != nil {
			panic(err)
		}
	}

	return p.vpcClient
}

func (p *Aliyun) CSClient() *cs.Client {
	if p.csClient == nil {
		p.csClient = cs.NewClient(p.AccessKeyId, p.AccessKeySecret)
	}

	return p.csClient
}

func (p *Aliyun) SLBClient() *slb.Client {
	if p.slbClient == nil {
		p.slbClient = slb.NewSLBClient(p.AccessKeyId, p.AccessKeySecret, common.Region(p.Region))
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

	switch v := err.(type) {
	case alierrors.Error:
		{
			if v.ErrorCode() == code {
				return true
			}
		}
	case *common.Error:
		{
			if v.Code == code {
				return true
			}
		}
	}

	return false
}

func setENV(key, value string) error {
	key = strings.Replace(key, "-", "_", -1)
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, " ", "_", -1)
	key = strings.Replace(key, ":", "_", -1)
	key = strings.ToUpper(key)
	key = "ENV_ALIYUN_" + key

	return os.Setenv(key, value)
}
