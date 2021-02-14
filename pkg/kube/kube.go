package kube

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lqshow/access-kubernetes-cluster/service"

	kube "k8s.io/client-go/kubernetes"
)

// Create the client configuration
func GetKubeConfig(config *service.Config) (*rest.Config, error) {
	// creates the in-cluster config
	if config.KubeConfig == "" {
		return rest.InClusterConfig()
	}
	// creates the out-of-cluster config
	return clientcmd.BuildConfigFromFlags("", config.KubeConfig)
}

// Create the new Clientset
func GetKubeClientset(config *service.Config) (*kube.Clientset, error) {
	kc, err := GetKubeConfig(config)
	if err != nil {
		panic(err)
	}
	return kube.NewForConfig(kc)
}
