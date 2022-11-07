package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"time"

	"log"
	"net/http"

	"os"
	"path/filepath"
	"sca_server/consul"
	"sca_server/container"

	"sca_server/model"
	"strconv"
	"strings"
)

type DataFileService interface {
	GetAllFiles() []model.DataFile
	SaveOnChainOfDownloadRecord(filePath, userName string) (bool, error)
}
type dataFileService struct {
	container container.Container
}

func (d dataFileService) SaveOnChainOfDownloadRecord(filePath, userName string) (bool, error) {
	var userBehaviour = new(model.DataReceipts)
	var receipt = new(model.Receipt)
	receipt.FileName = filePath
	receipt.FileSize = 104.00
	receipt.ReceiptValue = 1000
	receipt.Version = "v1.0"
	receipt.OperationType = "DOWNLOAD"
	receipt.KeyId = d.container.GetConfig().EtcdPrefixConfig.UserOperation + "_" + userName + "_" + filePath
	userBehaviour.EntityId = d.container.GetConfig().EtcdPrefixConfig.UserOperation + "_" + userName + "_" + filePath
	userBehaviour.CreateTimestamp = time.Now().String()
	userBehaviour.DataRecNum = 1
	userBehaviour.DataValue = 1
	receipt.UserName = userName
	receipt.AttachmentTotalHash = "hash"
	receipt.AttachmentFileUris = nil
	receipt.ParentKeyId = "hash"
	receipt.Uri = filePath
	receipt.FileHash = "hash"
	receipt.ServiceType = "user_behaviour"
	receipt.DataType = "user_behaviour"
	userBehaviour.Receipts = append(userBehaviour.Receipts, *receipt)
	service, err := consul.GetOneOnlineAddress(d.container.GetConfig())
	logger := d.container.GetLogger()
	logger.GetZapLogger().Info("Finding EdgeNode Server....")
	if service == nil {
		logger.GetZapLogger().Errorf("No Avaliable EdgeNode!")
		return false, errors.New("No Avaliable EdgeNode!")
	}
	if err != nil {
		logger.GetZapLogger().Errorf("Error on request Avaliable EdgeNode: %v\n", err)
		return false, errors.Errorf("Error on request Avaliable EdgeNode")
	}
	logger.GetZapLogger().Info("Got  EdgeNode Server " + service.Address)
	data, err := json.Marshal(userBehaviour)
	resp, err := http.Post("http://"+service.Address+":"+strconv.Itoa(service.Port)+"/storeOperationRecord", "application/json", bytes.NewBuffer(data))
	logger.GetZapLogger().Info("http.Post finished ")
	if err != nil {
		fmt.Printf("Error on request: %v\n", err)
		return false, errors.New(" storeOperationRecord Error on request:" + err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, nil
	}

}

var filePaths []string
var fileInfos []model.DataFile

func visit(files *[]string, fileInfos *[]model.DataFile) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		*files = append(*files, path)
		var fileInfo = new(model.DataFile)
		fileInfo.FileSize = strconv.FormatInt(info.Size(), 10) + "B"
		fileInfo.FilePath = strings.Replace(path, "\\", "\\\\", -1)
		fileInfo.UpdateTime = info.ModTime().String()
		fileInfo.FileName = info.Name()
		if info.IsDir() {
			fileInfo.FileType = "文件夹"
		} else if strings.Contains(path, "MINUTE") {
			fileInfo.FileType = "分钟块"
		} else if strings.Contains(path, "TENMINUTE") {
			fileInfo.FileType = "增强块"
		} else if strings.Contains(path, "DAY") {
			fileInfo.FileType = "天块文件"
		}

		if strings.Contains(path, "user_behaviour") || strings.Contains(path, "node_credible") || strings.Contains(path, "service_access") {
			fileInfo.ChainOfFile = "存证文件"
		} else {
			fileInfo.ChainOfFile = "交易文件"
		}
		//E:\Go_WorkSpace\hraft1102\scope\2022-10-29\user_behaviour\MINUTE\593 593 41824 2022-10-29 09:54:30.7801984 +0800 CST -rw-rw-rw-
		*fileInfos = append(*fileInfos, *fileInfo)
		//fmt.Println(fileInfo)
		//fmt.Println(path + " " + info.Name() + " " + strconv.Itoa(int(info.Size())) + " " + info.ModTime().String() + " " + info.Mode().String())
		return nil
	}
}
func (d dataFileService) GetAllFiles() []model.DataFile {
	root := "E:\\Go_WorkSpace\\hraft1102\\scope"
	err := filepath.Walk(root, visit(&filePaths, &fileInfos))
	if err != nil {
		fmt.Println("filepath.Walk err ", err)
	}
	return fileInfos
}

func NewDataFileService(container container.Container) DataFileService {
	return &dataFileService{container: container}
}
