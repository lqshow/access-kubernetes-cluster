# Overview

How to access a kubernetes cluster

## Build Client

首先，作为和 Kubernetes 交互的应用程序，必须先构建一个 clientset。 clientset 是多个 client 的集合，每个 client 可能包含不同版本的方法调用。

clientset 的使用分两种情况：集群内和集群外。

**集群内**

假设自定义控制器是以 Pod 的方式运行在 Kubernetes 集群里的，只需调用 rest.InClusterConfig()。
这个控制器就会直接使用默认 ServiceAccount 数据卷里的授权信息，来访问 APIServer。

```go
import (
    "k8s.io/client-go/rest"
    kube "k8s.io/client-go/kubernetes"
)

// Creates the in-cluster config
config, err := rest.InClusterConfig()
if err != nil {
    panic(err.Error())
}

// Create the new Clientset
clientset, err := kube.NewForConfig(config)
if err != nil {
    panic(err.Error())
}
```

**集群外**

比如在本地，可以使用与 kubectl 一样的 kube-config 来配置 clients。
或者发生在一个集群需要访问另外一个集群的服务。

```go
import (
    "os"
    
    "k8s.io/client-go/tools/clientcmd"
    kube "k8s.io/client-go/kubernetes"
)

// Set the kubernetes config file path as environment variable
kubeconfig := os.Getenv("KUBECONFIG")

// Creates the out-of-cluster config
config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
if err != nil {
    panic(err.Error())
}

// Create the new Clientset
clientset, err := kube.NewForConfig(config)
if err != nil {
    panic(err.Error())
}
```

**定制封装 clientset**

```go
import (
    "k8s.io/client-go/rest"
    "github.com/lqshow/access-kubernetes-cluster/service"
)

const (
    // UserAgent is an optional field that specifies the caller of this request.
    UserAgent = "informer-example"
)

// Load Config
config := service.LoadConfigFromEnv()

configModifier := func(c *rest.Config) {
    c.QPS = 5
    c.Burst = 10
    c.UserAgent = UserAgent
}
kubeClientSet, err := client.NewKubeClient("", config.KubeConfig, configModifier)
if err != nil {
    zap.S().Fatalf("Failed to get kube client: %v", err)
}
```

## Clientset

Clientset 是 k8s 中出镜率最高的 client，用法比较简单。

1. 先选 group，比如 core. 
2. 再选具体的 resource，比如 pod. 
3. 最后再把动词（create、get、list）填上.

```go
pods, err := clientset.CoreV1().Pods(c.config.KubeNamespace).List(c.ctx, metav1.ListOptions{})

for _, pod := range pods.Items {
    klog.Infof("Got pod name: %s/%v", pod.Namespace, pod.Name)
}
```

## Dynamic client


## Informer

**enqueue key**

format: namespace/name

## References
- [Authenticating inside the cluster](https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/README.md)
- [Authenticating outside the cluster](https://github.com/kubernetes/client-go/blob/master/examples/out-of-cluster-client-configuration/README.md)
- [client-go Examples](https://github.com/kubernetes/client-go/blob/master/examples/README.md)
- [informer example code](https://github.com/feiskyer/kubernetes-handbook/blob/master/examples/client/informer/informer.go)
- [使用 client-go 控制原生及拓展的 Kubernetes API](https://studygolang.com/articles/9270)
- [kubernetes 中 informer 的使用](https://www.jianshu.com/p/1e2e686fe363)
- [Extend Kubernetes via a Shared Informer](https://gianarb.it/blog/kubernetes-shared-informer)