package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sca_server/consul"
	"sca_server/container"
	"sca_server/model"

	"strconv"
)

type UserService interface {
	LoginMethod(user string, password string) (bool, error)
	GetAllOperationRecordsByUserName(user string) error
}

type userService struct {
	container container.Container
}

func (u userService) GetAllOperationRecordsByUserName(user string) error {
	var data model.DataReceipts
	logger := u.container.GetLogger()
	// get a avaliable server
	service, err := consul.GetOneOnlineAddress(u.container.GetConfig())
	//logger.GetZapLogger().Errorf(" QueryTimeReceiptsMethod No Avaliable EdgeNode! %v", u.container.GetConfig().Consul)
	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return errors.New("Error on request Avaliable EdgeNode")
	}
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/queryOperationRecordsByUserName", url.Values{"user": {user}})

}

func (u userService) LoginMethod(user string, password string) (bool, error) {

	logger := u.container.GetLogger()
	// get a avaliable server
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
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/login", url.Values{"user": {user}, "password": {password}})

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

	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	//fmt.Println(config)
	//fmt.Println(len(config))
	return true, nil
}

// NewUserService is constructor.
func NewUserService(container container.Container) UserService {
	return &userService{container: container}
}
