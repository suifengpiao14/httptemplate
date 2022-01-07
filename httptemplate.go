package httptemplate

import (
	"bufio"
	"bytes"
	"net/http"
	"text/template"
)

type HttpTpl interface {
	Request() (req *http.Request,err error)
	Cause () (httpTpl *httpTpl)
}

type httpTpl struct {
	Tpl      string
	Data     map[string]interface{}
	template *template.Template
	request   *http.Request
	iter *int
}
//New 实例化模版请求
func New(tpl string, data map[string]interface{},iter *int,funcMap template.FuncMap) ( HttpTpl,error) {
	t,err:=template.New("").Funcs(funcMap).Parse(tpl)
	if err !=nil{
		return nil,err
	}
	htPt:=&httpTpl{
		Tpl:      tpl,
		template: t,
		Data:     data,
		iter: iter,
	}
	return htPt,nil
}

//Request 解析模板，生成请求对象
func (htPt *httpTpl) Request() ( *http.Request,  error) {
	req ,err:= htPt.parseTpl()
	if err !=nil{
		return nil,err
	}
	return req,nil
}

func (htPt *httpTpl)increase(key string)  {

}

func (htPt *httpTpl) parseTpl() (req *http.Request, err error) {
	var b  *bytes.Buffer
	err = htPt.template.Execute(b,htPt.Data)
	if err !=nil{
		return
	}
	buf:=bufio.NewReader(b)
	req1, err := http.ReadRequest(buf)
	if err !=nil{
		return
	}
	req ,err= http.NewRequest(req1.Method,req1.URL.String(),req1.Body)
	req.Header=req1.Header
	return
}

func (htPt *httpTpl) Cause() *httpTpl {
	if htPt.iter ==nil{
		return nil
	}
	// 包含递增，生成递归对象
	cause:=&httpTpl{
		Tpl:      htPt.Tpl,
		template: htPt.template,
		Data:     htPt.Data,
		iter: htPt.iter,
	}
	return cause
}
