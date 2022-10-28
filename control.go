package main

//https://blog.csdn.net/u010649766/article/details/79623972
//Golang redis 入门操作
//https://blog.csdn.net/QianLiStudent/article/details/103990921
//Golang之redis中间件框架——go-redis的使用

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

//登录页
func login(c echo.Context) error {
	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	if err != nil || data == 0 {
		eLogger.Errorf("no user,goto first.html")
		return c.Render(http.StatusOK, "first.html", CS.boxStatus)
	}
	return c.Render(http.StatusOK, "login.html", CS.boxStatus)
}

type VerifyResponse struct {
	Code string `json:"code"` //00000
	Msg  string `json:"msg"`  //成功
	User string `json:"user"` //用户
	Url  string `json:"url"`  //跳转页面
}
type SessionStatus struct {
	HostName   string
	UserName   string
	UserAvatar string
	Nice       string
}
type FileSessionStatus struct {
	*SessionStatus
	UserName      string //原地址
	Uri           string //目的地址
	OperationType string //操作类型 ADD:创建文件；MODIFY：修改文件。Copy：同步文件
	FileName      string
	TimeStamp     string
	FileSize      string
}

func login_user_check(user string, pass string) (bool, *SessionStatus) {
	ss := &SessionStatus{CS.boxStatus.HostName, "", "", ""}
	data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+user).Result()
	if err != nil {
		return false, ss
	}
	if len(data) == 0 {
		return false, ss
	}
	if data["name"] != user {
		return false, ss
	}
	if data["password"] != pass {
		return false, ss
	}
	if data["active"] != "true" {
		return false, ss
	}
	ss.Nice = data["nice"]
	ss.UserAvatar = data["avatar"]
	return true, ss
}
func passwd_check(user string, pass string) bool {
	data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+user).Result()
	if err != nil {
		return false
	}
	if len(data) == 0 {
		return false
	}
	if data["name"] != user {
		return false
	}
	if data["password"] != pass {
		return false
	}
	return true
}
func login_user(c echo.Context) error {
	//获取登录请求参数
	username := c.FormValue("user")
	password := c.FormValue("password")
	remember := c.FormValue("remember")
	fmt.Println("login")
	fmt.Println(username)
	fmt.Println(remember)

	var verify VerifyResponse
	//校验帐号密码是否正确
	isl, ss := login_user_check(username, password)
	if isl {
		//密码正确, 下面开始注册用户会话数据
		//以user_session作为会话名字，获取一个session对象
		sess, _ := session.Get("user_session", c)

		//设置会话参数
		sess.Options = &sessions.Options{
			Path:   "/",       //所有页面都可以访问会话数据
			MaxAge: 86400 * 7, //会话有效期，单位秒
		}

		//记录会话数据, sess.Values 是map类型，可以记录多个会话数据
		sess.Values["id"] = username
		sess.Values["isLogin"] = true
		sess.Values["avatar"] = ss.UserAvatar
		sess.Values["nice"] = ss.Nice
		ss.UserName = username

		if remember == "on" {
			sess.Values["remember"] = true
		} else {
			sess.Values["remember"] = false
		}

		//保存用户会话数据
		sess.Save(c.Request(), c.Response())
		verify.Code = "00000"
		verify.Msg = "验证成功"
		verify.User = username
		verify.Url = "index.html"
		return c.JSON(http.StatusOK, verify)

	} else {
		verify.Code = "00001"
		verify.Msg = "验证失败或已停用：" + username
		return c.JSON(http.StatusOK, verify)
	}

}
func login_hash(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", CS.boxStatus)
}
func login_messages(c echo.Context) error {
	//获取登录请求参数
	mobile := c.FormValue("mobile")
	code := c.FormValue("code")
	message := c.FormValue("message")
	rememberp := c.FormValue("rememberp")
	fmt.Println("login")
	fmt.Println(mobile)
	fmt.Println(code)
	//fmt.Println(message)
	fmt.Println(rememberp)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "验证成功"
	//取mobile对应的user_name
	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	if err != nil || data == 0 {
		verify.Code = "00404"
		verify.Msg = "取记录数失败"
		return c.JSON(http.StatusOK, verify)
	}
	op := &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf", //Offset,Count
	}
	ret, err := CS.redisClient.ZRangeByScoreWithScores(CS.ctx, "userid", op).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "取记录集合失败"
		return c.JSON(http.StatusOK, verify)
	}
	for _, z := range ret {
		fmt.Println(z.Member, z.Score)
		data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+z.Member.(string)).Result()
		if err != nil || len(data) == 0 {
			continue
		}
		if data["phone"] == mobile {
			//检查短信和失效短信接口
			user_name := data["name"]
			lasttimestamp, _ := strconv.ParseInt(data["lasttimestamp"], 10, 64)
			fmt.Println(user_name)
			fmt.Println(lasttimestamp)

			if time.Now().Unix()-lasttimestamp < CS.config.Message.Expires && code == data["code"] && message == data["message"] && data["active"] == "true" {
				//建会话信息
				//以user_session作为会话名字，获取一个session对象
				sess, _ := session.Get("user_session", c)

				//设置会话参数
				sess.Options = &sessions.Options{
					Path:   "/",       //所有页面都可以访问会话数据
					MaxAge: 86400 * 7, //会话有效期，单位秒
				}

				//记录会话数据, sess.Values 是map类型，可以记录多个会话数据
				sess.Values["id"] = user_name
				sess.Values["mobile"] = mobile
				sess.Values["isLogin"] = true
				sess.Values["avatar"] = data["avatar"]

				if rememberp == "on" {
					sess.Values["rememberp"] = true
				} else {
					sess.Values["rememberp"] = false
				}

				//保存用户会话数据
				sess.Save(c.Request(), c.Response())
				verify.Url = "index.html"
				verify.User = user_name
				return c.JSON(http.StatusOK, verify)
			}
			break
		}
	}

	verify.Code = "00001"
	verify.Msg = "验证码与短信验证失败，过期或已停用：" + mobile

	return c.JSON(http.StatusOK, verify)
}
func login_recover(c echo.Context) error {
	//获取登录请求参数
	email := c.FormValue("email")
	fmt.Println("recover")
	fmt.Println(email)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "找回用户和新密码成功，已邮件通知你" + email
	op := &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf", //Offset,Count
	}
	ret, err := CS.redisClient.ZRangeByScoreWithScores(CS.ctx, "userid", op).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "取记录集合失败"
		return c.JSON(http.StatusOK, verify)
	}
	for _, z := range ret {
		fmt.Println(z.Member, z.Score)
		data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+z.Member.(string)).Result()
		if err != nil || len(data) == 0 {
			continue
		}
		if data["email"] == email {
			newpass := ResetEmail(data["name"], email)
			CS.redisClient.HSet(CS.ctx, "user_"+data["name"], "password", newpass)
			return c.JSON(http.StatusOK, verify)
		}
	}
	CS.redisClient.Save(CS.ctx)
	//防暴力破解
	verify.Code = "00404"
	verify.Msg = "取邮件记录集合失败"
	return c.JSON(http.StatusOK, verify)
}
func login_register(c echo.Context) error {
	user_name := c.FormValue("user")
	user_nice := user_name
	user_phone := c.FormValue("phone")
	user_email := c.FormValue("email")
	user_pass0 := c.FormValue("pass0")
	user_pass1 := c.FormValue("pass1")
	user_avatar := c.FormValue("avatar")
	user_agree := c.FormValue("agree")
	fmt.Println("register")
	fmt.Println(user_name)
	fmt.Println(user_nice)
	fmt.Println(user_phone)
	fmt.Println(user_email)
	//fmt.Println(user_pass0)
	//fmt.Println(user_pass1)
	fmt.Println(user_avatar)
	fmt.Println(user_agree)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "注册新用户成功，已通知管理员激活" + user_name
	if user_pass0 != user_pass1 {
		verify.Code = "00002"
		verify.Msg = "两次输入的新密码不同"
		return c.JSON(http.StatusOK, verify)
	}

	data, err := CS.redisClient.HGetAll(CS.ctx, "user_root").Result()
	if err != nil || len(data) == 0 {
		verify.Code = "00500"
		verify.Msg = "检查操作失败"
		return c.JSON(http.StatusOK, verify)
	}
	toroot := data["email"]
	data, err = CS.redisClient.HGetAll(CS.ctx, "user_"+user_name).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "检查操作失败"
		return c.JSON(http.StatusOK, verify)
	}
	if len(data) > 0 {
		verify.Code = "00002"
		verify.Msg = "用户已经存在"
		return c.JSON(http.StatusOK, verify)
	}

	op := &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}
	ret, err := CS.redisClient.ZRevRangeByScoreWithScores(CS.ctx, "userid", op).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "取记录集合失败"
		return c.JSON(http.StatusOK, verify)
	}
	z := ret[0]
	fmt.Println(z.Member, z.Score)
	newid := float64(int64(z.Score) + 1)
	member1 := &redis.Z{newid, user_name}
	err = CS.redisClient.ZAdd(CS.ctx, "userid", member1).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库增加记录失败" + strconv.FormatFloat(newid, 'E', -1, 64)
		return c.JSON(http.StatusOK, verify)
	}
	err = CS.redisClient.HMSet(CS.ctx, "user_"+user_name, "id", int64(newid), "name", user_name, "nice", user_nice, "phone", user_phone, "email", user_email, "password", user_pass1, "avatar", user_avatar, "active", "false").Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库增加记录失败" + user_name
		err = CS.redisClient.ZRem(CS.ctx, "userid", user_name).Err()
		return c.JSON(http.StatusOK, verify)
	}
	RegisterEmail(user_name, user_email, toroot)
	return c.JSON(http.StatusOK, verify)
}

