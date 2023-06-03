package main

import (
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/scheduler"
)

func main() {
	//s := scheduler.NewScheduler(defines.PodIdSetPrefix)
	//s.Run()

	s := scheduler.NewScheduler( /*defines.PodIdSetPrefix*/ )
	schedulerServer := apiserver.SchedulerServerInit()
	schedulerServer.Scheduler = s
	err := schedulerServer.SchedulerServerRun()
	if err != nil {
		panic(err)
	}
}
