package clientset

import (
	"context"
	"k8s.io/klog/v2"

	"github.com/lqshow/access-kubernetes-cluster/service"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

type PodExample struct {
	clientset *kube.Clientset
	config    *service.Config
	ctx       context.Context
}

func NewPodExample(clientset *kube.Clientset, config *service.Config, ctx context.Context) *PodExample {
	return &PodExample{
		clientset: clientset,
		config:    config,
		ctx:       ctx,
	}
}

// specify namespace to get pods in particular namespace
func (c *PodExample) List() error {
	pods, err := c.clientset.CoreV1().Pods(c.config.KubeNamespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	klog.Infof("There are %d pods in the cluster.", len(pods.Items))

	for _, pod := range pods.Items {
		klog.Infof("Got pod name: %s/%v", pod.Namespace, pod.Name)
	}

	return nil
}