//验证码
type CaptchaResponse struct {
	Code      string `json:"code"`      //00000
	Msg       string `json:"msg"`       //成功
	CaptchaId string `json:"captchaId"` //验证码Id
	ImageUrl  string `json:"imageUrl"`  //验证码图片url
	Id        string `json:"user"`      //记住用户名
	Remember  bool   `json:"remember"`  //是否记住用户
	Mobile    string `json:"mobile"`    //记住手机号
	Rememberp bool   `json:"rememberp"` //是否记住手机
}

func captcha_id(c echo.Context) error {
	length := 4 //c.DefaultLen
	captchaId := captcha.NewLen(length)
	var captcha CaptchaResponse
	captcha.Code = "00000"
	captcha.Msg = "成功"
	captcha.CaptchaId = captchaId
	captcha.ImageUrl = "/captcha/" + captchaId + ".png"
	sess, _ := session.Get("user_session", c)
	if !sess.IsNew {
		captcha.Id = sess.Values["id"].(string)
		if sess.Values["remember"] != nil {
			captcha.Remember = sess.Values["remember"].(bool)
		}
		if sess.Values["mobile"] != nil {
			captcha.Mobile = sess.Values["mobile"].(string)
		}
		if sess.Values["rememberp"] != nil {
			captcha.Rememberp = sess.Values["rememberp"].(bool)
		}
	}
	return c.JSON(http.StatusOK, captcha)
}

