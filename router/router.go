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
	setDataFileController(e, container)
	setEdgeNodeController(e, container)
	//setFormatController(e, container)
	//setAccountController(e, container)
	//setHealthController(e, container)

	setSwagger(container, e)
}

func setEdgeNodeController(e *echo.Echo, container container.Container) {
	edgeNodeController := controller.NewEdgeNodeController(container)
	e.GET("/countOfServices", func(c echo.Context) error { return edgeNodeController.GetCountOfService(c) })
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
	e.GET("/allFiles", func(c echo.Context) error { return dataFileController.ShowDataFiles(c) })
	e.POST("/downLoadFile", func(c echo.Context) error { return dataFileController.DownLoadFile(c) })
	e.GET("/downLoad", func(c echo.Context) error { return dataFileController.DownLoad(c) })

}
func setUserController(e *echo.Echo, container container.Container) {
	user := controller.NewUserController(container)
	e.POST("/user/login", func(c echo.Context) error { return user.Login(c) })
	e.GET("/getOperationRecords", func(c echo.Context) error { return user.GetOperationRecords(c) })
}

func setIndexController(e *echo.Echo, container container.Container) {
	index := controller.NewIndexController(container)
	e.GET("/", func(c echo.Context) error { return index.Index(c) })
	e.GET("/index.html", func(c echo.Context) error { return index.Index(c) })
	e.GET("/status.html", func(c echo.Context) error { return index.ShowStatus(c) })
	e.GET("/blockstatus.html", func(c echo.Context) error { return index.ShowBlockStatus(c) })
	e.GET("/datafiles.html", func(c echo.Context) error { return index.ShowDatafiles(c) })
	e.GET("/operation_record.html", func(c echo.Context) error { return index.ShowOperationRecord(c) })
}

func setBlockChainController(e *echo.Echo, container container.Container) {
	controller := controller.NewBlockChainController(container)
	e.GET("blockchain/queryTimeReceipts", func(c echo.Context) error { return controller.QueryTimeReceipts(c) })
	e.GET("blockchain/queryTimeTransaction", func(c echo.Context) error { return controller.QueryTimeTransaction(c) })
	e.GET("blockchain/queryBlockInfos/:blockType", func(c echo.Context) error { return controller.QueryBlockInfos(c) })

}
func setLoginController(e *echo.Echo, container container.Container) {
	controller := controller.NewUserController(container)
	e.GET("/login.html", func(c echo.Context) error { return controller.ShowLogin(c) })
	e.POST("/login", func(c echo.Context) error { return controller.Login(c) })

}

func setSwagger(container container.Container, e *echo.Echo) {
	if container.GetConfig().Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}
