package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type IndexController interface {
	Index(c echo.Context) error
	ShowStatus(c echo.Context) error
	ShowBlockStatus(c echo.Context) error
	ShowDatafiles(c echo.Context) error
	ShowOperationRecord(c echo.Context) error
	ShowRegister(c echo.Context) error
}

type indexController struct {
	container container.Container
	service   service.UserService
}

func (i indexController) ShowRegister(c echo.Context) error {
	return c.Render(http.StatusOK, "register.html", nil)
}

func (i indexController) ShowOperationRecord(c echo.Context) error {
	return c.Render(http.StatusOK, "operation_record.html", UserData)
}

func (i indexController) ShowDatafiles(c echo.Context) error {
	return c.Render(http.StatusOK, "datafiles.html", UserData)
}

func (i indexController) Index(c echo.Context) error {
	fmt.Println("返回主页！")
	return c.Render(http.StatusOK, "index.html", UserData)
}
func (i indexController) ShowStatus(c echo.Context) error {
	return c.Render(http.StatusOK, "status.html", UserData)
}
func (i indexController) ShowBlockStatus(c echo.Context) error {
	return c.Render(http.StatusOK, "blockstatus.html", UserData)
}
func NewIndexController(container container.Container) IndexController {
	return &indexController{
		container: container,
		service:   nil,
	}
}
