package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type IndexController interface {
	Index(c echo.Context) error
}

type indexController struct {
	container container.Container
	service   service.UserService
}

func (i indexController) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func NewIndexController(container container.Container) IndexController {
	return &indexController{
		container: container,
		service:   nil,
	}
}