func captcha_png(c echo.Context) error {
	captchaId := c.Param("captchaId")
	fmt.Println("GetCaptchaPng : " + captchaId)
	ServeHTTP(c.Response().Writer, c.Request())
	return nil
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir, file := path.Split(r.URL.Path)
	ext := path.Ext(file)
	id := file[:len(file)-len(ext)]
	fmt.Println("file : " + file)
	fmt.Println("ext : " + ext)
	fmt.Println("id : " + id)
	if ext == "" || id == "" {
		http.NotFound(w, r)
		return
	}
	fmt.Println("reload : " + r.FormValue("reload"))
	if r.FormValue("reload") != "" {
		captcha.Reload(id)
	}
	lang := strings.ToLower(r.FormValue("lang"))
	download := path.Base(dir) == "download"
	if Serve(w, r, id, ext, lang, download, captcha.StdWidth, captcha.StdHeight) == captcha.ErrNotFound {
		http.NotFound(w, r)
	}
}

func Serve(w http.ResponseWriter, r *http.Request, id, ext, lang string, download bool, width, height int) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	var content bytes.Buffer
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
		captcha.WriteImage(&content, id, width, height)
	case ".wav":
		w.Header().Set("Content-Type", "audio/x-wav")
		captcha.WriteAudio(&content, id, lang)
	default:
		return captcha.ErrNotFound
	}

	if download {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	http.ServeContent(w, r, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
	return nil
}

func captcha_verify(c echo.Context) error {
	captchaId := c.Param("captchaId")
	value := c.Param("value")
	mobile := c.Param("mobile")
	if captchaId == "" || value == "" {
		c.String(http.StatusBadRequest, "参数错误")
	}
	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "验证成功"
	if captcha.VerifyString(captchaId, value) {
		//查询手机号是注册用户的
		data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
		if err != nil || data == 0 {
			verify.Code = "00404"
			verify.Msg = "取记录数失败"
			return c.JSON(http.StatusOK, verify)
		}
		op := &redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf", //Offset,Count
		}
		ret, err := CS.redisClient.ZRangeByScoreWithScores(CS.ctx, "userid", op).Result()
		if err != nil {
			verify.Code = "00500"
			verify.Msg = "取记录集合失败"
			return c.JSON(http.StatusOK, verify)
		}
		for _, z := range ret {
			fmt.Println(z.Member, z.Score)
			data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+z.Member.(string)).Result()
			if err != nil || len(data) == 0 {
				continue
			}
			if data["phone"] == mobile {
				//调用发短信接口
				user_name := data["name"]
				message := SendSms(mobile, CS.boxStatus.HostName)
				if len(message) > 0 {
					//登记用户user_xxx表code/message/lasttimestamp
					err = CS.redisClient.HMSet(CS.ctx, "user_"+user_name, "code", value, "message", message, "lasttimestamp", strconv.FormatInt(time.Now().Unix(), 10)).Err()
					if err != nil {
						verify.Code = "00003"
						verify.Msg = "数据库记录保存短消息失败" + mobile
						return c.JSON(http.StatusOK, verify)
					}
					return c.JSON(http.StatusOK, verify)
				}
				break
			}
		}

		verify.Code = "00002"
		verify.Msg = "发短消息失败：" + mobile
		return c.JSON(http.StatusOK, verify)
	} else {
		verify.Code = "00001"
		verify.Msg = "验证失败"
		return c.JSON(http.StatusOK, verify)
	}
}

//产品介绍页
func introduct(c echo.Context) error {
	return c.Render(http.StatusOK, "introduct.html", CS.boxStatus)
}

//忘记密码页
func forgot(c echo.Context) error {
	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	if err != nil || data == 0 {
		eLogger.Errorf("no user,goto first.html")
		return c.Render(http.StatusOK, "first.html", CS.boxStatus)
	}
	return c.Render(http.StatusOK, "forgot-password.html", CS.boxStatus)
}

//注册新用户页
func register(c echo.Context) error {
	return c.Render(http.StatusOK, "register.html", CS.boxStatus)
}

var UserName string

//管理服务，都需要检查登录状态，用户种类、权限等访问控制
func index_account(c echo.Context, url string) (bool, *SessionStatus) {
	ss := &SessionStatus{CS.boxStatus.HostName, "", "", ""}
	sess, _ := session.Get("user_session", c)

	if sess.IsNew {
		return false, ss
	}
	//通过sess.Values读取会话数据
	username := sess.Values["id"].(string)
	ss.UserName = username
	if sess.Values["avatar"] != nil {
		ss.UserAvatar = sess.Values["avatar"].(string)
	}
	if sess.Values["nice"] != nil {
		ss.Nice = sess.Values["nice"].(string)
	}

	isLogin := false
	if sess.Values["isLogin"] != nil {
		isLogin = sess.Values["isLogin"].(bool)
	}

	//打印会话数据
	fmt.Println(url)
	fmt.Println(username)
	fmt.Println(isLogin)

	//停用和删除用户后，应该设置session失效
	return isLogin, ss
}

