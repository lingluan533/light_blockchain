package  main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	Application string `json:"application"`
}
type TMAP struct {
	Boxn string `json:"boxn"`
	Code string `json:"code"`
}
type SmsRequest struct {
	Mobiles []string `json:"mobiles"`
	Tid string `json:"tid"`
	Tmap TMAP `json:"tmap"`
}

func GetToken( ) {
	client := &http.Client{}

	turl := CS.config.Message.TokenUrl
	data := url.Values{}
	data.Set("client_id", CS.config.Message.ClientId)
	data.Add("client_secret", CS.config.Message.ClientSecret)
	data.Add("grant_type", "client_credentials")
	req, _ := http.NewRequest("POST", turl, bytes.NewBufferString(data.Encode()))

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		eLogger.Error(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Status:", resp.Status)  //200
	fmt.Println("response Headers:", resp.Header)
	if err != nil {
		eLogger.Error(err)
		return
	}
	fmt.Println("response Body:", string(body))
	if resp.Status != "200" {
		eLogger.Error(resp)
		return
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		eLogger.Error(err)
		return
	}
	CS.config.Message.Token = tr.AccessToken
	if tr.ExpiresIn < CS.config.Message.Retry {
		CS.config.Message.Retry = tr.ExpiresIn
	}
}
func SendSms( mobile string, hostname string ) string {
	var letters = []rune("0123456789")
	newmessage := make([]rune, 6)
	r:=rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range newmessage {
		newmessage[i] = letters[r.Intn(10)]
	}

	client := &http.Client{}

	surl := CS.config.Message.SendUrl
	var sms SmsRequest
	sms.Mobiles = []string{mobile}
	sms.Tid = CS.config.Message.Tid
	sms.Tmap.Boxn = hostname
	sms.Tmap.Code = string(newmessage)
	bytesData, err := json.Marshal(sms)
	reader := bytes.NewReader(bytesData)
	req, _ :=  http.NewRequest("POST", surl, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+CS.config.Message.Token)
	resp, err := client.Do(req)
	if err != nil {
		eLogger.Error(err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	if err != nil {
		eLogger.Error(err)
		return ""
	}
	fmt.Println("response Body:", string(body))

	return string(newmessage)
}

func VerifySms( mobile string, msgid string ) bool {

	return true
}