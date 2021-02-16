package informer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type NodeController struct {
	informerFactory informers.SharedInformerFactory
	nodeInformer    coreinformers.NodeInformer
	nodeLister      corelisters.NodeLister
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *NodeController) Run(stopCh <-chan struct{}) error {
	// run starts and runs the shared informer
	go c.nodeInformer.Informer().Run(stopCh)

	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.nodeInformer.Informer().HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	return nil
}

func (c *NodeController) List() error {
	// List lists all Nodes in the indexer.
	nodeList, err := c.nodeLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, node := range nodeList {
		klog.Infof("Got node detail info, name: %s, status: %v, labels: %v", node.Name, node.Status.Phase, node.Labels)
	}

	return nil
}

func (c *NodeController) onAdd(obj interface{}) {
	node := obj.(*corev1.Node)

	klog.Infof("NODE CREATED: %s", node.Name)
}

func (c *NodeController) onUpdate(old, new interface{}) {
	oldNode := old.(*corev1.Node)
	newNode := new.(*corev1.Node)
	if oldNode.ResourceVersion == newNode.ResourceVersion {
		return
	}

	klog.Info("NODE UPDATED: not implemented")
}

func (c *NodeController) onDelete(obj interface{}) {
	klog.Info("NODE UPDATED: not implemented")
}

func NewNodeController(informerFactory informers.SharedInformerFactory) *NodeController {
	// node informer
	nodeInformer := informerFactory.Core().V1().Nodes()
	// create node lister
	nodeLister := nodeInformer.Lister()

	c := &NodeController{
		informerFactory: informerFactory,
		nodeInformer:    nodeInformer,
		nodeLister:      nodeLister,
	}

	// Your custom resource event handlers.
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Called on creation
		AddFunc: c.onAdd,
		// Called on resource update and every resyncPeriod on existing resources.
		UpdateFunc: c.onUpdate,
		// Called on resource deletion.
		DeleteFunc: c.onDelete,
	})

	return c
}
