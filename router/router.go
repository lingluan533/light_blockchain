package router

import (
	"fmt"
	"github.com/labstack/echo-contrib/session"
	"net/http"
	"sca_server/container"
	"sca_server/controller"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	echoSwagger "github.com/swaggo/echo-swagger"
	// for using echo-swagger
)

// Init initialize the routing of this application.
func Init(e *echo.Echo, container container.Container) {
	setCORSConfig(e, container)
	setUserController(e, container)
	setIndexController(e, container)
	setBlockChainController(e, container)
	setLoginController(e, container)
	setDataFileController(e, container)
	setEdgeNodeController(e, container)
	setSwagger(container, e)
}

func setEdgeNodeController(e *echo.Echo, container container.Container) {
	edgeNodeController := controller.NewEdgeNodeController(container)
	e.GET("/countOfServices", func(c echo.Context) error { return edgeNodeController.GetCountOfService(c) }, loginCheck)
	e.GET("/getClusterInfo", func(c echo.Context) error { return edgeNodeController.GetEdgeNodeRunStatus(c) }, loginCheck)
	e.POST("/helloEdgeNode", func(c echo.Context) error { return edgeNodeController.SayHelloEdgeNode(c) }, loginCheck)
	e.POST("/rebootEdgeNode", func(c echo.Context) error { return edgeNodeController.RebootEdgeNode(c) }, loginCheck)
}

func setCORSConfig(e *echo.Echo, container container.Container) {
	if container.GetConfig().Extension.CorsEnabled {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowCredentials: true,
			AllowOrigins:     []string{"*"},
			AllowHeaders: []string{
				echo.HeaderAccessControlAllowHeaders,
				echo.HeaderContentType,
				echo.HeaderContentLength,
				echo.HeaderAcceptEncoding,
			},
			AllowMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodDelete,
			},
			MaxAge: 86400,
		}))
	}
}

func setDataFileController(e *echo.Echo, container container.Container) {
	dataFileController := controller.NewDataFileController(container)
	e.GET("/allFiles", func(c echo.Context) error { return dataFileController.ShowDataFiles(c) }, loginCheck)
	e.POST("/downLoadFile", func(c echo.Context) error { return dataFileController.DownLoadFile(c) }, loginCheck)
	e.GET("/downLoad", func(c echo.Context) error { return dataFileController.DownLoad(c) }, loginCheck)
	e.GET("/countDataSize", func(c echo.Context) error { return dataFileController.GetTotalSizeOfDataFiles(c) }, loginCheck)

}
func setUserController(e *echo.Echo, container container.Container) {
	user := controller.NewUserController(container)
	e.GET("/getOperationRecords", func(c echo.Context) error { return user.GetOperationRecords(c) })
}
func goLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", nil)
}
func setIndexController(e *echo.Echo, container container.Container) {
	index := controller.NewIndexController(container)
	e.GET("/", func(c echo.Context) error { return index.Index(c) }, loginCheck)
	e.GET("/index.html", func(c echo.Context) error { return index.Index(c) }, loginCheck)
	e.GET("/status.html", func(c echo.Context) error { return index.ShowStatus(c) }, loginCheck)
	e.GET("/blockstatus.html", func(c echo.Context) error { return index.ShowBlockStatus(c) }, loginCheck)
	e.GET("/datafiles.html", func(c echo.Context) error { return index.ShowDatafiles(c) }, loginCheck)
	e.GET("/operation_record.html", func(c echo.Context) error { return index.ShowOperationRecord(c) }, loginCheck)
	e.GET("/register.html", func(c echo.Context) error { return index.ShowRegister(c) })
}

func setBlockChainController(e *echo.Echo, container container.Container) {
	controller := controller.NewBlockChainController(container)
	e.GET("blockchain/queryTimeReceipts", func(c echo.Context) error { return controller.QueryTimeReceipts(c) }, loginCheck)
	e.GET("blockchain/queryTimeTransaction", func(c echo.Context) error { return controller.QueryTimeTransaction(c) }, loginCheck)
	e.GET("blockchain/queryBlockInfos/:blockType", func(c echo.Context) error { return controller.QueryBlockInfos(c) }, loginCheck)

}
func setLoginController(e *echo.Echo, container container.Container) {
	controller := controller.NewUserController(container)
	e.GET("/login.html", func(c echo.Context) error { return controller.ShowLogin(c) })
	e.POST("/login", func(c echo.Context) error { return controller.Login(c) })
	e.GET("/logout", func(c echo.Context) error { return controller.LogOut(c) })
	e.POST("/register", func(c echo.Context) error {
		return controller.Register(c)
	})

}

func setSwagger(container container.Container, e *echo.Echo) {
	if container.GetConfig().Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler, loginCheck)
	}
}

func loginCheck(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 路由拦截 - 登录身份、资源权限判断等
		println("Api路由拦截：", c.Path())
		sess, _ := session.Get("user_session", c)
		if sess.IsNew {
			return goLogin(c)
		}
		//通过sess.Values读取会话数据
		username, err := sess.Values["user_name"].(string)
		if err == false || username == "" {
			fmt.Println("未登录，请登录后使用！")
			return goLogin(c)
		}
		isLogin := false
		if sess.Values["isLogin"] != nil {
			isLogin = sess.Values["isLogin"].(bool)
		}

		//打印会话数据
		fmt.Println(username)
		fmt.Println(isLogin)
		return next(c)
	}
}
