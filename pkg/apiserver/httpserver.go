package apiserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/controller/replicaSet"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/scheduler"
	"mini-k8s/pkg/serverless"
)

type Server struct {
	engine               *gin.Engine
	es                   *etcd.Store
	NodeInfo             *defines.NodeInfo
	Scheduler            *scheduler.Scheduler
	Services             []*defines.EtcdService
	ReplicaController    *replicaSet.ReplicaController
	AutoScalerController *defines.AutoScalerController
	ServerlessController *serverless.SLController
}

func ServerInit() *Server {
	es := etcd.StoreStart()
	engine := gin.Default()
	apiServer := &Server{
		engine:   engine,
		es:       es,
		NodeInfo: nil,
	}
	api := engine.Group("/objectAPI")
	{
		api.POST("/createPod", apiServer.ServerCreatePod)
		api.POST("/deletePod", apiServer.ServerDeletePod)
		api.POST("/describePod", apiServer.ServerDescribePod)
		api.POST("/describeNode", apiServer.ServerDescribeNode)
		api.POST("/describeService", apiServer.ServerDescribeService)
		api.POST("/describeDNS", apiServer.ServerDescribeDNS)
		api.POST("/getPod", apiServer.ServerGetPods)
		api.POST("/getNode", apiServer.ServerGetNodes)
		api.POST("/getService", apiServer.ServerGetServices)
		api.POST("/getDNS", apiServer.ServerGetDNSs)
		api.POST("/registerNewNode", apiServer.ServerRegisterNewNode)
		api.POST("/heartBeat", apiServer.ServerReceiveNodeHeartBeat)
		api.POST("/updateNodeHealthState", apiServer.ServerUpdateNodeHealthState)
		api.POST("/createClusterIP", apiServer.ServerCreateClusterIPSvc)
		api.POST("/createNodePort", apiServer.ServerCreateNodePortSvc)
		// used for scheduler.
		api.GET("/getNodes", apiServer.ServerGetAllNodes)
		api.GET("/getUnhandledPods", apiServer.ServerGetUnhandledPods)
		api.POST("/updatePod", apiServer.ServerUpdatePodNodeId)
		api.POST("/createAutoScaler", apiServer.ServerCreateAutoScaler)
		api.POST("/deleteAutoScaler", apiServer.ServerDeleteAutoScaler)
		// api used to handle etcd-put for resources info collected by cadvisor.
		api.POST("/storePodResource", apiServer.ServerStorePodResource)
		api.POST("/storeNodeResource", apiServer.ServerStoreNodeResource)
		// api related to autoScaler pod replicas changes.
		api.POST("/addAutoPodReplica", apiServer.ServerCreateAutoPodReplica)
		api.POST("/delAutoPodReplica", apiServer.ServerDeleteAutoPodReplica)
		// api used to get and describe autoScaler.
		api.POST("/getAutoScaler", apiServer.ServerGetAutoScalers)
		api.POST("/describeAutoScaler", apiServer.ServerDescribeAutoScaler)
		api.POST("/updateEtcdAutoControllerList", apiServer.ServerUpdateEtcdAutoControllerList)
		// api used to update pod restart info.
		api.POST("/updatePodRestart", apiServer.ServerUpdatePodRestartInfo)
		// api related to deployment.
		api.POST("/createDeployment", apiServer.ServerCreateDeployment)
		api.POST("/createReplicaSet", apiServer.ServerCreateReplicaSet)
		api.POST("/deleteReplicaSet", apiServer.ServerDeleteReplicaSet)
		api.POST("/getReplicaSet", apiServer.ServerGetReplicaSet)
		api.POST("/describeReplicaSet", apiServer.ServerDesReplicaSet)
		// used for RC
		api.GET("/getPods", apiServer.ServerGetAllPods)
		api.GET("/getAllReplicaSets", apiServer.ServerGetAllReplicaSets)
		api.POST("/createDNS", apiServer.ServerCreateDNS)
		api.POST("/getAllServices", apiServer.ServerGetAllServices)
		// api used for GPUJob.
		api.POST("/createGPUJob", apiServer.ServerCreateGPUJob)

		api.POST("/updatePodInfoAfterRealOp", apiServer.ServerUpdatePodInfoAfterRealOp)
		api.POST("/updateNodeInfoAfterRealOp", apiServer.ServerUpdateNodeInfoAfterRealOp)
		//used for Serverless
		api.POST("/initFunc", apiServer.ServerInitFunc)
		api.GET("/getServerlessObject/:type", apiServer.ServerGetServerlessObject)
		api.DELETE("/deleteOldWorkflows", apiServer.ServerDelOldWorkflows)
		api.POST("/newFunction/:op", apiServer.ServerNewFunc)
		api.POST("/newWorkFlow/:op", apiServer.ServerNewWorkFlow)
		api.GET("/getFuncPods/:name", apiServer.ServerGetFuncPods)
		// api.DELETE("/delFuncPod/:name", apiServer.ServerDelFuncPod)
		api.POST("/describeGPUJob", apiServer.ServerDescribeGPUJob)
		api.POST("/getAllGPUJobs", apiServer.ServerGetAllGPUJobs)
		api.POST("/updateGPUJobState", apiServer.ServerUpdateGPUJobState)
		api.POST("/deleteGPUJob", apiServer.ServerDeleteGPUJob)
		//api.DELETE("/delFuncPod/:name", apiServer.ServerDelFuncPod)
		api.POST("/delImage", apiServer.ServerDelImage)
		api.POST("/trigger/:name/:type", apiServer.ServerTrigger)
		api.POST("/getAutoControllerScalers", apiServer.ServerGetAllControllerScalers)
		api.DELETE("/deleteServerlessObject/:type/:name", apiServer.ServerDeleteServerlessObject)
		api.POST("/updatePodHealthState", apiServer.ServerUpdatePodHealthState)
		api.POST("/getAllReplicaPodsNames", apiServer.ServerGetAllReplicaPodNames)
	}
	return apiServer
}

