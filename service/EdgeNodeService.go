package service

import (
	"sca_server/consul"
	"sca_server/container"
)

type EdgeNodeService interface {
	GetServiceNumberOfOnlineAndOffline() (int, int)
}
type edgeNodeService struct {
	container container.Container
}

func (e edgeNodeService) GetServiceNumberOfOnlineAndOffline() (int, int) {
	online, offline := consul.GetNumberOfServices(e.container.GetConfig())
	return online, offline
}

func NewEdgeNodeService(container container.Container) EdgeNodeService {
	return &edgeNodeService{container: container}
}