//管理首页
func index(c echo.Context) error {
	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	if err != nil || data == 0 {
		eLogger.Errorf("no user,goto first.html")
		return c.Render(http.StatusOK, "first.html", CS.boxStatus)
	}

	isi, ss := index_account(c, "index")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	return c.Render(http.StatusOK, "index.html", ss)
}
func userInit(c echo.Context) error {
	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	if err != nil || data != 0 {
		eLogger.Errorf("have user, goto login.html")
		bool1, _ := index_account(c, "index")
		if !bool1 {
			return c.Redirect(http.StatusFound, "login.html")
		}
		return c.Render(http.StatusOK, "index.html", CS.boxStatus)
	}

	return c.Render(http.StatusOK, "first.html", CS.boxStatus)
}
func logout(c echo.Context) error {
	fmt.Println("logout")

	sess, _ := session.Get("user_session", c)
	if sess.IsNew {
		return c.Redirect(http.StatusFound, "login.html")
	}

	remember := sess.Values["remember"]
	rememberp := sess.Values["rememberp"]
	fmt.Println(remember)
	fmt.Println(rememberp)
	if remember == nil {
		sess.Values["id"] = ""
	} else if !sess.Values["remember"].(bool) {
		sess.Values["id"] = ""
	}
	if rememberp == nil {
		sess.Values["mobile"] = ""
	} else if !sess.Values["rememberp"].(bool) {
		sess.Values["mobile"] = ""
	}
	sess.Values["isLogin"] = false
	//保存用户会话数据
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "login.html")
}

//盒子管理
func products(c echo.Context) error {

	isi, ss := index_account(c, "products")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	return c.Render(http.StatusOK, "products.html", ss)
}

//状态查询
func status(c echo.Context) error {

	isi, ss := index_account(c, "products")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	ss2 := &FileSessionStatus{}
	ss2.SessionStatus = ss
	data, _ := CS.redisClient.HGetAll(CS.ctx, "rclone").Result()
	ss2.FileName = data["filename"]
	ss2.Uri = data["path"]
	ss2.OperationType = data["type"]
	ss2.TimeStamp = data["time"]
	//ss3 := []*FileSessionStatus{}
	//ss3 = append(ss3,ss2)
	return c.Render(http.StatusOK, "status.html", ss2)
}

//区块状态查询
func bockStatus(c echo.Context) error {

	isi, ss := index_account(c, "products")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	return c.Render(http.StatusOK, "blockstatus.html", ss)
}

//用户管理
func accounts(c echo.Context) error {
	isi, ss := index_account(c, "accounts")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	return c.Render(http.StatusOK, "accounts.html", ss)
}

func accounts_add_user(c echo.Context) error {
	isi, ss := index_account(c, "accounts_add_user")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	return c.Render(http.StatusOK, "add-user.html", ss)
}

