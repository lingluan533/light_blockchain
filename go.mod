module sca_server

go 1.16

require (
	github.com/dchest/captcha v0.0.0-20200903113550-03f5f0333e1f
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-redis/redis/v8 v8.11.4
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/labstack/echo-contrib v0.11.0
	github.com/labstack/echo/v4 v4.9.0
	github.com/labstack/gommon v0.3.1
	github.com/mbobakov/grpc-consul-resolver v1.4.4
	github.com/swaggo/echo-swagger v1.3.5
	github.com/valyala/fasttemplate v1.2.1
	go.uber.org/zap v1.13.0
	google.golang.org/grpc v1.30.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/boj/redistore.v1 v1.0.0-20160128113310-fc113767cd6b
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.4.3
	gorm.io/driver/postgres v1.4.5 // indirect
	gorm.io/driver/sqlite v1.4.3 // indirect
	gorm.io/gorm v1.24.1-0.20221019064659-5dd2bb482755
)
