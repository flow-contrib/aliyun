package aliyun

import (
	"sync"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/cs"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("devops.aliyun.cs.cluster.create", CreateDockerCluster)
	flow.RegisterHandler("devops.aliyun.cs.cluster.delete", DeleteDockerCluster)
	flow.RegisterHandler("devops.aliyun.cs.cluster.volume.create", CreateDockerClusterVolume)
	flow.RegisterHandler("devops.aliyun.cs.cluster.running.wait", WaitForAllClusterRunning)
	flow.RegisterHandler("devops.aliyun.cs.cluster.deleted.wait", WaitForAllClusterDeleted)
	flow.RegisterHandler("devops.aliyun.cs.cluster.project.create", CreateDockerClusterProject)
	flow.RegisterHandler("devops.aliyun.cs.cluster.project.delete", DeleteDockerClusterProject)
}

func CreateDockerCluster(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateDockerClusterArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		var resp cs.ClusterCreationResponse
		resp, err = aliyun.CSClient().CreateCluster(common.Region(aliyun.Region), arg)

		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("DOCKER-CLUSTER-ID", resp.ClusterID).
			WithField("DOCKER-CLUSTER-NAME", arg.Name).Infoln("Docker cluster created")
	}

	return
}

func DeleteDockerCluster(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.DeleteDockerClusterArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		if arg.State == cs.Deleting ||
			arg.State == cs.Deleted {

			logrus.WithField("CODE", aliyun.Code).
				WithField("DOCKER-CLUSTER-ID", arg.ClusterID).
				WithField("DOCKER-CLUSTER-NAME", arg.Name).Infof("Docker cluster already in status of %s", arg.State)

			continue
		}

		err = aliyun.CSClient().DeleteCluster(arg.ClusterID)

		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("DOCKER-CLUSTER-ID", arg.ClusterID).
			WithField("DOCKER-CLUSTER-NAME", arg.Name).Infoln("Docker cluster deleted")
	}

	return
}

func WaitForAllClusterRunning(ctx context.Context, conf config.Configuration) (err error) {
	return waitCSClusterStatusTo(ctx, conf, cs.Running, 600)
}

func WaitForAllClusterDeleted(ctx context.Context, conf config.Configuration) (err error) {
	return waitCSClusterStatusTo(ctx, conf, cs.Deleted, 300)
}

func waitCSClusterStatusTo(ctx context.Context, conf config.Configuration, status cs.ClusterState, timeout int) (err error) {
	aliyun := NewAliyun(ctx, conf)

	clusters, err := aliyun.CSClient().DescribeClusters("")

	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}

	errChan := make(chan error, 1)

	for _, cluster := range clusters {
		wg.Add(1)

		go func(cluster cs.ClusterType) {

			defer wg.Done()

			logrus.WithField("CODE", aliyun.Code).
				WithField("DOCKER-CLUSTER-ID", cluster.ClusterID).
				WithField("DOCKER-CLUSTER-NAME", cluster.Name).Infof("Waiting for cluster status to %s", status)

			e := aliyun.CSClient().WaitForClusterAsyn(cluster.ClusterID, status, timeout)

			if e != nil {

				if IsAliErrCode(e, "ErrorClusterNotFound") {
					return
				}

				logrus.WithField("CODE", aliyun.Code).
					WithError(e).
					WithField("DOCKER-CLUSTER-ID", cluster.ClusterID).
					WithField("DOCKER-CLUSTER-NAME", cluster.Name).Errorln("Wait for cluster status to %s failure", status)

				select {
				case errChan <- e:
				default:
				}
				return
			}

			logrus.WithField("CODE", aliyun.Code).
				WithField("DOCKER-CLUSTER-ID", cluster.ClusterID).
				WithField("DOCKER-CLUSTER-NAME", cluster.Name).Infof("Cluster status is %s", status)

		}(cluster)
	}

	wg.Wait()

	select {
	case err = <-errChan:
	default:
	}

	return
}

func CreateDockerClusterVolume(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateDockerClusterVolumnArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		var client *cs.ProjectClient
		client, err = aliyun.CSClient().GetProjectClient(arg.Cluster.ClusterID)

		if err != nil {
			return
		}

		err = client.CreateVolume(&arg.VolumeCreationArgs)

		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("DOCKER-CLUSTER-ID", arg.Cluster.ClusterID).
			WithField("DOCKER-CLUSTER-NAME", arg.Cluster.Name).Infoln("Docker cluster created")
	}

	return
}

func CreateDockerClusterProject(ctx context.Context, conf config.Configuration) (err error) {
	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.CreateDockerProjectArgs()

	if err != nil {
		return
	}

	for _, arg := range args {

		err = arg.Wait()
		if err != nil {
			return
		}

		err = arg.Create()
		if err != nil {
			return
		}

		logrus.WithField("CODE", aliyun.Code).
			WithField("DOCKER-CLUSTER-ID", arg.Cluster.ClusterID).
			WithField("DOCKER-CLUSTER-NAME", arg.Cluster.Name).
			WithField("DOCKER-PROJECT-NAME", arg.CreationArgs.Name).
			Infoln("Docker cluster project created")

	}

	return
}

func DeleteDockerClusterProject(ctx context.Context, conf config.Configuration) (err error) {

	aliyun := NewAliyun(ctx, conf)

	args, err := aliyun.DeleteDockerProjectArgs()

	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}
	errChan := make(chan error, 1)

	for clusterName, projects := range args {
		for _, proj := range projects {
			wg.Add(1)
			go func(proj *DockerProject) {
				defer wg.Done()

				e := proj.Client.Client.DeleteProject(proj.Project.Name, true, true)
				if e != nil {
					select {
					case errChan <- e:
					default:
					}

					return
				}

				logrus.WithField("CODE", aliyun.Code).
					WithField("DOCKER-CLUSTER-ID", proj.Client.Cluster.ClusterID).
					WithField("DOCKER-CLUSTER-NAME", clusterName).
					WithField("DOCKER-PROJECT-NAME", proj.Project.Name).Infoln("Docker cluster project deleted")

			}(proj)
		}
	}

	wg.Wait()

	select {
	case err = <-errChan:
	default:
	}

	return
}
