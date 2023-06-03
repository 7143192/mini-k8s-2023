package apiserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"time"
)

func CreateDeployment(cli *clientv3.Client, deploy *defines.Deployment) *defines.EtcdDeployment {
	// first create a replicaSet object from deployment object.
	replicaSet := &defines.ReplicaSet{}
	replicaSet.ApiVersion = deploy.ApiVersion
	replicaSet.Kind = "ReplicaSet"
	replicaSet.Metadata.Name = defines.DeploymentReplicaSetPrefix + deploy.Metadata.Name
	replicaSet.Spec.Replicas = int32(deploy.Spec.Replicas)
	replicaSet.Spec.Selector.MatchLabels.App = deploy.Spec.Selector.Labels.App
	replicaSet.Spec.Selector.MatchLabels.Env = deploy.Spec.Selector.Labels.Env
	replicaSet.Spec.Template.Metadata.Label.App = deploy.Spec.Template.Metadata.Labels.App
	replicaSet.Spec.Template.Metadata.Label.Env = deploy.Spec.Template.Metadata.Labels.Env
	replicaSet.Spec.Template.Spec = deploy.Spec.Template.Spec
	// then create a new replicaSet object into system through reusing the logic implemented by God-Z.
	err := CreateReplicaSet(cli, replicaSet)
	if err != nil {
		log.Printf("[apiserver] an error occurs when creating a new replicaSet in func CreateDeployment: %v\n", err)
		return nil
	}
	// then generate a etcd-version deployment object from deployment object.
	etcdDeploy := &defines.EtcdDeployment{}
	etcdDeploy.DeployName = deploy.Metadata.Name
	etcdDeploy.PodName = deploy.Spec.Template.Metadata.Name
	etcdDeploy.Replicas = deploy.Spec.Replicas
	etcdDeploy.StartTime = time.Now()
	etcdDeploy.RsInfo = replicaSet
	// then store this object into etcd.
	key := defines.DeploymentPrefix + "/" + deploy.Metadata.Name
	etcdDeployByte, _ := yaml.Marshal(etcdDeploy)
	etcd.Put(cli, key, string(etcdDeployByte))
	// then store this deployment object name into deployments list.
	allKey := defines.DeploymentListPrefix + "/"
	names := make([]string, 0)
	kv := etcd.Get(cli, allKey).Kvs
	if len(kv) == 0 {
		names = append(names, deploy.Metadata.Name)
	} else {
		_ = yaml.Unmarshal(kv[0].Value, &names)
		names = append(names, deploy.Metadata.Name)
	}
	namesByte, _ := yaml.Marshal(&names)
	etcd.Put(cli, key, string(namesByte))
	return etcdDeploy
}
