package informer

import (
	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
)

type Controller struct {
	informerFactory informers.SharedInformerFactory
}

func NewController(informerFactory informers.SharedInformerFactory) *Controller {
	return &Controller{
		informerFactory: informerFactory,
	}
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	// Kubernetes serves an utility to handle API crashes
	defer runtime.HandleCrash()
	zap.S().Debugf("Starting Controller Manager.")

	// Starts all the shared informers that have been created by the factory so far.
	//go c.informerFactory.Start(stopCh)

	// node informer
	nodeController := NewNodeController(c.informerFactory)
	if err := nodeController.Run(stopCh); err != nil {
		return err
	}
	if err := nodeController.List(); err != nil {
		zap.S().Debugf("node list err: %v", err)
		return err
	}

	// pod informer
	podController := NewPodController(c.informerFactory)
	if err := podController.Run(stopCh); err != nil {
		return err
	}
	if err := podController.List(); err != nil {
		zap.S().Debugf("pod list err: %v", err)
		return err
	}

	return nil
}
