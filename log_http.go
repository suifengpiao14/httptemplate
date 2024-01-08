package httptemplate

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/suifengpiao14/logchan/v2"
	"moul.io/http2curl"
)

type LogName string

func (l LogName) String() string {
	return string(l)
}

const (
	LOG_INFO_CURL_RAW LogName = "LogInfoCURLRaw"
)

type LogInfoCURLRaw struct {
	HttpRaw string `json:"httpRaw"`
	Out     string `json:"out"`
	Err     error  `json:"error"`
	Level   string `json:"level"`
	logchan.EmptyLogInfo
}

func (l *LogInfoCURLRaw) GetName() logchan.LogName {
	return LOG_INFO_CURL_RAW
}
func (l *LogInfoCURLRaw) Error() error {
	return l.Err
}
func (l *LogInfoCURLRaw) GetLevel() string {
	return l.Level
}

const (
	LogInfoNameHttp LogName = "HttpLogInfo"
)

//LogInfoHttp 发送日志，只需填写 Request(GetRequest),Response 和RequestBody,其余字段会在BeforSend自动填充
type LogInfoHttp struct {
	Name           string         `json:"name"`
	Request        *http.Request  `json:"-"`
	Response       *http.Response `json:"-"`
	Method         string         `json:"method"`
	Url            string         `json:"url"`
	RequestHeader  http.Header    `json:"requestHeader"`
	RequestBody    string         `json:"requestBody"`
	ResponseHeader http.Header    `json:"responseHeader"`
	ResponseBody   string         `json:"responseBody"`
	CurlCmd        string         `json:"curlCmd"`
	Err            error
	GetRequest     func() (request *http.Request) //go-resty/resty/v2 RawRequest 一开始为空，提供函数延迟实现
	logchan.EmptyLogInfo
}

func (h *LogInfoHttp) GetName() (logName logchan.LogName) {
	return LogInfoNameHttp
}

func (h *LogInfoHttp) Error() (err error) {
	return err
}

// 简化发送方赋值
func (h *LogInfoHttp) BeforeSend() {
	if h.GetRequest != nil {
		h.Request = h.GetRequest() // 优先使用延迟获取
	}
	req := h.Request
	resp := h.Response
	if req == nil && resp != nil && resp.Request != nil {
		h.Request = resp.Request
		req = resp.Request
	}
	if req != nil {
		bodyReader, err := req.GetBody()
		if err == nil {
			body, err := io.ReadAll(bodyReader)
			if err == nil {
				h.RequestBody = string(body)
				contentType := req.Header.Get("Content-type")              //"multipart/form-data; boundary=39c610d0ba0041b90dad8f1d477a6c3dfde830e124274c9972d2668c52db"
				if !strings.Contains(contentType, "multipart/form-data") { // 上传文件时屏蔽body，因为太大了
					req.Body = io.NopCloser(bytes.NewReader(body))
				}
			}

		}
		h.Method = req.Method
		h.Url = req.URL.String()
		h.RequestHeader = req.Header.Clone()
		curlCommand, err := http2curl.GetCurlCommand(h.Request)
		if err == nil {
			h.CurlCmd = curlCommand.String()
		}
	}

	if resp != nil {
		if resp.Body != nil {
			responseBody, err := io.ReadAll(resp.Body)
			if err == nil {
				h.ResponseBody = string(responseBody)
			}
		}
		h.ResponseHeader = resp.Header.Clone()
	}
}

//DefaultPrintHttpLogInfo 默认日志打印函数
func DefaultPrintHttpLogInfo(logInfo logchan.LogInforInterface, typeName logchan.LogName, err error) {
	if typeName != LogInfoNameHttp {
		return
	}
	httpLogInfo, ok := logInfo.(*LogInfoHttp)
	if !ok {
		return
	}
	if err != nil {
		_, err1 := fmt.Fprintf(logchan.LogWriter, "%s|loginInfo:%s|error:%s\n", logchan.DefaultPrintLog(httpLogInfo), httpLogInfo.GetName(), err.Error())
		if err1 != nil {
			fmt.Printf("err: DefaultPrintHttpLogInfo fmt.Fprintf:%s\n", err1.Error())
		}
		return
	}
	_, err1 := fmt.Fprintf(logchan.LogWriter, "%s|curl:%s|response:%s\n", logchan.DefaultPrintLog(httpLogInfo), httpLogInfo.CurlCmd, httpLogInfo.ResponseBody)
	if err1 != nil {
		fmt.Printf("err: DefaultPrintHttpLogInfo fmt.Fprintf:%s\n", err1.Error())
	}
}
