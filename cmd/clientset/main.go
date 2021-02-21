package main

import (
	"fmt"
	"log"
	"os"

	"encoding/json"
	"net/http"

	"github.com/lqshow/access-kubernetes-cluster/pkg/kubernetes/client"
	"github.com/lqshow/access-kubernetes-cluster/service"
	"github.com/lqshow/access-kubernetes-cluster/version"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"

	clientsetexample "github.com/lqshow/access-kubernetes-cluster/pkg/clientset"
	kube "k8s.io/client-go/kubernetes"
)

const (
	// UserAgent is an optional field that specifies the caller of this request.
	UserAgent = "clientset-example"
)

func main() {
	showVersion := pflag.BoolP("version", "v", false, "Show version")

	pflag.Parse()
	if *showVersion {
		v, _ := json.MarshalIndent(version.Get(), "", "  ")
		fmt.Println(string(v))
		os.Exit(0)
	}

	// Load Config
	config := service.LoadConfigFromEnv()
	configModifier := func(c *rest.Config) {
		c.QPS = 5
		c.Burst = 10
		c.UserAgent = UserAgent
	}

	// crate kubernetes clientsets
	kubeClientSet, err := client.NewKubeClient("", config.KubeConfig, configModifier)
	if err != nil {
		zap.S().Fatalf("Failed to get kube client: %v", err)
	}

	r := initRouter(kubeClientSet, config)
	if err := r.Run(":3000"); err != nil {
		log.Fatalf("r.Run err: %v", err)
	}
}

func initRouter(clientset *kube.Clientset, config *service.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/ping", ping)
	r.GET("/k8s/pods", func(c *gin.Context) {
		clientsetExample := clientsetexample.NewPodExample(clientset, config, c)
		pods, err := clientsetExample.List()
		if err != nil {
			c.String(http.StatusInternalServerError, "GetPodList err: %v", err)
		}

		c.JSON(http.StatusOK, pods)
	})

	return r
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
