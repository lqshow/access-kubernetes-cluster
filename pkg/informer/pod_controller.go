package informer

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	//coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// PodController logs the name and namespace of pods that are added,
// deleted, or updated
type PodController struct {
	podLister corelisters.PodLister

	informer cache.SharedIndexInformer
	indexer  cache.Indexer

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *PodController) Run(stopCh <-chan struct{}) error {
	// run starts and runs the shared informer
	go c.informer.Run(stopCh)

	// wait for the initial synchronization of the local cache.
	klog.Info("Waiting for informer caches to sync.")
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	return nil
}

func (c *PodController) RunWorker() {
	workFunc := func() bool {
		// Wait until there is a new item in the working queue
		key, shutdown := c.workqueue.Get()
		if shutdown {
			return false
		}
		defer c.workqueue.Done(key)

		obj, exists, err := c.indexer.GetByKey(key.(string))
		if err != nil {
			klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
			return false
		}

		if !exists {
			klog.Infof("Pod %s does not exist anymore", key)
			return false
		}

		klog.Infof("Sync/Add/Update for Pod %s, phase: %v", obj.(*corev1.Pod).GetName(), obj.(*corev1.Pod).Status.Phase)
		return true
	}

	for {
		if shutdown := workFunc(); shutdown {
			klog.Infof("pod worker shutting down.")
			return
		}
	}
}

func (c *PodController) List() error {
	// List lists all Pods in the indexer.
	podList, err := c.podLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, pod := range podList {
		klog.Infof("Got pod detail info: %s/%v/%v", pod.Namespace, pod.Name, pod.Status.Phase)
	}

	c.podLister.Pods("enigma2").Get("")

	return nil
}

func (c *PodController) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *PodController) enqueueForDelete(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *PodController) onAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)

	c.enqueue(obj)

	klog.Infof("POD CREATED: %s/%s", pod.Namespace, pod.Name)
}

func (c *PodController) onUpdate(old, new interface{}) {
	oldPod := old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}

	c.enqueue(new)

	klog.Infof(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

func (c *PodController) onDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)

	c.enqueueForDelete(obj)

	klog.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

func NewPodController(informerFactory informers.SharedInformerFactory) *PodController {
	// pod informer
	podInformer := informerFactory.Core().V1().Pods()
	// create informer
	informer := podInformer.Informer()
	// create pod lister
	podLister := podInformer.Lister()

	c := &PodController{
		informer:  informer,
		indexer:   informer.GetIndexer(),
		podLister: podLister,

		// create the workqueue
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods"),
	}

	klog.Info("Setting up custom resource event handlers.")
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Called on creation
		AddFunc: c.onAdd,
		// Called on resource update and every resyncPeriod on existing resources.
		UpdateFunc: c.onUpdate,
		// Called on resource deletion.
		DeleteFunc: c.onDelete,
	})

	return c
}
