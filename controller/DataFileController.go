package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"sca_server/container"
	"sca_server/service"
	"sort"
)

type DataFileController interface {
	ShowDataFiles(c echo.Context) error
	DownLoadFile(c echo.Context) error
	DownLoad(c echo.Context) error
	GetTotalSizeOfDataFiles(c echo.Context) error
}

type dataFileController struct {
	container container.Container
	service   service.DataFileService
}

func (d dataFileController) GetTotalSizeOfDataFiles(c echo.Context) error {
	sizeMap := d.service.GetTotalSizeOfDataFiles()
	return c.JSON(http.StatusOK, sizeMap)
}

func (d dataFileController) DownLoad(c echo.Context) error {
	filePath := c.QueryParam("filePath")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("read file err,", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}
	return c.Blob(http.StatusOK, "text/csv", data)
}

func (d dataFileController) DownLoadFile(c echo.Context) error {
	//1.检查权限
	//2.记录上链
	filePath := c.FormValue("filePath")
	fmt.Println(filePath)
	res, err := d.service.SaveOnChainOfDownloadRecord(filePath, "zms")
	if err != nil {
		fmt.Println("DownLoadFile err=", err)
		return c.JSON(http.StatusInternalServerError, nil)
	}
	if res {
		return c.JSON(http.StatusOK, nil)
	} else {
		return c.JSON(http.StatusInternalServerError, nil)
	}

}

func (d dataFileController) ShowDataFiles(c echo.Context) error {
	fileInfos := d.service.GetAllFiles()
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].UpdateTime > fileInfos[j].UpdateTime
	})
	if len(fileInfos) < 25 {
		return c.JSON(http.StatusOK, fileInfos)
	}
	return c.JSON(http.StatusOK, fileInfos)
}
func NewDataFileController(container container.Container) DataFileController {
	return &dataFileController{
		container: container,
		service:   service.NewDataFileService(container),
	}
}
