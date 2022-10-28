package  main

import (
	"crypto/tls"
	"github.com/go-gomail/gomail"
	"math/rand"
	"strings"
	"time"
)

type EmailParam struct {
	// ServerHost 邮箱服务器地址，如腾讯企业邮箱为smtp.exmail.qq.com ,smtp.bupt.edu.cn
	ServerHost string
	// ServerPort 邮箱服务器端口，如腾讯企业邮箱为465, 465
	ServerPort int
	// FromEmail　发件人邮箱地址
	FromNice string //发件人昵称
	FromEmail string
	// FromPasswd 发件人邮箱密码（注意，这里是明文形式），TODO：如果设置成密文？
	FromPasswd string
	// Toers 接收者邮件，如有多个，则以英文逗号(“,”)隔开，不能为空
	Toers string
	// CCers 抄送者邮件，如有多个，则以英文逗号(“,”)隔开，可以为空
	CCers string
}

// 全局变量，因为发件人账号、密码，需要在发送时才指定
// 注意，由于是小写，外面的包无法使用
var serverHost, fromEmail, fromPasswd string
var serverPort int

var m *gomail.Message

func InitEmail(ep *EmailParam) {
	toers := []string{}

	serverHost = ep.ServerHost
	serverPort = ep.ServerPort
	fromEmail = ep.FromEmail
	fromPasswd = ep.FromPasswd

	m = gomail.NewMessage()

	if len(ep.Toers) == 0 {
		return
	}

	for _, tmp := range strings.Split(ep.Toers, ",") {
		toers = append(toers, strings.TrimSpace(tmp))
	}

	// 收件人可以有多个，故用此方式
	m.SetHeader("To", toers...)

	//抄送列表
	if len(ep.CCers) != 0 {
		for _, tmp := range strings.Split(ep.CCers, ",") {
			toers = append(toers, strings.TrimSpace(tmp))
		}
		m.SetHeader("Cc", toers...)
	}

	// 发件人
	// 第三个参数为发件人别名，如"李大锤"，可以为空（此时则为邮箱名称）
	m.SetAddressHeader("From", fromEmail, ep.FromNice)
}

// SendEmail body支持html格式字符串
func SendEmail(subject, body string) {
	// 主题
	m.SetHeader("Subject", subject)

	// 正文
	m.SetBody("text/html", body)

	d := gomail.NewPlainDialer(serverHost, serverPort, fromEmail, fromPasswd)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	// 发送
	err := d.DialAndSend(m)
	if err != nil {
		eLogger.Error(err)
	}
}

//用户使用邮箱号重置密码找回账户
func ResetEmail(user string, toers string) string {
	serverHost = CS.config.EMail.Host
	serverPort = CS.config.EMail.Port
	fromEmail = CS.config.EMail.Sender
	fromPasswd = CS.config.EMail.Password

	myToers := toers // 逗号隔开
	myCCers := CS.config.EMail.CC //""

	subject := "用户（"+user+"）发起的重置"
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	newpass := make([]rune, 6)
	r:=rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range newpass {
		newpass[i] = letters[r.Intn(62)]
	}
	body := "尊敬的用户，你好：<br>\n收到您的重置密码申请。<br>\n<h3>"+user+"的新密码：</h3>"+string(newpass)+"<br>Hello forgoter :(<a href='http://www.scope.org'>主页</a><br>"
	// 结构体赋值
	myEmail := &EmailParam {
		ServerHost: serverHost,
		ServerPort: serverPort,
		FromEmail:  fromEmail,
		FromNice:   CS.config.EMail.Nice,
		FromPasswd: fromPasswd,
		Toers:      myToers,
		CCers:      myCCers,
	}
	eLogger.Info("init reset email for "+myToers+".\n")
	InitEmail(myEmail)
	SendEmail(subject, body)
	return string(newpass)
}

//新用户注册，通知root用户和新用户，等待激活中状态
func RegisterEmail(user string, toers string, toroot string) {
	serverHost = CS.config.EMail.Host
	serverPort = CS.config.EMail.Port
	fromEmail = CS.config.EMail.Sender
	fromPasswd = CS.config.EMail.Password

	myToers := toers // 逗号隔开
	var  myCCers string
	if len(CS.config.EMail.CC) > 0 {
		myCCers = CS.config.EMail.CC+","+toroot //""
	} else {
		myCCers = toroot //""
	}
	subject := "新用户（"+user+"）发起的注册申请"

	body := "尊敬的用户，你好：<br>\n收到您的注册新用户申请。<br>\n<h3>管理员会尽快激活您注册的用户名为："+user+"的账户</h3><br>Hello new visitor :)<a href='http://www.scope.org'>主页</a><br>"
	// 结构体赋值
	myEmail := &EmailParam {
		ServerHost: serverHost,
		ServerPort: serverPort,
		FromEmail:  fromEmail,
		FromNice:   CS.config.EMail.Nice,
		FromPasswd: fromPasswd,
		Toers:      myToers,
		CCers:      myCCers,
	}
	eLogger.Info("init register email for:"+myToers+",CC:"+myCCers+".\n")
	InitEmail(myEmail)
	SendEmail(subject, body)
}