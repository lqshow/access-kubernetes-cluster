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

// PodController logs the name and namespace of pods that are added,
// deleted, or updated
type PodController struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
	podLister       corelisters.PodLister
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *PodController) Run(stopCh <-chan struct{}) error {
	// run starts and runs the shared informer
	go c.podInformer.Informer().Run(stopCh)

	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	return nil
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

	return nil
}

func (c *PodController) onAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)

	klog.Infof("POD CREATED: %s/%s", pod.Namespace, pod.Name)
}

func (c *PodController) onUpdate(old, new interface{}) {
	oldPod := old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)

	klog.Infof(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

func (c *PodController) onDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)

	klog.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

func NewPodController(informerFactory informers.SharedInformerFactory) *PodController {
	// pod informer
	podInformer := informerFactory.Core().V1().Pods()
	// create pod lister
	podLister := podInformer.Lister()

	c := &PodController{
		informerFactory: informerFactory,
		podInformer:     podInformer,
		podLister:       podLister,
	}

	// Your custom resource event handlers.
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Called on creation
		AddFunc: c.onAdd,
		// Called on resource update and every resyncPeriod on existing resources.
		UpdateFunc: c.onUpdate,
		// Called on resource deletion.
		DeleteFunc: c.onDelete,
	})

	return c
}
