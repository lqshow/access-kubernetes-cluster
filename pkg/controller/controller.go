package controller

import (
	"time"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"

	"github.com/lqshow/access-kubernetes-cluster/pkg/informer"
)

type Controller struct {
	informerFactory informers.SharedInformerFactory
}

func NewController(informerFactory informers.SharedInformerFactory) *Controller {
	return &Controller{
		informerFactory: informerFactory,
	}
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	// Kubernetes serves an utility to handle API crashes
	defer runtime.HandleCrash()
	zap.S().Debugf("Starting Shared Informer Controller Manager.")

	// Starts all the shared informers that have been created by the factory so far.
	//go c.informerFactory.Start(stopCh)

	// defined for which resource to be informed, we will be informed for nodes
	nodeController := informer.NewNodeController(c.informerFactory)
	if err := nodeController.Run(stopCh); err != nil {
		return err
	}
	//if err := nodeController.List(); err != nil {
	//	zap.S().Debugf("node list err: %v", err)
	//	return err
	//}

	deployController := informer.NewDeploymentController(c.informerFactory)
	if err := deployController.Run(stopCh); err != nil {
		return err
	}

	// defined for which resource to be informed, we will be informed for pods
	podController := informer.NewPodController(c.informerFactory)
	if err := podController.Run(stopCh); err != nil {
		return err
	}
	//if err := podController.List(); err != nil {
	//	zap.S().Debugf("pod list err: %v", err)
	//	return err
	//}

	klog.Info("Starting workers")
	// Launch three workers to process user-defined resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(deployController.RunWorker, time.Second, stopCh)
		go wait.Until(podController.RunWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
