package model

type DataFile struct {
	FilePath    string `json:filePath`
	FileName    string `json:fileName`
	FileType    string `json:fileType`
	ChainOfFile string `json:chainOfFile`
	FileSize    string `json:fileSize`
	UpdateTime  string `json:updateTime`
}
