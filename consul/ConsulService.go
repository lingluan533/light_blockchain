package consul

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"sca_server/config"
)

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
