package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TensRoses/iris"
	"github.com/TensRoses/iris/internal/configs"
	"github.com/TensRoses/iris/internal/log"
)

const (
	defaultConfigPath = "./internal/configs"
)

// type metricsOptions struct {
// 	PrometheusMetrics  bool
// 	PrintMetrics       bool
// 	StackdriverMetrics bool
// 	StatsdMetrics      bool
// }

// depart all core run into internal.
func main() {
	logger := log.CreateLogger("tensrose")
	defer logger.Infof("--shutdown %s--", logger.Name)

	// parse configs and secrets parent directory since viper will handle configs
	cpath := flag.String("cpath", defaultConfigPath, fmt.Sprintf("Config path for storing default configs and secrets, default: %s", defaultConfigPath))
	// NOTE: this is when parsing options to get metrics from prom
	// var opts metricsOptions

	flag.Parse()

	// load configs and secrets
	cfg, err := configs.LoadConfigFile(*cpath)
	if err != nil {
		logger.Fatal(err)
	}

	// NOTE: possible caveats with multiple instance of viper
	// https://stackoverflow.com/a/47185439/8643197
	secret, err := configs.LoadSecretsFile(*cpath)
	if err != nil {
		logger.Warnf("%s, loading from ENV instead", err.Error())
		secret.AuthToken = os.Getenv("AUTH_TOKEN")
		secret.ClientID = os.Getenv("CLIENTID")
	}

	// setup metrics here
	// ....

	// Start Iris finally
	ir := iris.NewIris(*cfg, *secret, *logger)
	err = ir.Start()
	if err != nil {
		logger.Fatal(err)
	}
}
