package cmd

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"kndu/cmd/src"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var namespace string
var allNamespacesFlag bool
var nodeFilter string
var podFilter string
var showContainersFlag bool
var containerViewTypeBlockFlag bool
var showTimesFlag bool
var showRunningFlag bool
var showReqLimitsFlag bool
var showMetricsFlag bool
var verbosity string
var kubeconfig string
var watchOn bool

var rootCmd = &cobra.Command{
	Use:   "kndu",
	Short: "'kndu' displays nodes and information about their filesystems.",
	Long: `
The 'kndu' plugin displays nodes and their disk usage + information about mounts and filesystems.
You can find the source code and usage documentation at GitHub: https://github.com/anishmukherjee123/kndu.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stopCh := make(chan bool)
		errCh := make(chan error)
		go handleErrors(errCh)
		var wg sync.WaitGroup
		wg.Add(1)
		go schedule(&wg, stopCh, errCh)
		if !watchOn {
			close(stopCh)
		}
		wg.Wait()
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func schedule(wg *sync.WaitGroup, stop <-chan bool, errCh chan<- error) {
	defer wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	vnd := executeLoadAndFilter(errCh)
	executePrintOut(vnd, errCh)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
			vnd = executeLoadAndFilter(errCh)
			executePrintOut(vnd, errCh)
		}
	}
}

func executeLoadAndFilter(errCh chan<- error) src.PrintNodeData {
	setup := src.Setup{KubeCfgPath: kubeconfig}
	err := setup.Initialize()
	if err != nil {
		errCh <- fmt.Errorf("init setup failed (%w)", err)
		return src.PrintNodeData{}
	}
	if namespace != "" {
		setup.Namespace = namespace
	}
	if allNamespacesFlag {
		setup.Namespace = ""
	}
	api := src.KubernetesApi{
		Setup: &setup,
	}
	ns := src.NodeSpec{
		Api: api,
	}
	var pns []src.PrintNode
	log.Tracef("getting nodes...")
	pns, err = ns.ViewNodes(pns)
	if err != nil {
		log.Debugf("ERROR: %s", err.Error())
	}
	pnd := src.PrintNodeData{
		Namespace: setup.Namespace,
		Nodes:     pns,
	}
	pnd.Config.ShowNamespaces = allNamespacesFlag
	pnd.Config.ShowTimes = showTimesFlag
	pnd.Config.ShowReqLimits = showReqLimitsFlag

	return pnd
}

func executePrintOut(pnd src.PrintNodeData, errCh chan<- error) {
	err := pnd.Printout(watchOn)
	if err != nil {
		errCh <- err
		return
	}
}

func handleErrors(errCh <-chan error) {
	for err := range errCh {
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		log.SetFormatter(&nested.Formatter{
			ShowFullLevel: true,
			HideKeys:      true,
			FieldsOrder:   []string{"component", "category"},
		})
		if err := initLog(os.Stdout, verbosity); err != nil {
			return err
		}
		return nil
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace to use")
	rootCmd.Flags().BoolVarP(&allNamespacesFlag, "all-namespaces", "A", false, "use all namespaces")
	rootCmd.Flags().BoolVarP(&showReqLimitsFlag, "show-requests-and-limits", "r", false, "show requests and limits for containers' cpu and memory (requires -c flag)")
	rootCmd.Flags().BoolVarP(&showTimesFlag, "show-pod-start-times", "t", false, "show start times of pods")
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", log.WarnLevel.String(), "defines log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "kubectl configuration file (default: ~/.kube/config or env: $KUBECONFIG)")
	rootCmd.PersistentFlags().BoolVarP(&watchOn, "watch", "w", false, "executes the command every second so that changes can be observed")
}

func initConfig() {
}

func initLog(out io.Writer, verbosity string) error {
	log.SetOutput(out)
	level, err := log.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}
