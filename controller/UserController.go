package controller

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/config"
	"sca_server/container"
	"sca_server/model"
	"sca_server/service"
)

var UserData model.UserData

type UserController interface {
	Login(c echo.Context) error
	LogOut(c echo.Context) error
	Registe(c echo.Context) error
	ShowLogin(c echo.Context) error
	GetOperationRecords(c echo.Context) error
	Register(c echo.Context) error
}

type userController struct {
	container container.Container
	service   service.UserService
}

func (controller *userController) Register(c echo.Context) error {
	user_name := c.FormValue("user")
	user_phone := c.FormValue("phone")
	user_email := c.FormValue("email")
	user_pass0 := c.FormValue("pass0")
	user_pass1 := c.FormValue("pass1")
	//user_agree := c.FormValue("agree")
	fmt.Println("register")
	fmt.Println(user_name)

	fmt.Println(user_phone)
	fmt.Println(user_email)
	fmt.Println(user_pass0)
	fmt.Println(user_pass1)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "注册新用户成功，已通知管理员激活" + user_name
	if user_pass0 != user_pass1 {
		verify.Code = "00002"
		verify.Msg = "两次输入的新密码不同"
		return c.JSON(http.StatusOK, verify)
	}
	var userInfo model.UserInfo
	userInfo.UserName = user_name
	userInfo.Password = user_pass0
	userInfo.Phone = user_phone
	userInfo.Email = user_email
	res, err := controller.service.RegisterMethod(userInfo)
	if err != nil && err.Error() == config.UserRepeatRegister {
		verify.Code = config.UserRepeatRegister
		verify.Msg = config.MsgUserRepeatRegister
		return c.JSON(http.StatusOK, verify)
	}
	if res {
		verify.Code = config.UserRegisterSuccess
		verify.Msg = config.MsgUserRegisterSuccess
		return c.JSON(http.StatusOK, verify)
	}
	verify.Code = config.ServerInternalError
	verify.Msg = config.MsgServerInternalError
	return c.JSON(http.StatusOK, verify)
}

func (controller *userController) LogOut(c echo.Context) error {
	sess, _ := session.Get("user_session", c)
	//记录会话数据, sess.Values 是map类型，可以记录多个会话数据
	sess.Values["user_name"] = ""
	sess.Values["isLogin"] = false
	//保存用户会话数据
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "login.html")
}

func (controller *userController) GetOperationRecords(c echo.Context) error {
	records, err := controller.service.GetAllOperationRecordsByUserName("zms")
	if err != nil {
		fmt.Println("GetOperationRecords err=", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}
	//
	return c.JSONBlob(http.StatusOK, records)
}

func (controller *userController) ShowLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", nil)
}

type VerifyResponse struct {
	Code string `json:"code"` //00000
	Msg  string `json:"msg"`  //成功
	User string `json:"user"` //用户
	Url  string `json:"url"`  //跳转页面
}

func (controller *userController) Login(c echo.Context) error {
	//获取登录请求参数
	username := c.FormValue("user")
	password := c.FormValue("password")
	remember := c.FormValue("remember")
	fmt.Println("login")
	fmt.Println(username)
	fmt.Println(remember)
	var Verify VerifyResponse
	res, err := controller.service.LoginMethod(username, password)
	if err != nil {
		Verify.Code = "00001"
		Verify.Msg = "验证失败或已停用：" + username
		return c.JSON(http.StatusOK, Verify)
	}
	if res {
		//密码正确, 下面开始注册用户会话数据
		//以user_session作为会话名字，获取一个session对象
		sess, _ := session.Get("user_session", c)

		//设置会话参数
		sess.Options = &sessions.Options{
			Path:   "/",    //所有页面都可以访问会话数据
			MaxAge: 5 * 60, //会话有效期，单位秒
		}

		//记录会话数据, sess.Values 是map类型，可以记录多个会话数据
		sess.Values["user_name"] = username
		sess.Values["isLogin"] = true

		//保存用户会话数据
		sess.Save(c.Request(), c.Response())
		Verify.Code = "00000"
		Verify.Msg = "验证成功"
		Verify.User = username
		Verify.Url = "index.html"
		UserData.UserName = username
		return c.JSON(http.StatusOK, Verify)
	} else {
		Verify.Code = "00001"
		Verify.Msg = "验证失败或已停用：" + username
		return c.JSON(http.StatusOK, Verify)
	}
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
