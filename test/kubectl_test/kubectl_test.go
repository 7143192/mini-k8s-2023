package kubectl_test

import (
	"mini-k8s/pkg/kubectl/cmds"
	"testing"
)

func TestKubectl(t *testing.T) {
	// create a new APP here.
	app := cmds.TestInit()
	// createCmd test.
	// hello test.
	err := cmds.TestParseArgs(app, "kubectl helloworld")
	if err != nil {
		t.Error("Fail kubectl hello test!")
		return
	} else {
		t.Log("Pass kubectl hello test!")
	}
	// create pod test.
	err = cmds.TestParseArgs(app, "kubectl create --file ../../utils/templates/pod_template.yaml")
	if err != nil {
		t.Error("Fail kubectl create pod test!")
		return
	} else {
		t.Log("Pass kubectl create pod test!")
	}
	// create ClusterIP service test.
	err = cmds.TestParseArgs(app, "kubectl create --file ../../utils/templates/service_template.yaml")
	if err != nil {
		t.Error("Fail kubectl create svc test!")
		return
	} else {
		t.Log("Pass kubectl create svc test!")
	}
	// TODO: in version 2.0, no test for NodePort Service.

	// getCmd test.
	// as tests for api-server have already tested the correctness of get info and describe info, we onyl test HTTP correction here.
	err = cmds.TestParseArgs(app, "kubectl get pods")
	if err != nil {
		t.Error("Fail kubectl get pod test!")
		return
	} else {
		t.Log("Pass kubectl get pod test!")
	}
	err = cmds.TestParseArgs(app, "kubectl get services")
	if err != nil {
		t.Error("Fail kubectl get svc test!")
		return
	} else {
		t.Log("Pass kubectl get svc test!")
	}
	// describeCmd test.
	err = cmds.TestParseArgs(app, "kubectl describe pod s")
	if err != nil {
		t.Error("Fail kubectl describe pod test!")
		return
	} else {
		t.Log("Pass kubectl describe pod test!")
	}
	err = cmds.TestParseArgs(app, "kubectl describe service service1")
	if err != nil {
		t.Error("Fail kubectl describe svc test!")
		return
	} else {
		t.Log("Pass kubectl describe svc test!")
	}
	// deleteCmd test.
	err = cmds.TestParseArgs(app, "kubectl del pod s")
	if err != nil {
		t.Error("Fail kubectl delete pod test!")
		return
	} else {
		t.Log("Pass kubectl delete pod test!")
	}
	// TODO: in version 2.0, only delete pod is supported.

}
