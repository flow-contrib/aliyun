package aliyun

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

	"github.com/denverdino/aliyungo/cs"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/sirupsen/logrus"
)

type DockerClusterVolume struct {
	Cluster *cs.ClusterType
	Volumes []cs.GetVolumeResponse
}

type DockerClusterVolumeCreationArg struct {
	cs.VolumeCreationArgs

	Cluster *cs.ClusterType
}

func (p *Aliyun) ListDockerClusters(clusterNames ...string) (clusters map[string]*cs.ClusterType, err error) {
	if len(clusterNames) == 0 {
		return
	}

	clusterFilter := map[string]bool{}

	for _, n := range clusterNames {
		clusterFilter[n] = true
	}

	clustersResp, err := p.CSClient().DescribeClusters("")
	if err != nil {
		return
	}

	if len(clustersResp) == 0 {
		return
	}

	clusters = make(map[string]*cs.ClusterType)

	for i, cluster := range clustersResp {
		clusters[cluster.Name] = &clustersResp[i]
	}

	return
}

func (p *Aliyun) CreateDockerClusterArgs() (createArgs []*cs.ClusterCreationArgs, err error) {

	csConfig := p.Config.GetConfig("aliyun.cs.swarm")

	if csConfig.IsEmpty() {
		return
	}

	clusters, err := p.ListDockerClusters(csConfig.Keys()...)
	if err != nil {
		return
	}

	var args []*cs.ClusterCreationArgs

	for _, clusterName := range csConfig.Keys() {

		if clusterResp, exist := clusters[clusterName]; exist {
			logrus.WithField("CODE", p.Code).
				WithField("DOCKER-CLUSTER-ID", clusterResp.ClusterID).
				WithField("DOCKER-CLUSTER-NAME", clusterResp.Name).Infoln("Docker cluster already created")

			continue
		}

		clusterConf := csConfig.GetConfig(clusterName)

		vpcName := clusterConf.GetString("vpc-name")
		vSwitchName := clusterConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("docker cluster config of %s's vpc-name or vswitch-name is empty", clusterName)
			return
		}

		var vSwitch *vpc.VSwitch
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)
		if err != nil {
			return
		}

		if vSwitch == nil {
			return
		}

		var ecsPassword, fnName string

		envPrompName := fmt.Sprintf("cs.swarm.%s.password", clusterName)

		ecsPassword, fnName, err = p.tryInvokeEnvFunc(envPrompName, clusterConf.GetString("password"))

		if err != nil {
			return
		}

		if len(fnName) > 0 {
			logrus.WithField("CODE", p.Code).
				WithField("FUNC", fnName).
				WithField("DOCKER-CLUSTER-NAME", clusterName).
				WithField("NAME", envPrompName).
				WithField("VALUE", ecsPassword).Debugln("Environment func invoked")
		}

		arg := &cs.ClusterCreationArgs{
			Name:             clusterName,
			Size:             clusterConf.GetInt64("size", 1),
			NetworkMode:      cs.NetworkModeType(clusterConf.GetString("network-mode", "vpc")),
			SubnetCIDR:       clusterConf.GetString("subnet-cidr"),
			InstanceType:     clusterConf.GetString("instance-type", "ecs.n4.large"),
			VPCID:            vSwitch.VpcId,
			VSwitchID:        vSwitch.VSwitchId,
			Password:         ecsPassword,
			DataDiskSize:     clusterConf.GetInt64("data-disk-size", 100),
			DataDiskCategory: ecs.DiskCategory(clusterConf.GetString("data-disk-category")),
			ECSImageID:       clusterConf.GetString("ecs-image-id"),
			IOOptimized:      ecs.IoOptimized(clusterConf.GetString("io-optimized")),
		}

		args = append(args, arg)
	}

	createArgs = args

	return
}

func (p *Aliyun) DeleteDockerClusterArgs() (clusters []*cs.ClusterType, err error) {

	csConfig := p.Config.GetConfig("aliyun.cs.swarm")

	if csConfig.IsEmpty() {
		return
	}

	existClusters, err := p.ListDockerClusters(csConfig.Keys()...)
	if err != nil {
		return
	}

	var args []*cs.ClusterType

	for _, clusterName := range csConfig.Keys() {

		cluster, exist := existClusters[clusterName]
		if !exist {
			continue
		}

		clusterConf := csConfig.GetConfig(clusterName)

		vpcName := clusterConf.GetString("vpc-name")
		vSwitchName := clusterConf.GetString("vswitch-name")

		if len(vpcName) == 0 || len(vSwitchName) == 0 {
			err = fmt.Errorf("docker cluster config of %s's vpc-name or vswitch-name is empty", clusterName)
			return
		}

		var vSwitch *vpc.VSwitch
		vSwitch, err = p.FindVSwitch(vpcName, vSwitchName)
		if err != nil {
			return
		}

		if vSwitch == nil {
			return
		}

		if cluster.VPCID == vSwitch.VpcId &&
			cluster.VSwitchID == vSwitch.VSwitchId {
			args = append(args, cluster)
		}
	}

	clusters = args

	return
}