func accounts_username_validate(c echo.Context) error {
	user_name := c.FormValue("user")
	fmt.Println("accounts_username_validate")
	fmt.Println(user_name)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "用户" + user_name + "已存在"
	score, err := CS.redisClient.ZScore(CS.ctx, "userid", user_name).Result()
	if err != nil || score <= 0 {
		verify.Code = "00404"
		verify.Msg = "数据库未查询到用户：" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	return c.JSON(http.StatusOK, verify)
}
func accounts_new(c echo.Context) error {
	isi, _ := index_account(c, "accounts_new")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	user_name := c.FormValue("user")
	user_nice := c.FormValue("nice")
	user_phone := c.FormValue("phone")
	user_email := c.FormValue("email")
	user_pass0 := c.FormValue("pass0")
	user_pass1 := c.FormValue("pass1")
	user_avatar := c.FormValue("avatar")
	fmt.Println(user_name)
	fmt.Println(user_nice)
	fmt.Println(user_phone)
	fmt.Println(user_email)
	//fmt.Println(user_pass0)
	//fmt.Println(user_pass1)
	fmt.Println(user_avatar)
	file, err := os.Open(user_avatar)
	if err != nil {
		fmt.Println(err)
	}
	fi, _ := file.Stat()
	perm := fi.Mode()
	desFile, err := os.OpenFile("upload/"+user_name+path.Ext(user_avatar), os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	//dst, err := os.Create("./upload/abc"+path.Ext(file.Filename))
	if err != nil {
		fmt.Println(err)
	}
	defer desFile.Close()
	io.Copy(desFile, file)

	file.Close()
	if user_avatar != "upload/profile-image.png" {
		os.Remove(user_avatar)
	}

	user_avatar = desFile.Name()

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "增加用户成功" + user_name
	if user_pass0 != user_pass1 {
		verify.Code = "00002"
		verify.Msg = "两次输入的新密码不同"
		return c.JSON(http.StatusOK, verify)
	}

	data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+user_name).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "检查操作失败"
		return c.JSON(http.StatusOK, verify)
	}
	if len(data) > 0 {
		verify.Code = "00002"
		verify.Msg = "用户已经存在"
		return c.JSON(http.StatusOK, verify)
	}

	op := &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}
	ret, err := CS.redisClient.ZRevRangeByScoreWithScores(CS.ctx, "userid", op).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "取记录集合失败"
		return c.JSON(http.StatusOK, verify)
	}
	z := ret[0]
	fmt.Println(z.Member, z.Score)
	newid := float64(int64(z.Score) + 1)
	member1 := &redis.Z{newid, user_name}
	err = CS.redisClient.ZAdd(CS.ctx, "userid", member1).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库增加记录失败" + strconv.FormatFloat(newid, 'E', -1, 64)
		return c.JSON(http.StatusOK, verify)
	}
	err = CS.redisClient.HMSet(CS.ctx, "user_"+user_name, "id", int64(newid), "name",
		user_name, "nice", user_nice, "phone", user_phone, "email", user_email,
		"password", user_pass1, "avatar", user_avatar, "active", "false").Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库增加记录失败" + user_name
		err = CS.redisClient.ZRem(CS.ctx, "userid", user_name).Err()
		return c.JSON(http.StatusOK, verify)
	}
	CS.redisClient.Save(CS.ctx)

	//adduser
	cmd := exec.Command("sh", "-c", "nsenter --mount=/hostip/1/ns/mnt --net=/hostip/1/ns/net -- useradd -d /home/"+user_name+" -s /bin/ash -m "+user_name)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	//eLogger.Info(cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	eLogger.Infof(out.String())

	//passwd命令， 更新/etc/passwd和/etc/shadow
	cmd = exec.Command("sh", "-c", "echo -e '"+user_pass0+"\n"+user_pass1+"'|nsenter --mount=/hostip/1/ns/mnt --net=/hostip/1/ns/net -- passwd "+user_name)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()

	//htpasswd命令 -b -D/ -b， 更新/opt/nginx/conf/htpasswd
	cmd = exec.Command("htpasswd", "-b", "/root/scas/conf/htpasswd.dat", user_name, user_pass1)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()

	verify.Url = "accounts.html"
	return c.JSON(http.StatusOK, verify)
}
func first_new(c echo.Context) error {
	user_name := c.FormValue("user")
	user_nice := c.FormValue("nice")
	user_phone := c.FormValue("phone")
	user_email := c.FormValue("email")
	user_pass0 := c.FormValue("pass0")
	user_ip := c.FormValue("ip")
	user_hash := c.FormValue("hash")
	user_redis := c.FormValue("redis")

	user_avatar := c.FormValue("avatar")
	fmt.Println(user_name)
	fmt.Println(user_nice)
	fmt.Println(user_phone)
	fmt.Println(user_email)
	//fmt.Println(user_pass0)
	//fmt.Println(user_pass1)
	fmt.Println(user_ip)
	fmt.Println(user_hash)
	fmt.Println(user_avatar)
	fmt.Println(c.FormValue("redis_port"))
	fmt.Println(c.FormValue("redis"))
	fmt.Println(c.FormValue("token_url"))
	fmt.Println(c.FormValue("host"))
	fmt.Println(c.FormValue("sender"))
	fmt.Println(c.FormValue("email_password"))

	var redisConfig RedisConfig = RedisConfig{
		"tcp",
		user_redis, //这里要改成user_redis
		"6379",
		"password",
		0,
		16,
		0,
	}
	var emailConfig EMailConfig = EMailConfig{
		Vendor:   "bupt",
		Host:     "smtp.bupt.edu.cn",
		Port:     465,
		Sender:   "roottest@bupt.edu.cn",
		Password: "roottest",
		Nice:     user_nice,
		CC:       c.FormValue("email_cc"),
	}
	var messageConfig MessageConfig = MessageConfig{
		Vendor:       "easemob",
		TokenUrl:     "http://a1.easemob.com/1113211109110751/scope/token",
		ClientId:     "YXA6GfgasKrsSKSfCIVa7SSLNg",
		ClientSecret: "YXA6NnCMnQL8d4GtSnLVggcwx8P8HG8",
		Retry:        86400,
		Token:        "",
		SendUrl:      "http://a1.easemob.com/1113211109110751/scope/sms/send",
		Tid:          "802",
		Expires:      120,
	}
	var globle GlobalConfig = GlobalConfig{
		Redis:   redisConfig,
		EMail:   emailConfig,
		Message: messageConfig,
	}
	SetConfig(globle)
	//InitRedisStore()
	//user_ip设置conf文件的网段，默认为192.168.216.205，src是/etc/config/zero/local.conf
	WriteToJson("./conf/local.conf", user_ip)
	WriteRc()
	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "增加用户成功" + user_name

	data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+user_name).Result()
	if err != nil {
		verify.Code = "00500"
		verify.Msg = "检查操作失败"
		return c.JSON(http.StatusOK, verify)
	}
	if len(data) > 0 {
		verify.Code = "00002"
		verify.Msg = "用户已经存在"
		return c.JSON(http.StatusOK, verify)
	}

	var newid float64
	newid = 1
	member1 := &redis.Z{newid, user_name}
	err = CS.redisClient.ZAdd(CS.ctx, "userid", member1).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库增加记录失败" + strconv.FormatFloat(newid, 'E', -1, 64)
		return c.JSON(http.StatusOK, verify)
	}
	err = CS.redisClient.HMSet(CS.ctx, "user_"+user_name, "id", int64(newid), "name", user_name, "nice", user_nice, "phone", user_phone, "email", user_email, "password", user_pass0, "avatar", user_avatar, "active", "true").Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库增加记录失败" + user_name
		err = CS.redisClient.ZRem(CS.ctx, "userid", user_name).Err()
		return c.JSON(http.StatusOK, verify)
	}
	CS.redisClient.Save(CS.ctx)

	//adduser
	cmd := exec.Command("adduser", "-h", "/home/"+user_name, "-D", user_name)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	eLogger.Infof(out.String())

	//passwd命令， 更新/etc/passwd和/etc/shadow
	cmd = exec.Command("sh", "-c", "echo -e '"+user_pass0+"\n"+user_pass0+"'|nsenter --mount=/hostip/1/ns/mnt --net=/hostip/1/ns/net -- passwd "+user_name)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()

	//htpasswd命令 -b -D/ -b， 更新/opt/nginx/conf/htpasswd
	cmd = exec.Command("htpasswd", "-bc", "/root/scas/conf/htpasswd.dat", user_name, user_pass0)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()

	verify.Url = "login.html"
	return c.JSON(http.StatusOK, verify)
}
func accounts_newimg(c echo.Context) error {
	// Read form fields
	//name := c.FormValue("name")
	//email := c.FormValue("email")
	isi, _ := index_account(c, "accounts_newimg")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	// Source

	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	fmt.Print(data)
	if err != nil || data == 0 {
		eLogger.Errorf("no user,goto first.html")
		return c.Render(http.StatusOK, "first.html", CS.boxStatus)
	}
	sess, _ := session.Get("user_session", c)

	if sess.IsNew {
		return err
	}
	//通过sess.Values读取会话数据
	username := sess.Values["id"].(string)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.User = username

	file, err := c.FormFile("file")
	fmt.Println("文件名：", file.Filename)
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	rand.Seed(time.Now().UnixNano())
	data2 := rand.Int63n(100)
	str := fmt.Sprint(data2)
	// Destination
	dst, err := os.Create("upload/" + str + path.Ext(file.Filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	verify.Msg = "增加头像成功"
	verify.Url = dst.Name()

	return c.JSON(http.StatusOK, verify)
	//return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully with fields </p>", file.Filename))
}
func accounts_updateimg(c echo.Context) error {
	isi, _ := index_account(c, "accounts_updateimg")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	//获取登录请求参数
	user_name := c.FormValue("username")

	fmt.Println(user_name)
	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "更新用户" + user_name + "头像成功"

	file, err := c.FormFile("file")
	fmt.Println("文件名：", file.Filename)
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "上传文件失败"
		return c.JSON(http.StatusOK, verify)
	}

	src, err := file.Open()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "上传文件失败"
		return c.JSON(http.StatusOK, verify)
	}
	defer src.Close()
	rand.Seed(time.Now().UnixNano())
	data2 := rand.Int63n(100)
	str := fmt.Sprint(data2)
	// Destination
	dst, err := os.Create("upload/" + str + path.Ext(file.Filename))
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "上传文件失败"
		return c.JSON(http.StatusOK, verify)
	}
	defer dst.Close()
	user_avatar := dst.Name()
	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		verify.Code = "00003"
		verify.Msg = "上传文件失败"
		return c.JSON(http.StatusOK, verify)
	}
	//CS.redisClient.HSet(CS.ctx, "user_"+user_name, "password", pass1 )
	err = CS.redisClient.HMSet(CS.ctx, "user_"+user_name, "avatar", user_avatar).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库修改失败"
		return c.JSON(http.StatusOK, verify)
	}
	verify.Url = "accounts.html"
	return c.JSON(http.StatusOK, verify)
}
func accounts_start(c echo.Context) error {
	isi, ss := index_account(c, "accounts_start")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	user_name := c.Param("user")
	fmt.Println(user_name)
	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "激活用户成功" + user_name

	if ss.UserName != "root" {
		verify.Code = "00001"
		verify.Msg = "不是root用户" + ss.UserName
		return c.JSON(http.StatusOK, verify)
	}
	data, err := CS.redisClient.HGet(CS.ctx, "user_"+user_name, "active").Result()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库取用户状态失败" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	if data == "true" {
		verify.Code = "00002"
		verify.Msg = "用户已经是激活状态" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	err = CS.redisClient.HSet(CS.ctx, "user_"+user_name, "active", "true").Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库设置用户状态失败" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	verify.Url = "accounts.html"
	return c.JSON(http.StatusOK, verify)
}
func accounts_stop(c echo.Context) error {
	isi, ss := index_account(c, "accounts_stop")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	user_name := c.Param("user")
	fmt.Println(user_name)
	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "用户停用成功" + user_name

	if ss.UserName != "root" {
		verify.Code = "00001"
		verify.Msg = "不是root用户" + ss.UserName
		return c.JSON(http.StatusOK, verify)
	}
	if user_name == "root" {
		verify.Code = "00001"
		verify.Msg = "root用户不能停用" + ss.UserName
		return c.JSON(http.StatusOK, verify)
	}
	data, err := CS.redisClient.HGet(CS.ctx, "user_"+user_name, "active").Result()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库取用户状态失败" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	if data != "true" {
		verify.Code = "00002"
		verify.Msg = "用户不是激活状态" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	err = CS.redisClient.HSet(CS.ctx, "user_"+user_name, "active", "false").Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库设置用户状态失败" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	verify.Url = "accounts.html"
	return c.JSON(http.StatusOK, verify)
}