func KubeletServerInit() *Server {
	es := etcd.StoreStart()
	engine := gin.Default()
	kubeletServer := &Server{
		engine:   engine,
		es:       es,
		NodeInfo: nil,
	}
	api := engine.Group("/objectAPI")
	{
		api.POST("/changeNodePod", kubeletServer.KubeletServerHandlePodChange)
		// used to get scheduler-required info.
		api.POST("/getNodeResourceUsage", kubeletServer.ServerGetNodeResourceUsage)
		api.POST("/changeNodePodNew", kubeletServer.KubeletServerHandlePodChangeNew)
		api.POST("/changeNodeRealPod", kubeletServer.KubeletServerHandleRealPodChanges)
		// api used to delete one image in the node.
		api.POST("/removeOneImage", kubeletServer.ServerRemoveOneImage)
		api.POST("/checkRSPodState", kubeletServer.ServerCheckRCPodState)
	}
	return kubeletServer
}

func KubeProxyServerInit() *Server {
	es := etcd.StoreStart()
	engine := gin.Default()
	kubeProxyServer := &Server{
		engine:   engine,
		es:       es,
		NodeInfo: nil,
		Services: make([]*defines.EtcdService, 0),
	}
	api := engine.Group("/objectAPI")
	{
		api.POST("/addClusterIPRule", kubeProxyServer.ServerAddClusterIPRule)
		api.POST("/addNodePortRule", kubeProxyServer.ServerAddNodePortRule)
		api.POST("/delClusterIPRule", kubeProxyServer.ServerDelClusterIPRule)
		api.POST("/delNodePortRule", kubeProxyServer.ServerDelNodePortRule)
		api.POST("/createClusterIPService", kubeProxyServer.ServerCreateClusterIPService)
		api.POST("/createNodePortService", kubeProxyServer.ServerCreateNodePortService)
	}
	return kubeProxyServer
}

//func KubeDNSServerInit() *Server {
//	es := etcd.StoreStart()
//	engine := gin.Default()
//	kubeDNSServer := &Server{
//		engine:   engine,
//		es:       es,
//		NodeInfo: nil,
//	}
//	api := engine.Group("/objectAPI")
//	{
//		api.POST("/createActualDNS", kubeDNSServer.ServerCreateActualDNS)
//	}
//	return kubeDNSServer
//}

func SchedulerServerInit() *Server {
	es := etcd.StoreStart()
	engine := gin.Default()
	SchedulerServer := &Server{
		engine:   engine,
		es:       es,
		NodeInfo: nil,
	}
	api := engine.Group("/objectAPI")
	{
		api.POST("/schedulePodNode", SchedulerServer.ServerScheduleNodeForPod)
		api.POST("/schedulePolicy", SchedulerServer.ServerSchedulePolicy)
	}
	return SchedulerServer
}

func AutoScalerServerInit() *Server {
	es := etcd.StoreStart()
	engine := gin.Default()
	AutoScalerServer := &Server{
		engine:    engine,
		es:        es,
		NodeInfo:  nil,
		Scheduler: nil,
	}
	api := engine.Group("/objectAPI")
	{
		api.POST("/addAutoScaler", AutoScalerServer.ServerHandleAddAutoScaler)
		api.POST("/removeAutoScaler", AutoScalerServer.ServerHandleRemoveAutoScaler)
	}
	return AutoScalerServer
}

func RCServerInit() *Server {
	engine := gin.Default()
	RCServer := &Server{
		engine: engine,
	}

	api := engine.Group("/objectAPI")

	{
		api.POST("/create", RCServer.ServerHandleNewReplica)
		api.POST("/delete", RCServer.ServerHandleDelReplica)
		api.POST("/describeReplicaSet", RCServer.ServerHandleDesReplicaSet)
	}
	return RCServer
}

func SLServerInit() *Server {
	engine := gin.Default()
	SLServer := &Server{
		engine: engine,
	}

	api := engine.Group("/objectAPI")

	{
		api.POST("/newFunction/:op", SLServer.ServerHandleNewFunction)
		api.POST("/trigger/:function", SLServer.ServerHandleTrigger)
		api.POST("/newWorkFlow/:op", SLServer.ServerHandleNewWorkFlow)
		api.POST("/triggerWorkFlow/:workflow", SLServer.ServerHandleTriggerWorkFlow)
		api.DELETE("/del/:type/:name", SLServer.ServerHandleDelServerlessObject)
	}

	return SLServer
}

func TestServerInit() *Server {
	engine := gin.Default()
	testServer := &Server{
		engine: engine,
	}
	return testServer
}

func (s *Server) Run() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.MasterPort))
}

func (s *Server) KubeletServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.WorkerPort))
}

func (s *Server) KubeProxyServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.ProxyPort))
}

func (s *Server) SchedulerServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.SchedulerPort))
}

func (s *Server) AutoScalerServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.AutoScalerPort))
}

func (s *Server) RCServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.RCPort))
}

func (s *Server) SLServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.SLPort))
}

func (s *Server) KubeDNSServerRun() error {
	return s.engine.Run(fmt.Sprintf(":%s", config.DNSPort))
}

func (s *Server) TestServerRun() error {
	return s.engine.Run(fmt.Sprintf(":11100"))
}
