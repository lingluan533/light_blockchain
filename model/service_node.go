package model

type LSysInfo struct {
	OnlineStatus   string  `json:"onlineStatus"`
	IpAddress      string  `json:"ipAddress"`
	NodeName       string  `json:"nodeName"`
	MemAll         uint64  `json:"memAll"`
	MemFree        uint64  `json:"memFree"`
	MemUsed        uint64  `json:"memUsed"`
	MemUsedPercent float64 `json:"memUsedPercent"`
	Days           int64   `json:"days"`
	Hours          int64   `json:"hours"`
	Minutes        int64   `json:"minutes"`
	Seconds        int64   `json:"seconds"`
	CpuUsedPercent float64 `json:"cpuUsedPercent"`
	OS             string  `json:"os"`
	Arch           string  `json:"arch"`
	CpuCores       int     `json:"cpuCores"`
}