type User struct {
	UserId     int64  `json:"user_id"`
	UserName   string `json:"user_name"`
	UserNice   string `json:"user_nice"`
	UserPhone  string `json:"user_phone"`
	UserEmail  string `json:"user_email"`
	UserAvatar string `json:"user_avatar"`
	UserActive string `json:"user_active"`
}
type AccountsResponse struct {
	Code   string `json:"code"`    //00000
	Msg    string `json:"msg"`     //成功
	RecNum int64  `json:"rec_num"` //用户数
	ByUser string `json:"by_user"` //查询的用户
	User   []User `json:"user"`    //用户信息
}

func accounts_getall(c echo.Context) error {
	isi, ss := index_account(c, "accounts_getall")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	var ar AccountsResponse
	data, err := CS.redisClient.ZCard(CS.ctx, "userid").Result()
	if err != nil || data == 0 {
		ar.Code = "00404"
		ar.Msg = "取记录数失败"
		ar.RecNum = 0
		return c.JSON(http.StatusOK, ar)
	}
	ar.RecNum = data

	op := &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf", //Offset,Count
	}
	ret, err := CS.redisClient.ZRangeByScoreWithScores(CS.ctx, "userid", op).Result()
	if err != nil {
		ar.Code = "00500"
		ar.Msg = "取记录集合失败"
		ar.RecNum = 0
		return c.JSON(http.StatusOK, ar)
	}
	for _, z := range ret {
		fmt.Println(z.Member, z.Score)
		data, err := CS.redisClient.HGetAll(CS.ctx, "user_"+z.Member.(string)).Result()
		if err != nil || len(data) == 0 {
			continue
		}
		ar.User = append(ar.User, User{UserId: int64(z.Score), UserName: data["name"], UserNice: data["nice"], UserPhone: data["phone"], UserEmail: data["email"], UserAvatar: data["avatar"], UserActive: data["active"]})
	}
	ar.Code = "00000"
	ar.Msg = "取用户记录成功"
	ar.ByUser = ss.UserName
	return c.JSON(http.StatusOK, ar)
}

