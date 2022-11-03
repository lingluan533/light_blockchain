package service

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sca_server/consul"
	"sca_server/container"
	"strconv"
	"time"
)

type BlockChainService interface {
	QueryTimeReceiptsMethod() ([]byte, error)
	QueryTimeTransactionMethod() ([]byte, error)
	QueryBlockInfosMethod(blockType string) ([]byte, error)
}

type blockChainService struct {
	container container.Container
}

// 查询起止时间 结构体
type QueryBlocks struct {
	StartTime int64 `json:"StartTime" xml:"StartTime" form:"StartTime" query:"StartTime"`
	EndTime   int64 `json:"EndTime" xml:"EndTime" form:"EndTime" query:"EndTime"`
}

func (u *blockChainService) QueryTimeReceiptsMethod() ([]byte, error) {
	currentTime := time.Now()
	m, _ := time.ParseDuration("-5m")

	logger := u.container.GetLogger()
	result := currentTime.Add(m)

	end := currentTime.Unix()
	start := result.Unix()
	q := &QueryBlocks{
		StartTime: start,
		EndTime:   end,
	}
	fmt.Println(q.StartTime)
	fmt.Println(q.EndTime)
	// get a avaliable server
	service, err := consul.GetOneOnlineAddress(u.container.GetConfig())
	//logger.GetZapLogger().Errorf(" QueryTimeReceiptsMethod No Avaliable EdgeNode! %v", u.container.GetConfig().Consul)
	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return nil, errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return nil, errors.Errorf("Error on request Avaliable EdgeNode")
	}
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/queryTimeReceipt", url.Values{"StartTime": {strconv.Itoa(int(q.StartTime))}, "EndTime": {strconv.Itoa(int(q.EndTime))}})

	if err != nil {
		logger.GetZapLogger().Errorf("Error on request: %v\n", err)
		return nil, errors.Errorf("Unmarshalerr error")
	}
	defer resp.Body.Close()
	u.container.GetLogger().GetZapLogger().Info(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, errors.Errorf("Unmarshalerr error")
	}

	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	//fmt.Println(config)
	fmt.Println(len(config))
	return body, nil

}

func (u *blockChainService) QueryTimeTransactionMethod() ([]byte, error) {
	currentTime := time.Now()
	m, _ := time.ParseDuration("-5m")
	result := currentTime.Add(m)
	start := result.Unix()
	q := &QueryBlocks{
		StartTime: start,
		EndTime:   currentTime.Unix(),
	}
	fmt.Println(q.StartTime)
	fmt.Println(q.EndTime)
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
		return nil, errors.Errorf("Error on request Avaliable EdgeNode")
	}
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/queryTimeTransaction", url.Values{"StartTime": {strconv.Itoa(int(q.StartTime))}, "EndTime": {strconv.Itoa(int(q.EndTime))}})

	if err != nil {
		fmt.Printf("Error on request: %v\n", err)
		return nil, errors.New(" QueryTimeTransactionMethod Unmarshalerr error")
	}
	defer resp.Body.Close()
	u.container.GetLogger().GetZapLogger().Info(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, errors.New(" QueryTimeTransactionMethod Unmarshalerr error")
	}
	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	//fmt.Println(config)
	fmt.Println(len(config))

	return body, nil
}
func (u *blockChainService) QueryBlockInfosMethod(blockType string) ([]byte, error) {
	currentTime := time.Now()
	logger := u.container.GetLogger()
	start := currentTime.Format("2006-01-02 15:04:05")
	fmt.Println(start)
	service, err := consul.GetOneOnlineAddress(u.container.GetConfig())
	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return nil, errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return nil, errors.Errorf("Error on request Avaliable EdgeNode")
	}
	resp, err := http.PostForm("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/queryBlockInfos", url.Values{"blockType": {blockType}, "StartTime": {start}})
	//fmt.Println(resp)
	if err != nil {
		fmt.Printf("Error on request: %v\n", err)
		return nil, errors.New(" QueryTimeTransactionMethod Error on request:" + err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, errors.New(" QueryTimeTransactionMethod Error on ioutil.ReadAll:" + err.Error())
	}
	//	fmt.Println(string(body))

	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	//fmt.Println(config)
	fmt.Println(len(config))
	//marshal, err := json.Marshal(config)
	//fmt.Println(string(marshal))

	return body, nil
}

// NewBlockChainService is constructor.
func NewBlockChainService(container container.Container) BlockChainService {
	return &blockChainService{container: container}
}
