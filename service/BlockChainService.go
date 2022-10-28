package service

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sca_server/container"
	"time"
)

type BlockChainService interface {
	QueryTimeReceiptsMethod() ([]byte, error)
	QueryTimeTransactionMethod() ([]byte, error)
}

type blockChainService struct {
	container container.Container
}

// 查询起止时间 结构体
type QueryBlocks struct {
	StartTime string `json:"StartTime" xml:"StartTime" form:"StartTime" query:"StartTime"`
	EndTime   string `json:"EndTime" xml:"EndTime" form:"EndTime" query:"EndTime"`
}

func (u *blockChainService) QueryTimeReceiptsMethod() ([]byte, error) {
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
		return nil, errors.Errorf("Unmarshalerr error")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, errors.Errorf("Unmarshalerr error")
	}

	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	fmt.Println(config)
	fmt.Println(len(config))

	return body, nil

}

func (u *blockChainService) QueryTimeTransactionMethod() ([]byte, error) {
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
		return nil, errors.New(" QueryTimeTransactionMethod Unmarshalerr error")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, errors.New(" QueryTimeTransactionMethod Unmarshalerr error")
	}
	var config []map[string]interface{}

	err = json.Unmarshal([]byte(body), &config)
	fmt.Println(config)
	fmt.Println(len(config))

	return body, nil
}

// NewBlockChainService is constructor.
func NewBlockChainService(container container.Container) BlockChainService {
	return &blockChainService{container: container}
}
