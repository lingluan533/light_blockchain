package model

type Receipt struct {
	KeyId               string   `json:"keyId" validate:"required"`
	ReceiptValue        float64  `json:"receiptValue"`
	Version             string   `json:"version"`
	UserName            string   `json:"userName"`
	OperationType       string   `json:"operationType"`
	DataType            string   `json:"dataType" validate:"required"`
	ServiceType         string   `json:"serviceType"`
	FileName            string   `json:"fileName"`
	FileSize            float64  `json:"fileSize"`
	FileHash            string   `json:"fileHash"`
	Uri                 string   `json:"uri"`
	ParentKeyId         string   `json:"parentKeyId"`
	AttachmentFileUris  []string `json:"attachmentFileUris"`
	AttachmentTotalHash string   `json:"attachmentTotalHash"`
}

type DataReceipts struct {
	CreateTimestamp string    `json:"createTimestamp" validate:"required"`
	EntityId        string    `json:"entityId"`
	DataValue       float64   `json:"dataValue"`
	DataRecNum      int64     `json:"dataRecNum"`
	Receipts        []Receipt `json:receipts`
}
