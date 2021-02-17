package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"k8s.io/client-go/informers"

	"github.com/lqshow/access-kubernetes-cluster/pkg/kubernetes/client"
	"github.com/lqshow/access-kubernetes-cluster/pkg/signals"
	"github.com/lqshow/access-kubernetes-cluster/service"
	"github.com/lqshow/access-kubernetes-cluster/version"

	clientsetexample "github.com/lqshow/access-kubernetes-cluster/pkg/clientset"
	pkgcontroller "github.com/lqshow/access-kubernetes-cluster/pkg/controller"
)

func main() {
	showVersion := pflag.BoolP("version", "v", false, "Show version")

	pflag.Parse()
	if *showVersion {
		v, _ := json.MarshalIndent(version.Get(), "", "  ")
		fmt.Println(string(v))
		os.Exit(0)
	}

	// Set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupStopSignalHandler()

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
	kubeClientSet, err := client.NewKubeClient("", config.KubeConfig)
	if err != nil {
		zap.S().Fatalf("Failed to get kube client: %v", err)
	}
	zap.L().Info("Kubernetes connected")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	clientsetExample := clientsetexample.NewPodExample(kubeClientSet, config, ctx)
	if err := clientsetExample.List(); err != nil {
	}

	// Create the shared informer factory and use the client to connect to Kubernetes
	factory := informers.NewSharedInformerFactory(kubeClientSet, 0)
	controller := pkgcontroller.NewController(factory)

	if err := controller.Run(config.WorkerThreadiness, stopCh); err != nil {
		zap.S().Panicf("Failed to controller run: %v", err)
	}
}
