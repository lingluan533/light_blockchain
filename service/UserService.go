package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"sca_server/config"
	"sca_server/consul"
	"sca_server/container"
	"sca_server/model"
	"strconv"
	"strings"
)

type UserService interface {
	LoginMethod(user string, password string) (bool, error)
	GetAllOperationRecordsByUserName(user string) ([]byte, error)
	RegisterMethod(userInfo model.UserInfo) (bool, error)
}

type userService struct {
	container container.Container
}

func (u userService) RegisterMethod(userInfo model.UserInfo) (bool, error) {
	logger := u.container.GetLogger()
	// get a avaliable server
	log.Info(userInfo)
	service, err := consul.GetOneOnlineAddress(u.container.GetConfig())
	//logger.GetZapLogger().Errorf(" QueryTimeReceiptsMethod No Avaliable EdgeNode! %v", u.container.GetConfig().Consul)
	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return false, errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return false, errors.New("Error on request Avaliable EdgeNode")
	}
	marshal, err := json.Marshal(userInfo)
	resp, err := http.Post("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/register", "application/json", strings.NewReader(string(marshal)))
	//resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/register", url.Values{"UserName": {userInfo.UserName}, "Password": {userInfo.Password}, "Phone": {userInfo.Phone}, "Email": {userInfo.Email}})
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request: %v\n", err)
		return false, errors.New("Unmarshalerr error")
	}
	defer resp.Body.Close()
	u.container.GetLogger().GetZapLogger().Info(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return false, errors.New("Unmarshalerr error")
	}
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	m := res["err"]
	if m == config.UserRegisterSuccess {
		return true, nil
	} else if m == config.UserRepeatRegister {
		return false, errors.New(m.(string))
	} else {
		return false, errors.New(config.ServerInternalError)
	}

}

func (u userService) GetAllOperationRecordsByUserName(user string) ([]byte, error) {
	logger := u.container.GetLogger()
	// get a avaliable server
	service, err := consul.GetOneOnlineAddress(u.container.GetConfig())
	//logger.GetZapLogger().Errorf(" QueryTimeReceiptsMethod No Avaliable EdgeNode! %v", u.container.GetConfig().Consul)
	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return nil, errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return nil, errors.New("Error on request Avaliable EdgeNode")
	}
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/queryOperationRecordsByUserName", url.Values{"user": {user}})
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request: %v\n", err)
		return nil, errors.New("Unmarshalerr error")
	}
	//defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, nil

}

func (u userService) LoginMethod(user string, password string) (bool, error) {

	logger := u.container.GetLogger()
	// get a avaliable server
	service, err := consul.GetOneOnlineAddress(u.container.GetConfig())

	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return false, errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return false, errors.New("Error on request Avaliable EdgeNode")
	}
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/login", url.Values{"UserName": {user}, "Password": {password}})

	if err != nil {
		logger.GetZapLogger().Errorf("Error on request: %v\n", err)
		return false, errors.New("Unmarshalerr error")
	}
	defer resp.Body.Close()
	u.container.GetLogger().GetZapLogger().Info(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return false, errors.New("Unmarshalerr error")
	}
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	m := res["err"]
	fmt.Println(m)
	if m == config.UserLoginSuccess {
		return true, nil
	} else {
		return false, errors.New(res["data"].(string))
	}
}

// NewUserService is constructor.
func NewUserService(container container.Container) UserService {
	return &userService{container: container}
}
