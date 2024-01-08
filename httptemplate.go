package httptemplate

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
)

type HttpTpl interface {
	Parse(data interface{}) (rawHttp string, err error)
	Request(data interface{}) (r *http.Request, err error)
}

const (
	Window_EOF           = "\r\n"
	Linux_EOF            = "\n"
	HTTP_HEAD_BODY_DELIM = Window_EOF + Window_EOF
)

type httpTpl struct {
	Tpl      string
	template *template.Template
}

//NewHttpTpl 实例化模版请求
func NewHttpTpl(tpl string) (HttpTpl, error) {
	// 检测模板是否符合 http 协议
	req, err := ReadRequest(tpl)
	if err != nil {
		return nil, err
	}
	req.Header.Del("Content-Length") // 实际请求需要重新计算长度
	// 生成统一符合http 协议规范的模板
	reqByte, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	htPt := &httpTpl{
		Tpl: string(reqByte),
	}

	t, err := template.New("").Funcs(sprig.FuncMap()).Funcs(TemplatefuncMap).Parse(htPt.Tpl)
	if err != nil {
		return nil, err
	}
	htPt.template = t
	return htPt, nil
}

//Request 解析模板，生成http raw 协议文本
func (htPt *httpTpl) Parse(data interface{}) (rawHttp string, err error) {
	var b bytes.Buffer
	err = htPt.template.Execute(&b, data)
	if err != nil {
		return
	}
	rawHttp = b.String()
	return rawHttp, nil
}

//Request 解析模板，生成http raw 协议文本
func (htPt *httpTpl) Request(data interface{}) (r *http.Request, err error) {
	rawHttp, err := htPt.Parse(data)
	if err != nil {
		return nil, err
	}
	r, err = ReadRequest(rawHttp)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//ReadRequest http 文本协议格式转http.Request 对象
func ReadRequest(httpRaw string) (req *http.Request, err error) {
	httpRaw = funcs.TrimSpaces(httpRaw)
	lineArr := strings.Split(httpRaw, Linux_EOF)
	formatLineArr := make([]string, 0)
	for _, line := range lineArr {
		formatLine := strings.TrimSpace(line) // 去除每行的空格、制表符\r 等符号
		formatLineArr = append(formatLineArr, formatLine)
	}
	httpRaw = strings.Join(formatLineArr, Window_EOF)
	if httpRaw == "" {
		err = errors.Errorf("http raw is empty")
		return nil, err
	}

	headerRaw := strings.TrimSpace(httpRaw) // 默认只有请求头
	bodyRaw := ""                           // 默认body为空
	bodyIndex := strings.Index(headerRaw, HTTP_HEAD_BODY_DELIM)
	if bodyIndex > -1 {
		headerRaw, bodyRaw = strings.TrimSpace(headerRaw[:bodyIndex]), strings.TrimSpace(headerRaw[bodyIndex:])
		bodyLen := len(bodyRaw)
		headerRaw = fmt.Sprintf("%s%sContent-Length: %d", headerRaw, Window_EOF, bodyLen)
	}
	formatHttpRaw := fmt.Sprintf("%s%s%s", headerRaw, HTTP_HEAD_BODY_DELIM, bodyRaw)

	buf := bufio.NewReader(strings.NewReader(formatHttpRaw))
	req, err = http.ReadRequest(buf)
	if err != nil {
		return
	}
	if req.URL.Scheme == "" {
		queryPre := ""
		if req.URL.RawQuery != "" {
			queryPre = "?"
		}
		req.RequestURI = fmt.Sprintf("http://%s%s%s%s", req.Host, req.URL.Path, queryPre, req.URL.RawQuery)
	}

	return req, nil
}

type RequestDTO struct {
	URL     string         `json:"url"`
	Method  string         `json:"method"`
	Header  http.Header    `json:"header"`
	Cookies []*http.Cookie `json:"cookies"`
	Body    string         `json:"body"`
}

//Request2RequestDTO 将 http.Request 转换为 request 结构体，方便将http raw 转换为常见的构造http请求参数
func Request2RequestDTO(req *http.Request) (requestDTO *RequestDTO, err error) {
	requestDTO = &RequestDTO{}
	bodyReader, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	bodyByte, err := io.ReadAll(bodyReader)
	if err != nil {
		return
	}
	req.Header.Del("Content-Length")
	requestDTO = &RequestDTO{
		URL:     req.URL.String(),
		Method:  req.Method,
		Header:  req.Header,
		Cookies: req.Cookies(),
		Body:    string(bodyByte),
	}

	return requestDTO, nil
}

type ResponseDTO struct {
	HttpStatus  string         `json:"httpStatus"`
	Header      http.Header    `json:"header"`
	Cookies     []*http.Cookie `json:"cookies"`
	Body        string         `json:"body"`
	RequestData *RequestDTO    `json:"requestData"`
}

func ParseResponse(b []byte, r *http.Request) (responseDTO *ResponseDTO, err error) {
	byteReader := bytes.NewReader(b)
	reader := bufio.NewReader(byteReader)
	rsp, err := http.ReadResponse(reader, r)
	if err != nil {
		return nil, err
	}
	reqData := new(RequestDTO)
	if r != nil {
		reqData, err = Request2RequestDTO(r)
		if err != nil {
			return nil, err
		}
	}
	responseDTO = &ResponseDTO{
		HttpStatus:  strconv.Itoa(rsp.StatusCode),
		Header:      rsp.Header,
		Cookies:     rsp.Cookies(),
		RequestData: reqData,
	}
	return responseDTO, nil
}
