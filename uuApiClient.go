// Copyright 2013 www.xwi.me
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package UuApiClient

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	UU_VERSION = "1.1.0.1"
	macAddress = "00e021ac7d"
	LOGIN_URL  = "%s/Upload/Login.aspx?U=%s&P=%s&R=%s"
	UPLOAD_URL = "%s/Upload/Processing.aspx?R=%s"
	RESULT_URL = "%s/Upload/GetResult.aspx?KEY=%s&ID=%s&Random=%s"
	REPORT_URL = "%s/Upload/ReportError.aspx?KEY=%s&ID=%s&SID=%s&SKEY=%s&R=%s"
	ERROR_CODE = "-3"
)

var (
	chttp          *httpclient.HttpClient
	timeOut        = 60 //超时 单位秒
	uhash          string
	gkey           string
	userAgent      string
	userKey        string
	softContentKey string
)

type UuApiClient struct {
	*uuApi
}

type uuApi struct {
	softID   string
	softKey  string
	userName string
	passWord string
	uid      string
	serverUrl
}
type serverUrl struct {
	serviceHost string
	uploadHost  string
	codeHost    string
}

func newUuApi(softID, softKey, userName, passWord string) *uuApi {
	c := &uuApi{
		softID:   softID,
		softKey:  softKey,
		userName: userName,
		passWord: passWord,
		uid:      "100",
	}
	uhash = getMd5Str(softID)

	userAgent = getMd5Str(strings.ToUpper(softKey) + strings.ToUpper(c.userName) + macAddress)
	gkey = getMd5Str(strings.ToUpper(softKey+userName)) + macAddress
	return c
}
func getMd5Str(value string) string {
	h := md5.New()
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}
func New(softID, softKey, userName, passWord string) *UuApiClient {
	c := newUuApi(softID, softKey, userName, passWord)
	C := &UuApiClient{c}
	return C
}

func (u *uuApi) AfterPropertiesSet() {
	u.initHeaders()
	u.refreshServer()
}

/**
 * Http请求
 */
func (u *uuApi) uuGetUrl() string {
	res, _ := chttp.Get("http://common.taskok.com:9000/Service/ServerConfig.aspx", nil)
	bodyString, _ := res.ToString()
	return bodyString
}

/**
 * 获取服务器的地址
 */
func (u *uuApi) refreshServer() bool {

	bodyString := u.uuGetUrl()
	reg, err := regexp.Compile(",(.*?):101,(.*?):102,(.*?):103")
	if err != nil {
		fmt.Println("LEN OF LOGS = ", err)
		return false
	}

	arr := reg.FindStringSubmatch(bodyString)
	if len(arr) < 1 {
		fmt.Println("出错")
		//return "-1001"
		return false
	}
	u.serverUrl.uploadHost = "http://" + arr[2]
	u.serverUrl.codeHost = "http://" + arr[3]
	u.serverUrl.serviceHost = "http://" + arr[1]
	return true
}

/*
 * 初始化头
 */
func (u *uuApi) initHeaders() {
	chttp = httpclient.NewHttpClient(nil)
	chttp.WithOption(httpclient.OPT_USERAGENT, userAgent)
	chttp.WithHeader("Accept", "text/html, application/xhtml+xml, */*")
	chttp.WithHeader("Accept-Language", "zh-CN")
	chttp.WithHeader("Connection", "Keep-Alive")
	chttp.WithHeader("Cache-Control", "no-cache")
	chttp.WithHeader("SID", u.softID)
	chttp.WithHeader("HASH", uhash)
	chttp.WithHeader("UUVersion", UU_VERSION)
	chttp.WithHeader("UID", u.uid)
	chttp.WithHeader("KEY", gkey)
	chttp.WithOption(httpclient.OPT_PROXY, "127.0.0.1:8888")
}

/*
  用户登录,登录成功返回用户的ID
*/
func (u *uuApi) UserLogin() error {
	if u.softID == "" && u.softKey == "" && u.userName == "" && u.passWord == "" {
		return errors.New("-1")
	}
	url := fmt.Sprintf(LOGIN_URL, u.serverUrl.serviceHost, u.userName, getMd5Str(u.passWord), fmt.Sprintf("%d", currentTimeMillis()))
	res, _ := chttp.Get(url, nil)
	bodyString, _ := res.ToString()
	if strings.Contains(bodyString, "_") {
		array := strings.Split(bodyString, "_")
		u.uid = array[0]
		userKey = bodyString
		softContentKey = getMd5Str(strings.ToLower(userKey + u.softID + u.softKey))
		u.initHeaders()
		return nil
	}

	return errors.New(bodyString)
}

/*
  上传，返回服务上传码，根据返回码获取结果
*/
func (u *uuApi) Upload(imagePath, codeType string) string {
	url := fmt.Sprintf(UPLOAD_URL, u.serverUrl.uploadHost, fmt.Sprintf("%d", currentTimeMillis()))

	res, _ := chttp.PostMultipart(url, map[string]string{
		"@img":    imagePath,
		"key":     userKey,
		"sid":     u.softID,
		"skey":    softContentKey,
		"TimeOut": strconv.Itoa(timeOut),
		"Type":    codeType,
		"Version": "100",
	})
	bodyString, _ := res.ToString()
	return bodyString
}

/*
  传入图片地址自动获得结果
*/
func (u *uuApi) AutoRecognition(imagePath, codeType string) string {
	codeId := u.Upload(imagePath, codeType)
	var result = ERROR_CODE
	if !strings.Contains(codeId, "|") {
		var timer = 0
		for timer < timeOut && strings.EqualFold(ERROR_CODE, result) {
			result = u.GetResult(codeId)
			if !strings.EqualFold(ERROR_CODE, result) {
				return result
			}
			time.Sleep(time.Second)
			timer++
		}
	} else {
		result = codeId
	}
	if !strings.EqualFold(result, ERROR_CODE) {
		arrayResult := strings.Split(result, "|")
		if len(arrayResult) >= 2 {
			return arrayResult[1]
		}
	}
	return "-1002"
}

/*
  根据服务码获取结果
*/
func (u *uuApi) GetResult(codeId string) string {
	url := fmt.Sprintf(RESULT_URL, userKey, codeId, fmt.Sprintf("%d", currentTimeMillis()))
	res, _ := chttp.Get(url, nil)
	bodyString, _ := res.ToString()
	if bodyString == "-3" {
		return "-1002"
	}
	return bodyString
}
func (u *uuApi) ReportError(codeId string) string {

	url := fmt.Sprintf(REPORT_URL, userKey, codeId, u.softID, softContentKey, fmt.Sprintf("%d", currentTimeMillis()))
	res, _ := chttp.Get(url, nil)
	result, _ := res.ToString()
	if result == "OK" {
		return "OK"
	}
	return result
}
func currentTimeMillis() int64 {
	return time.Now().UnixNano() / 1000000
}
