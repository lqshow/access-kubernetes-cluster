package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Option func(c *rest.Config)

// NewKubeClient generates a kubernetes client by master URL and kube config.
func NewKubeClient(masterUrl, kubeconfigPath string, options ...Option) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return newFromConfig(config, options...)
}

// newFromConfig create a kubernetes client configuration
func newFromConfig(c *rest.Config, options ...Option) (*kubernetes.Clientset, error) {
	for _, option := range options {
		option(c)
	}

	return kubernetes.NewForConfig(c)
}
