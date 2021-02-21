package clientset

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/lqshow/access-kubernetes-cluster/service"

	corev1 "k8s.io/api/core/v1"
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
func (c *PodExample) List() ([]corev1.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(c.config.KubeNamespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	klog.Infof("There are %d pods in the cluster.", len(podList.Items))

	var pods []corev1.Pod
	for _, pod := range podList.Items {
		klog.Infof("Got pod name: %s/%v", pod.Namespace, pod.Name)

		pods = append(pods, pod)
	}

	return pods, nil
}
