package controller

import (
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
}

type indexController struct {
	container container.Container
	service   service.UserService
}

func (i indexController) ShowOperationRecord(c echo.Context) error {
	return c.Render(http.StatusOK, "operation_record.html", nil)
}

func (i indexController) ShowDatafiles(c echo.Context) error {
	return c.Render(http.StatusOK, "datafiles.html", nil)
}

func (i indexController) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}
func (i indexController) ShowStatus(c echo.Context) error {
	return c.Render(http.StatusOK, "status.html", nil)
}
func (i indexController) ShowBlockStatus(c echo.Context) error {
	return c.Render(http.StatusOK, "blockstatus.html", nil)
}
func NewIndexController(container container.Container) IndexController {
	return &indexController{
		container: container,
		service:   nil,
	}
}
