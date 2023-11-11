package src

import (
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

type PrintNode struct {
	Name             string
	Os               string
	Arch             string
	ContainerRuntime string
}

type PrintNodeData struct {
	Config    PrintNodeDataConfig
	Namespace string
	Nodes     []PrintNode
}

type PrintNodeDataConfig struct {
	ShowNamespaces bool
	ShowTimes      bool
	ShowReqLimits  bool
}

type Print interface {
	Printout(cls bool) error
}

func (pnd PrintNodeData) Printout(cls bool) error {
	if cls {
		fmt.Print("\033[2J\033[0;0H")
	}
	if pnd.Nodes == nil {
		return errors.New("list of view nodes must not be null")
	}
	if pnd.Namespace != "" {
		fmt.Printf("namespace: %s\n", pnd.Namespace)
	}
	l := len(pnd.Nodes)
	if l <= 1 {
		fmt.Println("no nodes to display...")
		return nil
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NODE", "OS", "ARCH", "RUNTIME"})
	for _, n := range pnd.Nodes {
		if n.Name != "" {
			t.AppendRow([]interface{}{n.Name, n.Os, n.Arch, n.ContainerRuntime})
			//fmt.Printf("\n- %s : \n\tOS: %s\n\tArch: %s\n\tContainer Runtime: %s", n.Name, n.Os, n.Arch, n.ContainerRuntime)
		}
	}
	t.Render()
	return nil
}
