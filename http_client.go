package httptemplate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/logchan/v2"
)

// RequestFn 封装http请求数据格式
type RequestFn func(ctx context.Context, req *http.Request, transport *http.Transport) (out []byte, err error)

// RestyRequestFn 通用请求方法
func RestyRequestFn(ctx context.Context, req *http.Request, transport *http.Transport) (out []byte, err error) {
	client := resty.New()
	if transport != nil {
		client.SetTransport(transport)
	}
	r := resty.New().R()
	urlstr := req.URL.String()
	r.Header = req.Header
	r.FormData = req.Form
	r.RawRequest = req
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		r.SetBody(body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	logInfo := &LogInfoHttp{
		GetRequest: func() *http.Request { return r.RawRequest },
	}
	defer func() {
		logchan.SendLogInfo(logInfo)
	}()
	res, err := r.Execute(strings.ToUpper(req.Method), urlstr)
	if err != nil {
		return nil, err
	}

	responseBody := res.Body()
	if !res.IsSuccess() {
		err = errors.Errorf("http status:%d,body:%s", res.StatusCode(), string(responseBody))
		err = errors.WithMessage(err, fmt.Sprintf("%v", res.Error()))
		return nil, err
	}
	logInfo.ResponseBody = string(responseBody)
	logInfo.Response = res.RawResponse
	return responseBody, nil
}

var CURL_TIMEOUT = 30 * time.Millisecond

type TransportConfig struct {
	Proxy               string `json:"proxy"`
	Timeout             int    `json:"timeout"`
	KeepAlive           int    `json:"keepAlive"`
	MaxIdleConns        int    `json:"maxIdleConns"`
	MaxIdleConnsPerHost int    `json:"maxIdleConnsPerHost"`
	IdleConnTimeout     int    `json:"idleConnTimeout"`
}

//NewTransport 创建一个htt连接,兼容代理模式
func NewTransport(cfg *TransportConfig) *http.Transport {
	maxIdleConns := 200
	maxIdleConnsPerHost := 20
	idleConnTimeout := 90
	if cfg.MaxIdleConns > 0 {
		maxIdleConns = cfg.MaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost > 0 {
		maxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	}
	if cfg.IdleConnTimeout > 0 {
		idleConnTimeout = cfg.IdleConnTimeout
	}
	timeout := 10
	if cfg.Timeout > 0 {
		timeout = 10
	}
	keepAlive := 300
	if cfg.KeepAlive > 0 {
		keepAlive = cfg.KeepAlive
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,   // 连接超时时间
			KeepAlive: time.Duration(keepAlive) * time.Second, // 连接保持超时时间
		}).DialContext,
		MaxIdleConns:        maxIdleConns,                                 // 最大连接数,默认0无穷大
		MaxIdleConnsPerHost: maxIdleConnsPerHost,                          // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
		IdleConnTimeout:     time.Duration(idleConnTimeout) * time.Second, // 多长时间未使用自动关闭连
	}
	if cfg.Proxy != "" {
		proxy, err := url.Parse(cfg.Proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	return transport
}