type DockerProjectClient struct {
	Cluster *cs.ClusterType
	Client  *cs.ProjectClient
}

func (p *Aliyun) GetDockerClusterProjectClient(clusterNames ...string) (projectClients map[string]*DockerProjectClient, err error) {

	clusters, err := p.ListDockerClusters(clusterNames...)
	if err != nil {
		return
	}

	clients := map[string]*DockerProjectClient{}

	for _, cluster := range clusters {

		var client *cs.ProjectClient
		client, err = p.CSClient().GetProjectClient(cluster.ClusterID)
		if err != nil {
			return
		}

		clients[cluster.Name] = &DockerProjectClient{Client: client, Cluster: cluster}

	}

	projectClients = clients

	return
}

func (p *Aliyun) ListDockerClusterVolumes(clusterNames ...string) (clusterVolumes map[string]DockerClusterVolume, err error) {

	clients, err := p.GetDockerClusterProjectClient(clusterNames...)
	if err != nil {
		return
	}

	var volumes = make(map[string]DockerClusterVolume)

	for clusterName, client := range clients {

		var resp cs.GetVolumesResponse
		resp, err = client.Client.GetVolumes()
		if err != nil {
			return
		}

		volumes[clusterName] = DockerClusterVolume{
			Cluster: client.Cluster,
			Volumes: resp.Volumes,
		}
	}

	clusterVolumes = volumes

	return
}

func (p *Aliyun) CreateDockerClusterVolumnArgs() (createArgs []*DockerClusterVolumeCreationArg, err error) {

	csConfig := p.Config.GetConfig("aliyun.cs.swarm")

	if csConfig.IsEmpty() {
		return
	}

	clustersVols, err := p.ListDockerClusterVolumes(csConfig.Keys()...)

	var args []*DockerClusterVolumeCreationArg

	for _, clusterName := range csConfig.Keys() {

		clusterVols, exist := clustersVols[clusterName]
		if !exist {
			err = fmt.Errorf("cluster %s not exist", clusterName)
			return
		}

		volumesConf := csConfig.GetConfig(clusterName + ".volumes")

	nextVol:
		for _, volumeName := range volumesConf.Keys() {
			if len(clusterVols.Volumes) > 0 {

				for _, v := range clusterVols.Volumes {
					if v.Name == volumeName {

						logrus.WithField("CODE", p.Code).
							WithField("CLUSTER-NAME", clusterName).
							WithField("VOLUME-NAME", volumeName).
							Infoln("Cluster volume already exist")

						continue nextVol
					}
				}
			}

			volumeConf := volumesConf.GetConfig(volumeName)

			driver := volumeConf.GetString("driver")

			arg := cs.VolumeCreationArgs{
				Name:   volumeName,
				Driver: cs.VolumeDriverType(driver),
			}

			if driver == "ossfs" {

				bucket := volumeConf.GetString("options.bucket")
				url := volumeConf.GetString("options.url")
				akId := volumeConf.GetString("options.ak-id")
				akSec := volumeConf.GetString("options.ak-secret")

				if len(bucket) == 0 {
					err = fmt.Errorf("bucket not set, volume: %s", volumeName)
					return
				}

				if len(url) == 0 {
					err = fmt.Errorf("url not set, volume: %s", volumeName)
					return
				}

				if len(akId) == 0 {
					err = fmt.Errorf("oss volume AccessKeyId is empty, volume: %s", volumeName)
					return
				}

				if len(akSec) == 0 {
					err = fmt.Errorf("oss volume AccessKeySecret is empty, volume: %s", volumeName)
					return
				}

				arg.DriverOpts = &cs.OSSOpts{
					Bucket:          bucket,
					AccessKeyId:     akId,
					AccessKeySecret: akSec,
					URL:             url,
					NoStatCache:     volumeConf.GetString("options.no_stat_cache", "false"),
					OtherOpts:       volumeConf.GetString("options.other_opts"),
				}

			} else if driver == "nas" {
				arg.DriverOpts = &cs.NASOpts{
					DiskId: volumeConf.GetString("options.disk-id"),
					Host:   volumeConf.GetString("options.host"),
					Path:   volumeConf.GetString("options.path"),
					Mode:   volumeConf.GetString("options.mode"),
				}
			} else {
				err = fmt.Errorf("unknown driver %s", driver)
				return
			}

			args = append(args,
				&DockerClusterVolumeCreationArg{
					VolumeCreationArgs: arg,
					Cluster:            clusterVols.Cluster,
				})
		}
	}

	createArgs = args

	return
}
