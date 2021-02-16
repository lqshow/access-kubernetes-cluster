package service

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	envPrefixKey          = "XDP_DATASETX_ENV_PREFIX"
	defaultEnvPrefixValue = "DATASETX"
)

type Config struct {
	Debug     bool   `default:"true" split_words:"true"`
	DevMode   bool   `default:"true" split_words:"true"`
	LogSource string `default:"XDP_DATASET" split_words:"true"`

	LogUnaryPayload  bool `default:"true" split_words:"true"`
	LogStreamPayload bool `default:"false" split_words:"true"`

	KubeConfig    string `default:"" envconfig:"KUBE_CONFIG"`
	KubeNamespace string `default:"" envconfig:"KUBE_NAMESPACE"`

	WorkerThreadiness int `default:"3" split_words:"true"`
}

func DefaultConfig() *Config {
	return &Config{}
}

func LoadConfigFromEnv(fileNames ...string) *Config {
	err := godotenv.Load(fileNames...)
	if err != nil {
		fmt.Println(".env config file not found, skip it")
	}

	envPrefix := defaultEnvPrefixValue
	prefix, exist := os.LookupEnv(envPrefixKey)
	if exist {
		envPrefix = prefix
	}

	var config Config
	err = envconfig.Process(envPrefix, &config)
	if err != nil {
		panic(err)
	}

	return &config
}