func accounts_reset(c echo.Context) error {
	isi, ss := index_account(c, "accounts_reset")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	//获取登录请求参数
	user_name := c.FormValue("user")
	pass0 := c.FormValue("pass0")
	pass1 := c.FormValue("pass1")
	pass2 := c.FormValue("pass2")
	fmt.Println(user_name)
	//fmt.Println(pass0)
	//fmt.Println(pass1)
	//fmt.Println(pass2)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "重置密码成功"
	isl, _ := login_user_check(user_name, pass0)
	if !isl {
		verify.Code = "00001"
		verify.Msg = "密码验证失败或未激活：" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	if pass1 != pass2 {
		verify.Code = "00002"
		verify.Msg = "两次输入的新密码不同"
		return c.JSON(http.StatusOK, verify)
	}
	if ss.UserName != user_name && ss.UserName != "root" {
		verify.Code = "00002"
		verify.Msg = "需要使用root用户重置其他用户密码"
		return c.JSON(http.StatusOK, verify)
	}
	err := CS.redisClient.HSet(CS.ctx, "user_"+user_name, "password", pass1).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库修改失败"
		return c.JSON(http.StatusOK, verify)
	}
	CS.redisClient.Save(CS.ctx)

	//passwd命令， 更新/etc/passwd和/etc/shadow
	cmd := exec.Command("sh", "-c", "echo -e '"+pass1+"\n"+pass2+"'|nsenter --mount=/hostip/1/ns/mnt --net=/hostip/1/ns/net -- passwd "+user_name)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	//eLogger.Info(cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	eLogger.Infof(out.String())

	//htpasswd命令 -b -D/ -b， 更新/opt/nginx/conf/htpasswd
	cmd = exec.Command("htpasswd", "-D", "/root/scas/conf/htpasswd.dat", user_name, pass0)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
	cmd = exec.Command("htpasswd", "-b", "/root/scas/conf/htpasswd.dat", user_name, pass1)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
	return c.JSON(http.StatusOK, verify)
}

func accounts_update(c echo.Context) error {
	isi, _ := index_account(c, "accounts_update")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	//获取登录请求参数
	user_name := c.FormValue("user")
	pass0 := c.FormValue("pass0")
	user_nice := c.FormValue("nice")
	email := c.FormValue("email")
	phone := c.FormValue("phone")
	fmt.Println(user_name)
	//fmt.Println(pass0)
	fmt.Println(user_nice)
	fmt.Println(email)
	fmt.Println(phone)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "更新用户" + user_name + "信息成功"
	isl, _ := login_user_check(user_name, pass0)
	if !isl {
		verify.Code = "00001"
		verify.Msg = "密码验证失败：" + user_name
		return c.JSON(http.StatusOK, verify)
	}

	err := CS.redisClient.HMSet(CS.ctx, "user_"+user_name, map[string]string{"nice": user_nice, "email": email, "phone": phone}).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库修改失败"
		return c.JSON(http.StatusOK, verify)
	}
	return c.JSON(http.StatusOK, verify)
}

