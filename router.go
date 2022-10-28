package main

//func router(e *echo.Echo) {
//
//	//登录和介绍页面
//	e.GET("/", index)
//	e.GET("/introduct.html", introduct)
//	e.GET("/login.html", login)
//	e.GET("/forgot-password.html", forgot)
//	e.GET("/register.html", register)
//	e.GET("/logout.html", logout)
//	e.GET("/first.html", userInit)
//	//登录系列login
//	e.POST("/login", login_user)
//	e.POST("/hash", login_hash)
//
//	e.POST("/messages", login_messages)
//	//1.获取验证码信息，刷新页面时提交
//	//http://localhost:8000/captcha
//	e.GET("/captcha", captcha_id)
//	//2.获取验证码图片，支持点击后Form增加reload=true
//	//http://localhost:8000/captcha/gHEIwh7nWreTFb53MkVk.png
//	//http://localhost:8000/captcha/gHEIwh7nWreTFb53MkVk.png?reload=true
//	//var xhr = new XMLHttpRequest();
//	// 向logout接口发起post请求，是异步的
//	//xhr.open("POST", "/logout", true);
//	//xhr.send();
//	e.GET("/captcha/:captchaId", captcha_png)
//	//3.验证
//	//http://localhost:8000/verify/gHEIwh7nWreTFb53MkVk/647489/13910327212
//	e.GET("/verify/:captchaId/:value/:mobile", captcha_verify)
//
//	e.POST("/recover-password", login_recover)
//	e.POST("/register", login_register)
//
//	//管理功能
//	e.GET("/index.html", index)
//	e.GET("/products.html", products)
//	e.GET("/status.html", status)
//	e.GET("/blockstatus.html", bockStatus)
//	e.GET("/accounts.html", accounts)
//	e.GET("/add-user.html", accounts_add_user)
//	e.POST("/accounts_new", accounts_new)
//	e.GET("/accounts_start/:user", accounts_start)
//	e.GET("/accounts_stop/:user", accounts_stop)
//	e.POST("/accounts_username_validate", accounts_username_validate)
//	e.GET("/accounts_getall", accounts_getall)
//	e.POST("/accounts_reset", accounts_reset)
//	e.POST("/accounts_update", accounts_update)
//	e.POST("/accounts_delete", accounts_delete)
//	e.GET("/defense.html", accounts_defense)
//
//	e.GET("/system_netdata.html", system_netdata)
//	e.GET("/getVPNToken", getAuthToken)
//	e.POST("/first_new", first_new)
//	e.POST("/add_userimg", accounts_newimg)
//	e.POST("/update_userimg", accounts_updateimg)
//	e.GET("/getstatus", getStatus)
//	//查询区块链存证交易信息接口
//	e.GET("/queryTimeTransaction",queryTimeTransaction)
//	e.GET("/queryTimeReceipts",queryTimeReceipts)
//	// 查询区块链区块信息
//	e.GET("/queryBlockInfos/:blockType",queryBlockInfos)
//
//	// 404 等跳转xxx.html处理
//	e.HTTPErrorHandler = customHTTPErrorHandler
//}
//
