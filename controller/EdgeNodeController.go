package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type EdgeNodeController interface {
	GetCountOfService(c echo.Context) error
	GetEdgeNodeRunStatus(c echo.Context) error
	SayHelloEdgeNode(c echo.Context) error
	RebootEdgeNode(c echo.Context) error
}

type edgeNodeController struct {
	container container.Container
	service   service.EdgeNodeService
}

// TODO: 实现重启功能
func (e edgeNodeController) RebootEdgeNode(c echo.Context) error {
	e.service.RebootEdgeNodeMethod(c.FormValue("ipAddress"))
	return c.JSON(http.StatusOK, "")
}

// 节点探测信息
func (e edgeNodeController) SayHelloEdgeNode(c echo.Context) error {
	err := e.service.SayHelloEdgeNodeMethod(c.FormValue("ipAddress"))
	if err != nil {
		return c.JSON(http.StatusOK, "error")
	}
	return c.JSON(http.StatusOK, "Hello EdgeNode!")
}

func (e edgeNodeController) GetEdgeNodeRunStatus(c echo.Context) error {
	res := e.service.GetCLusterRunStatus()
	fmt.Println("最终结果数：", len(res))
	return c.JSON(http.StatusOK, res)
}

func (e edgeNodeController) GetCountOfService(c echo.Context) error {
	online, offline := e.service.GetServiceNumberOfOnlineAndOffline()
	var number = make(map[string]int)
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
