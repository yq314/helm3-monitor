package pkg

import (
	"fmt"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"log"
	"os"
)

var (
	settings = cli.New()
)

func GetActionConfig(kubeConfig KubeConfig) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	// Add kube config settings passed by user
	settings.KubeConfig = kubeConfig.File
	settings.KubeContext = kubeConfig.Context

	err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), debug)
	if err != nil {
		return nil, err
	}

	return actionConfig, err
}

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Output(2, fmt.Sprintf(format, v...))
	}
}
