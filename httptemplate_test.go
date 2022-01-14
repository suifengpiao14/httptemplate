package httptemplate

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestNew(t *testing.T) {
	tpl := `
	POST / HTTP/1.1
	Host: new-merchant-api.hsb.com
	Content-Type: application/json
	HSB-OPENAPI-CALLERSERVICEID: 214001
	HSB-OPENAPI-SIGNATURE: 767a9cd8148fc5bc460c16372fbac532



	{"_head":{"_interface":"NewMerchantCenterServer.Api.V1.getMerchantInfo","_msgType":"request","_remark":"","_version":"0.01","_timestamps":"1439261904","_invokeId":"563447634257324435","_callerServiceId":"210015","_groupNo":"1"},"_param":{"merchantId":"{{.merchantId}}","queryType":"{{.queryType}}"}}
	`

	data := map[string]string{
		"merchantId": "141218",
		"queryType":  "businessInfo",
	}

	httpTpl, err := NewHttpTpl(tpl, TemplatefuncMap)
	if err != nil {
		panic(err)
	}
	req, err := httpTpl.ReadOnlyRequest(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", req)
	//req1.Method, req1.URL.String(), req1.Body
	request := resty.New().R()
	request.URL = req.URL.String()
	request.Method = req.Method
	request.Header = req.Header
	request.SetCookies(req.Cookies())
	request.SetBody(req.Body)
	resp, err := request.Send()
	if err != nil {
		panic(err)
	}
	if resp.IsError() {
		fmt.Print(resp.Error())
	}

	if resp.IsSuccess() {
		body := resp.Body()
		fmt.Println(string(body))
	}
}
