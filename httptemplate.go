package httptemplate

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"text/template"
)

type HttpTpl interface {
	Request(data interface{}) (req *http.Request, err error)
}

const (
	EOF = "\r\n"
)

type httpTpl struct {
	Tpl      string
	template *template.Template
	request  *http.Request
}

//NewHttpTpl 实例化模版请求
func NewHttpTpl(tpl string, funcMap template.FuncMap) (HttpTpl, error) {

	tpl = strings.TrimSpace(tpl)
	lineArr := strings.Split(tpl, "\n")
	formatLineArr := make([]string, 0)
	for _, line := range lineArr {
		formatLine := strings.TrimSpace(line)
		if strings.Index(strings.ToLower(formatLine), "content-length:") == 0 {
			continue // 因为动态计算请求体长度，所以删除模版中的content-length: 头
		}
		formatLineArr = append(formatLineArr, formatLine)
	}
	formatTpl := strings.Join(formatLineArr, EOF)
	htPt := &httpTpl{
		Tpl: formatTpl,
	}
	// 检测模板是否符合 http 协议
	req, err := htPt.ReadRequest(formatTpl)
	if err != nil {
		return nil, err
	}

	if req.URL.Scheme == "" {
		queryPre := ""
		if req.URL.RawQuery != "" {
			queryPre = "?"
		}
		req.RequestURI = fmt.Sprintf("http://%s%s%s%s", req.Host, req.URL.Path, queryPre, req.URL.RawQuery)
		req.Header.Del("Content-Length") // 实际请求需要重新计算长度
		reqByte, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		htPt.Tpl = string(reqByte)
	}
	t, err := template.New("").Funcs(funcMap).Parse(htPt.Tpl)
	if err != nil {
		return nil, err
	}
	htPt.template = t

	return htPt, nil
}

//Request 解析模板，生成请求对象
func (htPt *httpTpl) Request(data interface{}) (req *http.Request, err error) {
	var b bytes.Buffer
	err = htPt.template.Execute(&b, data)
	if err != nil {
		return
	}
	httpRaw := b.String()
	req, err = htPt.ReadRequest(httpRaw)
	if err != nil {
		return nil, err
	}
	return
}

//ReadRequest 解析模板
func (htPt *httpTpl) ReadRequest(httpRaw string) (req *http.Request, err error) {
	bodyIndex := strings.LastIndex(httpRaw, EOF)
	headerRaw := strings.TrimSpace(httpRaw[:bodyIndex])
	bodyRaw := httpRaw[bodyIndex+len(EOF):]
	bodyLen := len(bodyRaw)
	formatHttpRaw := fmt.Sprintf("%s%sContent-Length: %d%s%s%s", headerRaw, EOF, bodyLen, EOF, EOF, bodyRaw)
	buf := bufio.NewReader(strings.NewReader(formatHttpRaw))
	req, err = http.ReadRequest(buf)
	if err != nil {
		return
	}
	return
}
