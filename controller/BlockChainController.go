package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"sca_server/container"
	"sca_server/service"
)

type BlockChainController interface {
	QueryBlockInfos(c echo.Context) error
	QueryTimeReceipts(c echo.Context) error
	QueryTimeTransaction(c echo.Context) error
}

type blockChainController struct {
	container container.Container
	service   service.BlockChainService
}

func (controller blockChainController) QueryBlockInfos(c echo.Context) error {
	blockType := c.Param("blockType")
	data, err := controller.service.QueryBlockInfosMethod(blockType)
	if err != nil {
		return c.JSONBlob(http.StatusInternalServerError, nil)
	}
	return c.JSONBlob(http.StatusOK, data)
}
func (controller blockChainController) QueryTimeTransaction(c echo.Context) error {
	data, err := controller.service.QueryTimeTransactionMethod()
	if err != nil {
		return c.JSONBlob(http.StatusInternalServerError, nil)
	}
	return c.JSONBlob(http.StatusOK, data)
}

func (controller blockChainController) QueryTimeReceipts(c echo.Context) error {
	data, err := controller.service.QueryTimeReceiptsMethod()
	if err != nil {
		return c.JSONBlob(http.StatusInternalServerError, nil)
	}
	return c.JSONBlob(http.StatusOK, data)
}

func NewBlockChainController(container container.Container) BlockChainController {
	return &blockChainController{
		container: container,
		service:   service.NewBlockChainService(container),
	}
}
