package config

import (
	"embed"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// 业务状态码
const (
	EdgeNodeNotFound    = "10000"
	UserRepeatRegister  = "50001"
	UserRegisterSuccess = "50002"
	ServerInternalError = "50000"
	PasswordNotSame     = "50003"
	UserLoginSuccess    = "60002"
)

// 错误提示信息
const (
	MsgUserRepeatRegister  = "用户名已存在！"
	MsgUserRegisterSuccess = "用户注册成功！"
	MsgServerInternalError = "服务器内部错误，请联系管理员！"
	MsgPasswordNotSame     = "两次输入密码不一致，请重试!"
	MsgUserLoginSuccess    = "用户登录成功！"
)

type MysqlConfig struct {
	Dialect   string `default:"mysql"`
	Host      string
	Port      string
	Dbname    string
	Username  string
	Password  string
	Migration bool `default:"false"`
}
type RedisConfig struct {
	Network  string `yaml:"network"`
	Addr     string `yaml:"addr"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
	Pools    int    `yaml:"pools"`
	MinConns int    `yaml:"min_conns"`
}
type EMailConfig struct {
	Vendor   string `yaml:"vendor"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Sender   string `yaml:"sender"`
	Password string `yaml:"password"`
	Nice     string `yaml:"nice"`
	CC       string `yaml:"cc"`
}
type MessageConfig struct {
	Vendor       string `yaml:"vendor"`
	TokenUrl     string `yaml:"token_url"`
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Retry        int    `yaml:"retry"`
	Token        string `yaml:"token"`
	SendUrl      string `yaml:"send_url"`
	Tid          string `yaml:"tid"`
	Expires      int64  `yaml:"expires"`
}

type Extension struct {
	MasterGenerator bool `yaml:"master_generator" default:"false"`
	CorsEnabled     bool `yaml:"cors_enabled" default:"false"`
	SecurityEnabled bool `yaml:"security_enabled" default:"false"`
}
type Swagger struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type ConsulConfig struct {
	ConsulAddress    string `yaml:"consul_address"`
	ConsulPort       string `yaml:"consul_port"`
	ID               string `yaml:"ID"`
	Name             string `yaml:"Name"`
	LocalAddress     string `yaml:"LocalAddress"`
	LocalServicePort int    `yaml:"LocalServicePort"`
	HealthCheckID    string `yaml:"HealthCheckID"`
	HealthTCP        string `yaml:"HealthTCP"`
	HealthTimeout    string `yaml:"HealthTimeout"`
	HealthInterval   string `yaml:"HealthInterval"`
}

type LogConfig struct {
	RequestLogFormat string `yaml:"request_log_format" default:"${remote_ip} ${account_name} ${uri} ${method} ${status}"`
}
type EtcdPrefixConfig struct {
	UserOperation string `yaml:"user_operation"`
}
type BlockConfig struct {
	DataFileRootPath string `yaml:"root_path"`
}
type Config struct {
	MySql            MysqlConfig      `yaml:mysql`
	Redis            RedisConfig      `yaml:"redis"`
	EMail            EMailConfig      `yaml:"email"`
	Message          MessageConfig    `yaml:"message"`
	Extension        Extension        `yaml:"extension"`
	Swagger          Swagger          `yaml:"swagger"`
	Consul           ConsulConfig     `yaml:"ConsulConfig"`
	LogConfig        LogConfig        `yaml:"LogConfig"`
	EtcdPrefixConfig EtcdPrefixConfig `yaml:"EtcdKeyPrefix"`
	BlockConfig      BlockConfig      `yaml:"block"`
}

const (
	// DEV represents development environment
	DEV = "develop"
	// PRD represents production environment
	PRD = "production"
	// DOC represents docker container
	DOC = "docker"
)

// Load reads the settings written to the yml file
func Load(yamlFile embed.FS) (*Config, string) {
	var env *string
	if value := os.Getenv("WEB_APP_ENV"); value != "" {
		env = &value
	} else {
		env = flag.String("env", "develop", "To switch configurations.")
		flag.Parse()
	}

	file, err := yamlFile.ReadFile("application." + *env + ".yml")
	if err != nil {
		fmt.Printf("Failed to read application.%s.yml: %s", *env, err)
		os.Exit(2)
	}

	config := &Config{}
	if err := yaml.Unmarshal(file, config); err != nil {
		fmt.Printf("Failed to read application.%s.yml: %s", *env, err)
		os.Exit(2)
	}

	return config, *env
}
