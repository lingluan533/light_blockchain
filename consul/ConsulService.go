package consul

import (
	"encoding/json"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"sca_server/config"
	"sca_server/model"
	"strconv"
)

// 获取节点的运行状态
func GetServiceSystemInfo(service *consulapi.CatalogService) model.LSysInfo {
	var res model.LSysInfo
	res.IpAddress = service.ServiceAddress
	res.NodeName = service.ServiceID
	resp, err := http.Get("http://" + service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort) + "/querySystemInfo")
	if err != nil {
		log.Infof("GetServiceSystemInfo Error on request: %v\n", err)
		res.OnlineStatus = "离线"
		return res
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(" GetServiceSystemInfo Error reading response body: %v\n", err)
		res.OnlineStatus = "离线"
		return res
	}
	//fmt.Printf(string(body))
	err = json.Unmarshal(body, &res)
	//log.Printf("%s", res)
	res.OnlineStatus = "在线"
	return res

}

//获取服务下的所有节点的信息
func GetAllServiceByName(config *config.Config) ([]*consulapi.CatalogService, error) {
	// 创建连接consul服务配置
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = config.Consul.ConsulAddress + ":" + config.Consul.ConsulPort
	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		fmt.Println("consul client error : ", err)
		return nil, err
	}

	// 获取指定service
	res, _, _ := client.Catalog().Service(config.Consul.Name, "", nil)
	fmt.Println("一共找到服务个数：", len(res))
	return res, nil

}

//获取服务的数量 在线，离线
func GetNumberOfServices(config *config.Config) (int, int) {
	var onlineNumer int
	var offlineNumber int
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = config.Consul.ConsulAddress + ":" + config.Consul.ConsulPort
	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		fmt.Println("consul client error : ", err)
		return onlineNumer, offlineNumber
	}

	// 获取指定service
	serviceHealthy, _, err := client.Health().State("any", nil)
	if err != nil {
		fmt.Println("consul client error : ", err)
		return onlineNumer, offlineNumber
	}
	for i := 0; i < len(serviceHealthy); i++ {
		service := serviceHealthy[i]
		if service.Status == "critical" {
			offlineNumber++
		} else if service.Status == "passing" && service.Name != "Serf Health Status" {
			onlineNumer++
		}
	}
	return onlineNumer, offlineNumber
}

// 随机获取一个节点
func GetOneOnlineAddress(config *config.Config) (*consulapi.AgentService, error) {
	// 创建连接consul服务配置
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = config.Consul.ConsulAddress + ":" + config.Consul.ConsulPort
	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		fmt.Println("consul client error : ", err)
		return nil, err
	}

	// 获取指定service
	serviceHealthy, _, err := client.Health().Service(config.Consul.Name, "", true, nil)

	if err != nil {
		fmt.Println("consul client error : ", err)
		return nil, err
	}
	if len(serviceHealthy) == 0 {
		return nil, nil
	}
	return serviceHealthy[0].Service, nil

}
