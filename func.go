package httptemplate

import (
	"strings"
	"text/template"
	"time"
)

var TemplatefuncMap = template.FuncMap{
	"zeroTime":      ZeroTime,
	"currentTime":   CurrentTime,
	"permanentTime": PermanentTime,
	"Contains":      strings.Contains,
}

func ZeroTime() string {
	return "0000-00-00 00:00:00"
}

func CurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func PermanentTime() string {
	return "3000-12-31 23:59:59"
}
