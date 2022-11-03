package router

import (
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

	//setErrorController(e, container)
	setUserController(e, container)
	setIndexController(e, container)
	setBlockChainController(e, container)
	setLoginController(e, container)
	//setFormatController(e, container)
	//setAccountController(e, container)
	//setHealthController(e, container)

	setSwagger(container, e)
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

//func setErrorController(e *echo.Echo, container container.Container) {
//	errorHandler := controller.NewErrorController(container)
//	e.HTTPErrorHandler = errorHandler.JSONError
//	e.Use(middleware.Recover())
//}

func setUserController(e *echo.Echo, container container.Container) {
	user := controller.NewUserController(container)
	e.POST("/user/login", func(c echo.Context) error { return user.Login(c) })

}

func setIndexController(e *echo.Echo, container container.Container) {
	index := controller.NewIndexController(container)
	e.GET("/", func(c echo.Context) error { return index.Index(c) })
	e.GET("/status.html", func(c echo.Context) error { return index.ShowStatus(c) })
	e.GET("/blockstatus.html", func(c echo.Context) error { return index.ShowBlockStatus(c) })
}

func setBlockChainController(e *echo.Echo, container container.Container) {
	controller := controller.NewBlockChainController(container)
	e.GET("blockchain/queryTimeReceipts", func(c echo.Context) error { return controller.QueryTimeReceipts(c) })
	e.GET("blockchain/queryTimeTransaction", func(c echo.Context) error { return controller.QueryTimeTransaction(c) })

}
func setLoginController(e *echo.Echo, container container.Container) {
	controller := controller.NewUserController(container)
	e.GET("/login.html", func(c echo.Context) error { return controller.ShowLogin(c) })
	e.POST("/login", func(c echo.Context) error { return controller.Login(c) })

}

//func setCategoryController(e *echo.Echo, container container.Container) {
//	category := controller.NewCategoryController(container)
//	e.GET(controller.APICategories, func(c echo.Context) error { return category.GetCategoryList(c) })
//}
//
//func setFormatController(e *echo.Echo, container container.Container) {
//	format := controller.NewFormatController(container)
//	e.GET(controller.APIFormats, func(c echo.Context) error { return format.GetFormatList(c) })
//}
//
//func setAccountController(e *echo.Echo, container container.Container) {
//	account := controller.NewAccountController(container)
//	e.GET(controller.APIAccountLoginStatus, func(c echo.Context) error { return account.GetLoginStatus(c) })
//	e.GET(controller.APIAccountLoginAccount, func(c echo.Context) error { return account.GetLoginAccount(c) })
//
//	if container.GetConfig().Extension.SecurityEnabled {
//		e.POST(controller.APIAccountLogin, func(c echo.Context) error { return account.Login(c) })
//		e.POST(controller.APIAccountLogout, func(c echo.Context) error { return account.Logout(c) })
//	}
//}
//
//func setHealthController(e *echo.Echo, container container.Container) {
//	health := controller.NewHealthController(container)
//	e.GET(controller.APIHealth, func(c echo.Context) error { return health.GetHealthCheck(c) })
//}

func setSwagger(container container.Container, e *echo.Echo) {
	if container.GetConfig().Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}
