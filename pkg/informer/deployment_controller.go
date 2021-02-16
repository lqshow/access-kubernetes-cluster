package informer

import (
	"fmt"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	applisters "k8s.io/client-go/listers/apps/v1"
)

type DeploymentController struct {
	informer         cache.SharedIndexInformer
	deploymentLister applisters.DeploymentLister

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *DeploymentController) Run(stopCh <-chan struct{}) error {
	// run starts and runs the shared informer
	go c.informer.Run(stopCh)

	// wait for the initial synchronization of the local cache.
	klog.Info("Waiting for informer caches to sync.")
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *DeploymentController) RunWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *DeploymentController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		var (
			key string
			ok  bool
		)
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)

		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *DeploymentController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	deploy, err := c.deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infof("Deploy: %s/%s does not exist in local cache, will delete it ...", namespace, name)
			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list deploy by: %s/%s", namespace, name))
		return err
	}
	klog.Infof("Try to process deploy, name: %v, ResourceVersion: %v ...", deploy.Name, deploy.ResourceVersion)

	return nil
}

// enqueue takes a Deployment resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Deployment.
func (c *DeploymentController) enqueue(obj interface{}) {
	var (
		key string
		err error
	)
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	c.workqueue.AddRateLimited(key)
	deploy := obj.(*v1.Deployment)
	klog.Infof("DEPLOYMENT SYNCED: %s/%s %v", deploy.Namespace, deploy.Name, deploy.Status.AvailableReplicas)
}

// enqueueForDelete takes a deleted Deployment resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Deployment.
func (c *DeploymentController) enqueueForDelete(obj interface{}) {
	var (
		key string
		err error
	)
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)

	deploy := obj.(*v1.Deployment)
	klog.Infof("DEPLOYMENT DELETED: %s/%v", deploy.Namespace, deploy.Name)
}

func NewDeploymentController(informerFactory informers.SharedInformerFactory) *DeploymentController {
	// Deployment Informer
	deployInformer := informerFactory.Apps().V1().Deployments()
	// create informer
	informer := deployInformer.Informer()
	// create deployment lister
	deploymentLister := deployInformer.Lister()

	c := &DeploymentController{
		informer:         informer,
		deploymentLister: deploymentLister,

		// create the workqueue
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "deployments"),
	}

	// Set up an event handler for when Deployment resources change
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(old, new interface{}) {
			oldDeploy := old.(*v1.Deployment)
			newDeploy := new.(*v1.Deployment)
			if oldDeploy.ResourceVersion == newDeploy.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			c.enqueue(new)
		},
		DeleteFunc: c.enqueueForDelete,
	})

	return c
}