func accounts_delete(c echo.Context) error {
	isi, ss := index_account(c, "accounts_delete")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}

	//获取登录请求参数
	user_name := c.FormValue("user")
	pass0 := c.FormValue("pass0")
	fmt.Println(user_name)
	//fmt.Println(pass0)

	var verify VerifyResponse
	verify.Code = "00000"
	verify.Msg = "删除用户" + user_name + "成功"
	if !passwd_check(user_name, pass0) {
		verify.Code = "00001"
		verify.Msg = "密码验证失败：" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	if user_name == ss.UserName || user_name == "root" {
		verify.Code = "00002"
		verify.Msg = "不能删除登录用户：" + user_name
		return c.JSON(http.StatusOK, verify)
	}

	err := CS.redisClient.ZRem(CS.ctx, "userid", user_name).Err()
	if err != nil {
		verify.Code = "00004"
		verify.Msg = "数据库删除ZREM失败" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	err = CS.redisClient.Del(CS.ctx, "user_"+user_name).Err()
	if err != nil {
		verify.Code = "00003"
		verify.Msg = "数据库删除KEY失败" + user_name
		return c.JSON(http.StatusOK, verify)
	}
	CS.redisClient.Save(CS.ctx)

	//deluser命令， 更新/etc/passwd和/etc/shadow
	cmd := exec.Command("sh", "-c", "nsenter --mount=/hostip/1/ns/mnt --net=/hostip/1/ns/net -- userdel "+user_name)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	//eLogger.Info(cmd.Args)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	eLogger.Infof(out.String())

	//htpasswd命令 -b -D， 更新/opt/nginx/conf/htpasswd
	cmd = exec.Command("htpasswd", "-D", "/root/scas/conf/htpasswd.dat", user_name, pass0)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()

	return c.JSON(http.StatusOK, verify)
}

//用户的防御策略
func accounts_defense(c echo.Context) error {
	isi, ss := index_account(c, "accounts_defense")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	return c.Render(http.StatusOK, "defense.html", ss)
}

//系统管理
func system_netdata(c echo.Context) error {
	isi, ss := index_account(c, "system_netdata")
	if !isi {
		return c.Redirect(http.StatusFound, "login.html")
	}
	return c.Render(http.StatusOK, "system_netdata.html", ss)
}

// 路径应该改为/etc/config/zero/authtoken.secret
func getAuthToken(c echo.Context) error {
	data, err := ioutil.ReadFile("./conf/authtoken.secret")
	if err != nil {
		log.Error("autoToken src read error")
		return c.JSON(http.StatusOK, "autoToken src read error")
	}
	var str AuthToken
	str.AuthToken = string(data)
	fmt.Println(str)
	str.AuthToken = strings.Trim(str.AuthToken, "\n\r")
	return c.JSON(http.StatusOK, str)
}
func getStatus(c echo.Context) error {
	var vpnstatus StatusRespone
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://172.17.0.1:9993/status", nil)
	if err != nil {
		log.Error("get error", err)
		return c.JSON(http.StatusOK, "get error")
	}
	req.Header.Add("X-ZT1-Auth", "o7pd9m6sltw3v552jils2zd0")
	resp, err := client.Do(req)
	result, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(result))
	err = json.Unmarshal(result, &vpnstatus)
	if err != nil {
		//fmt.Print("Unmarshalerr=", err)
		log.Error("Unmarshalerr error")
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}

	return c.JSON(http.StatusOK, vpnstatus)
}

type QueryBlocks struct {
	StartTime string `json:"StartTime" xml:"StartTime" form:"StartTime" query:"StartTime"`
	EndTime   string `json:"EndTime" xml:"EndTime" form:"EndTime" query:"EndTime"`
}

func queryTimeTransaction(c echo.Context) error {
	currentTime := time.Now()
	m, _ := time.ParseDuration("-5m")
	result := currentTime.Add(m)
	start := currentTime.Format("2006-01-02 15:04:05")
	end := result.Format("2006-01-02 15:04:05")
	q := &QueryBlocks{
		StartTime: start,
		EndTime:   end,
	}
	fmt.Println(q.StartTime)
	fmt.Println(q.EndTime)
	resp, err := http.PostForm("http://101.43.155.36:9000/queryTimeTranscation", url.Values{"StartTime": {q.StartTime}, "EndTime": {q.EndTime}})

	if err != nil {
		fmt.Printf("Error on request: %v\n", err)
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}

	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	fmt.Println(config)
	fmt.Println(len(config))

	return c.JSONBlob(http.StatusOK, body)
}

func queryTimeReceipts(c echo.Context) error {
	currentTime := time.Now()
	m, _ := time.ParseDuration("-5m")
	result := currentTime.Add(m)
	start := currentTime.Format("2006-01-02 15:04:05")
	end := result.Format("2006-01-02 15:04:05")
	q := &QueryBlocks{
		StartTime: start,
		EndTime:   end,
	}
	fmt.Println(q.StartTime)
	fmt.Println(q.EndTime)
	resp, err := http.PostForm("http://101.43.155.36:9000/queryTimeReceipt", url.Values{"StartTime": {q.StartTime}, "EndTime": {q.EndTime}})

	if err != nil {
		fmt.Printf("Error on request: %v\n", err)
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}

	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	fmt.Println(config)
	fmt.Println(len(config))

	return c.JSONBlob(http.StatusOK, body)
}

func queryBlockInfos(c echo.Context) error {
	blockType := c.Param("blockType")
	currentTime := time.Now()
	start := currentTime.Format("2006-01-02 15:04:05")
	resp, err := http.PostForm("http://101.43.155.36:9000/queryBlockInfos", url.Values{"blockType": {blockType}, "StartTime": {start}})

	if err != nil {
		fmt.Printf("Error on request: %v\n", err)
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return c.JSON(http.StatusOK, "Unmarshalerr error")
	}

	var config []map[string]interface{}

	err = json.Unmarshal(body, &config)
	fmt.Println(config)
	fmt.Println(len(config))
	marshal, err := json.Marshal(config)
	fmt.Println(string(marshal))
	return c.JSONBlob(http.StatusOK, marshal)
}
