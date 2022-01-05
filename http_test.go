package http_pro

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	h := GetHttp(1, 0)
	url := "http://172.16.88.35:18888/userReq/doLogin"
	method := "POST"
	payload := strings.NewReader("loginEmail=admin&password=admin")
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := h.Request(req, context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	body, err := GetStringResponseBody(res)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)
}
