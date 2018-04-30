package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/ElisaOyj/openshift-lb-controller/pkg/controller"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// lb providers
	_ "github.com/ElisaOyj/openshift-lb-controller/pkg/controller/providers/f5"
)

func main() {
	log.SetOutput(os.Stdout)

	sigs := make(chan os.Signal, 1)
	stop := make(chan struct{})

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	wg := &sync.WaitGroup{}

	runOutsideCluster := flag.Bool("run-outside-cluster", false, "Set this flag when running outside of the cluster.")
	flag.Parse()
	// Create clientset for interacting with the kubernetes cluster
	clientset, config, err := newClientSet(*runOutsideCluster)

	if err != nil {
		panic(err.Error())
	}

	go controller.NewRouteController(clientset, config).Run(stop, wg)

	<-sigs
	log.Printf("Shutting down...")

	close(stop)
	wg.Wait()
}

func newClientSet(runOutsideCluster bool) (*kubernetes.Clientset, *restclient.Config, error) {
	kubeConfigLocation := ""
	if runOutsideCluster == true {
		homeDir := os.Getenv("HOME")
		kubeConfigLocation = filepath.Join(homeDir, ".kube", "config")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)

	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, config, err
}
