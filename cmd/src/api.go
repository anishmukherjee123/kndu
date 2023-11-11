package src

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type Api interface {
	RetrieveNodeList() (*v1.NodeList, error)
}

type KubernetesApi struct {
	Setup *Setup
}

type NodeSpec struct {
	Api Api
}

func (k KubernetesApi) RetrieveNodeList() (*v1.NodeList, error) {
	nl, err := k.Setup.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nl, nil
}

func (ns NodeSpec) ViewNodes(vns []PrintNode) (result []PrintNode, err error) {
	list, err := ns.Api.RetrieveNodeList()
	if err != nil {
		return nil, err
	}
	if vns == nil {
		vns = make([]PrintNode, 1, len(list.Items)+1)
		vns[0].Name = "" // create placeholder for unscheduled pods
	}
	for _, n := range list.Items {
		vn := PrintNode{
			Name:             n.Name,
			Os:               n.Status.NodeInfo.OperatingSystem,
			Arch:             n.Status.NodeInfo.Architecture,
			ContainerRuntime: n.Status.NodeInfo.ContainerRuntimeVersion,
		}
		vns = append(vns, vn)
	}
	return vns, nil
}
