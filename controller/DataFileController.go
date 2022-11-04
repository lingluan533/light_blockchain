package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type DataFileController interface {
	ShowDataFiles(c echo.Context) error
	DownLoadFile(c echo.Context) error
}

type dataFileController struct {
	container container.Container
	service   service.DataFileService
}

func (d dataFileController) DownLoadFile(c echo.Context) error {
	//1.检查权限
	//2.记录上链
	filePath := c.QueryParam("filePath")
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

	return c.JSON(http.StatusOK, fileInfos[:20])
}
func NewDataFileController(container container.Container) DataFileController {
	return &dataFileController{
		container: container,
		service:   service.NewDataFileService(container),
	}
}
