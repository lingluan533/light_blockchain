package consul

import (
	"embed"
	"sca_server/config"
	"testing"
)

//go:embed application.*.yml
var yamlFile embed.FS

func TestGetNumberOfServices(t *testing.T) {
	conf, _ := config.Load(yamlFile)
	GetNumberOfServices(conf)
}
