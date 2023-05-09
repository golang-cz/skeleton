package core

import (
	"fmt"

	"github.com/golang-cz/skeleton/config"
)

func SetupApp(conf *config.AppConfig, appName, version string) error {
	fmt.Printf("Appname: %v, Version %v", appName, version)
	fmt.Printf("%+v\n", conf)

	return nil
}
