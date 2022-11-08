package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type EdgeNodeController interface {
	GetCountOfService(c echo.Context) error
}

type edgeNodeController struct {
	container container.Container
	service   service.EdgeNodeService
}

func (e edgeNodeController) GetCountOfService(c echo.Context) error {
	online, offline := e.service.GetServiceNumberOfOnlineAndOffline()
	var number map[string]int
	number["online"] = online
	number["offline"] = offline
	return c.JSON(http.StatusOK, number)
}

func NewEdgeNodeController(container container.Container) EdgeNodeController {
	return &edgeNodeController{
		container: container,
		service:   service.NewEdgeNodeService(container),
	}
}
