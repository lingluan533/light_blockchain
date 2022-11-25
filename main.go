package main

//https://echo.laily.net/
//https://www.tizi365.com/archives/28.html
//https://blog.csdn.net/darjun/article/details/119122806
//https://studygolang.com/articles/31491 golang&echo 开启HTTPS 服务

import (
	"embed"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"html/template"
	"io"
	_ "net/http"
	"os"
	"path/filepath"
	"sca_server/config"
	"sca_server/container"
	"sca_server/logger"
	myMiddleware "sca_server/middleware"
	"sca_server/mysessions"
	"sca_server/repository"
	"sca_server/router"

	// 这里一定要import;很重要
	_ "github.com/mbobakov/grpc-consul-resolver"
	//导入echo包
	"github.com/labstack/echo/v4"
)

//Implement echo.Renderer interface
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

//func customHTTPErrorHandler(err error, c echo.Context) {
//	code := http.StatusInternalServerError
//	fmt.Println(err)
//	eLogger.Error(err)
//	if he, ok := err.(*echo.HTTPError); ok {
//		code = he.Code
//	}
//	errorPage := fmt.Sprintf("%d.html", code)
//
//	c.Render(http.StatusOK, errorPage, CS.boxStatus)
//}

var Logfile *os.File
var eLogger echo.Logger

//go:embed application.*.yml
var yamlFile embed.FS

//go:embed zaplogger.*.yml
var zapYamlFile embed.FS

//go:embed statics/*
var staticFile embed.FS

func main() {

	Rundir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	os.Chdir(Rundir)

	//初始化echo实例
	e := echo.New()

	//Pre-compile templates
	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
	//Register templates
	e.Renderer = t
	// 使用session
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	conf, env := config.Load(yamlFile)
	logger := logger.InitLogger(env, zapYamlFile)
	logger.GetZapLogger().Infof("Loaded this configuration : application." + env + ".yml")
	rep := repository.NewBlockChainRepository(logger, conf)
	sess := mysessions.NewSession()
	container := container.NewContainer(rep, sess, conf, logger, env)
	// 路由注册
	router.Init(e, container)
	myMiddleware.InitLoggerMiddleware(e, container)
	myMiddleware.StaticContentsMiddleware(e, container, staticFile)

	e.Logger.Fatal(e.Start(":8000"))
	defer rep.Close()

}
