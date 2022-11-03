package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type UserController interface {
	Login(c echo.Context) error
	Registe(c echo.Context) error
	ShowLogin(c echo.Context) error
}

type userController struct {
	container container.Container
	service   service.UserService
}

func (controller *userController) ShowLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", nil)
}

func (controller *userController) Login(c echo.Context) error {
	//获取登录请求参数
	username := c.FormValue("user")
	password := c.FormValue("password")
	remember := c.FormValue("remember")
	fmt.Println("login")
	fmt.Println(username)
	fmt.Println(remember)
	controller.service.LoginMethod(username, password)
	return c.Render(http.StatusOK, "index.html", nil)
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
