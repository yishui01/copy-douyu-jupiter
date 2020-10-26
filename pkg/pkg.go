package pkg

import (
	"github.com/douyu/jupiter/pkg/constant"
	"os"
)

const jupiterVersion = "0.2.0"

var (
	startTime string
	goVersion string
)

var (
	appName         string
	appID           string
	hostName        string
	buildAppVersion string
	buildUser       string
	buildHost       string
	buildStatus     string
	buildTime       string
)

func init() {
	if appName == "" {
		appName = os.Getenv(constant.EnvAppHost)
	}
}
