package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"k8s.io/client-go/informers"
	"os"

	"github.com/lqshow/access-kubernetes-cluster/pkg/kube"
	"github.com/lqshow/access-kubernetes-cluster/service"
	"github.com/lqshow/access-kubernetes-cluster/version"

	clientsetexample "github.com/lqshow/access-kubernetes-cluster/pkg/clientset"
	informercontroller "github.com/lqshow/access-kubernetes-cluster/pkg/informer"
)

func main() {
	showVersion := pflag.BoolP("version", "v", false, "Show version")

	pflag.Parse()
	if *showVersion {
		v, _ := json.MarshalIndent(version.Get(), "", "  ")
		fmt.Println(string(v))
		os.Exit(0)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load Config
	config := service.LoadConfigFromEnv()

	// new logger
	var logger *zap.Logger
	if config.DevMode {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	// flushes buffer, if any
	defer logger.Sync()
	zap.ReplaceGlobals(logger)
	zap.RedirectStdLog(logger)

	zap.S().Info("Connecting to Kubernetes")
	kubeClientSet, err := kube.GetKubeClientset(config)
	if err != nil {
		zap.S().Fatalf("Failed to get kube client: %v", err)
	}
	zap.L().Info("Kubernetes connected")

	clientsetExample := clientsetexample.NewPodExample(kubeClientSet, config, ctx)
	if err := clientsetExample.List(); err != nil {

	}

	// Create the shared informer factory and use the client to connect to Kubernetes
	factory := informers.NewSharedInformerFactory(kubeClientSet, 0)
	controller := informercontroller.NewController(factory)

	// Create a channel to stops the shared informer gracefully
	stopCh := make(chan struct{})
	defer close(stopCh)

	if err := controller.Run(stopCh); err != nil {
		zap.S().Panicf("Failed to controller run: %v", err)
	}

	select {}
	zap.S().Debugf("Shutting down Controller.")
}
