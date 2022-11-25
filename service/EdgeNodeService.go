package service

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"sca_server/consul"
	"sca_server/container"
	"sca_server/model"
	"strconv"
)

type EdgeNodeService interface {
	GetServiceNumberOfOnlineAndOffline() (int, int)
	GetCLusterRunStatus() []model.LSysInfo
	SayHelloEdgeNodeMethod(ip string) error
	RebootEdgeNodeMethod(ip string) (bool, error)
}
type edgeNodeService struct {
	container container.Container
}

func (e edgeNodeService) RebootEdgeNodeMethod(ip string) (bool, error) {
	resp, err := http.Get("http://" + ip + ":" + strconv.Itoa(9000) + "/restart")
	logger := e.container.GetLogger()
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request  RebootEdgeNodeMethod =  %v\n", err)
		return false, err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response RebootEdgeNodeMethod body: %v\n", err)
		return false, err
	}
	return true, nil
}

func (e edgeNodeService) SayHelloEdgeNodeMethod(ip string) error {
	resp, err := http.Get("http://" + ip + ":" + strconv.Itoa(9000) + "/sayHello")
	logger := e.container.GetLogger()
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request  SayHelloEdgeNodeMethod =  %v\n", err)
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response SayHelloEdgeNodeMethod body: %v\n", err)
		return nil
	}
	return err
}

func (e edgeNodeService) GetCLusterRunStatus() []model.LSysInfo {
	services, err := consul.GetAllServiceByName(e.container.GetConfig())
	if err != nil {
		log.Infof("GetServiceRunStatus server get err!")
		return nil
	}
	var sysInfos = make([]model.LSysInfo, 0, 100)
	fmt.Println("个数：len(services)=", len(services))
	for i := 0; i < len(services); i++ {
		fmt.Printf("当前服务iP:%v,i=%v\n\t", services[i].ServiceAddress, i)
		temp := consul.GetServiceSystemInfo(services[i])
		sysInfos = append(sysInfos, temp)
		fmt.Println("????????", len(sysInfos))
	}
	fmt.Println("****************", len(sysInfos))
	fmt.Println(sysInfos)
	return sysInfos
}

func (e edgeNodeService) GetServiceNumberOfOnlineAndOffline() (int, int) {
	online, offline := consul.GetNumberOfServices(e.container.GetConfig())
	return online, offline
}

func NewEdgeNodeService(container container.Container) EdgeNodeService {
	return &edgeNodeService{container: container}
}
