package informer

import (
	"time"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
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
	nodeController := NewNodeController(c.informerFactory)
	if err := nodeController.Run(stopCh); err != nil {
		return err
	}
	//if err := nodeController.List(); err != nil {
	//	zap.S().Debugf("node list err: %v", err)
	//	return err
	//}

	// defined for which resource to be informed, we will be informed for pods
	podController := NewPodController(c.informerFactory)
	if err := podController.Run(stopCh); err != nil {
		return err
	}
	//if err := podController.List(); err != nil {
	//	zap.S().Debugf("pod list err: %v", err)
	//	return err
	//}

	for i := 0; i < threadiness; i++ {
		go wait.Until(podController.Worker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
