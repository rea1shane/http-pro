package http_pro

import (
	"context"
	"encoding/json"
	"github.com/morikuni/failure"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Http struct {
	client    *http.Client
	attempts  int
	sleepTime time.Duration // sleepTime will increase with attempt times, value = sleepTime * (2 * times - 1)
}

func GetHttp(attempts int, sleepTime time.Duration) Http {
	if attempts < 1 {
		panic("Wrong value of attempts: " + strconv.Itoa(attempts) + ", should >= 1")
	}
	if sleepTime < 0 {
		panic("Wrong value of sleep time: " + sleepTime.String() + ", should >= 0")
	}
	return Http{
		client:    &http.Client{},
		attempts:  attempts,
		sleepTime: sleepTime,
	}
}

// Request 如果不做 context 控制, 传入 context.Background 即可
func (h *Http) Request(req *http.Request, ctx context.Context) (*http.Response, error) {
	req = req.WithContext(ctx)
	escape(req)
	res, err := h.attemptDo(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Http) attemptDo(req *http.Request) (*http.Response, error) {
	var (
		res *http.Response
		err = map[int]error{}
	)
	for i := 0; i < h.attempts; i++ {
		res, err[i] = h.client.Do(req)
		if err[i] != nil {
			if i == h.attempts-1 {
				c := getReqFailureContext(req)
				for k := 0; k < h.attempts-1; k++ {
					c["attempt "+strconv.Itoa(k+1)+" err"] = err[k].Error()
				}
				return res, failure.Wrap(err[i], c)
			} else {
				time.Sleep(h.sleepTime * time.Duration(2*i+1))
				continue
			}
		}
		break
	}
	return res, nil
}

func GetStringResponseBody(res *http.Response) (string, error) {
	bytes, err := readResponseBody(res)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func GetStructResponseBody(res *http.Response, responseStruct interface{}) error {
	bytes, err := readResponseBody(res)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &responseStruct); err != nil {
		return failure.Wrap(err, getResFailureContext(res, bytes))
	}
	return nil
}

func escape(req *http.Request) {
	req.URL.RawQuery = url.PathEscape(req.URL.RawQuery)
}

func readResponseBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, failure.Wrap(err, getResFailureContext(res, responseBody))
	}
	return responseBody, nil
}

func getReqFailureContext(req *http.Request) failure.Context {
	return failure.Context{
		"protocol":    req.Proto,
		"host":        req.URL.Hostname(),
		"port":        req.URL.Port(),
		"request_url": req.URL.RequestURI(),
	}
}

func getResFailureContext(res *http.Response, body []byte) failure.Context {
	c := getReqFailureContext(res.Request)
	c["response_body"] = string(body)
	return c
}
