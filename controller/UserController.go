package controller

import (
	"github.com/labstack/echo/v4"
	"sca_server/container"
	"sca_server/service"
)

type UserController interface {
	Login(c echo.Context) error
	Registe(c echo.Context) error
}

type userController struct {
	container container.Container
	service   service.UserService
}

func (controller *userController) Login(c echo.Context) error {
	controller.service.LoginMethod()
	panic("implement me")
}

func (u userController) Registe(c echo.Context) error {
	panic("implement me")
}

func NewUserController(container container.Container) UserController {
	return &userController{
		container: container,
		service:   service.NewUserService(container),
	}
}
