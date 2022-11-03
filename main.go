package main

//https://echo.laily.net/
//https://www.tizi365.com/archives/28.html
//https://blog.csdn.net/darjun/article/details/119122806
//https://studygolang.com/articles/31491 golang&echo 开启HTTPS 服务

import (
	"embed"
	"html/template"
	"io"
	_ "net/http"
	"os"
	"path/filepath"
	"sca_server/config"
	"sca_server/container"
	"sca_server/logger"
	myMiddleware "sca_server/middleware"
	"sca_server/repository"
	"sca_server/router"
	"sca_server/session"
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

	//注册中间件
	//e.Use(middleware.Recover())
	//注册中间件
	//	Logfile, err := os.OpenFile("logs/sca_server.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)

	//if err != nil {
	//	log.Fatalf("打开日志文件失败：%s:%v\n", "logs/sca_server.log", err)
	//}
	//e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
	//	Format: "time=${time_rfc3339}, remote_ip=${remote_ip}, method=${method}, uri=${uri}, status=${status} latency=${latency}, bytes_in=${bytes_in},bytes_out=${bytes_out}\n",
	//	Output: io.MultiWriter(Logfile, os.Stderr),
	//}))
	//e.Logger.SetLevel(echolog.)
	//e.Logger.SetOutput(io.MultiWriter(Logfile, os.Stderr))

	conf, env := config.Load(yamlFile)
	logger := logger.InitLogger(env, zapYamlFile)
	logger.GetZapLogger().Infof("Loaded this configuration : application." + env + ".yml")
	rep := repository.NewBlockChainRepository(logger, conf)
	sess := session.NewSession()
	container := container.NewContainer(rep, sess, conf, logger, env)
	// 路由注册
	router.Init(e, container)
	myMiddleware.InitLoggerMiddleware(e, container)
	myMiddleware.StaticContentsMiddleware(e, container, staticFile)

	e.Logger.Fatal(e.Start(":8000"))
	defer rep.Close()

}
