package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/robertkrimen/otto"
)

func initExtender() {
	JSVM.Set("addCustomEncoder", func(name string, f otto.Value) {
		encoders[name] = Encoder(func(x string) string {
			v, _ := f.Call(otto.NullValue(), x)
			sv, _ := v.ToString()
			return sv
		})
	})

	JSVM.Set("setHTTPInterceptor", func(f otto.Value) {
		haveHTTPInterceptor = true
		var v otto.Value
		callback := func(x interface{}, y interface{}, b bool) {
			if b {
				v, _ = f.Call(otto.NullValue(), x, otto.NullValue(), otto.TrueValue())
			} else {
				v, _ = f.Call(otto.NullValue(), x, y, otto.FalseValue())
			}
		}
		HTTPInterceptor = JSHTTPInterceptor(callback)
	})

	JSVM.Set("dumpResponse", func(req_resp interface{}, path string) otto.Value {
		switch obj := req_resp.(type) {
		case *http.Response:
			content := response2String(obj)
			err := ioutil.WriteFile(path, content, 0644)
			if err != nil {
				return otto.FalseValue()
			}
			return otto.TrueValue()
		default:
			return otto.FalseValue()
		}
	})

	JSVM.Set("sendRequestSync", func(method otto.Value, http_url otto.Value, post_data otto.Value, headers interface{}) interface{} {
		method_str, m_err := method.ToString()
		url_str, u_err := http_url.ToString()
		post_str, p_err := post_data.ToString()
		if m_err == nil && u_err == nil && p_err == nil {
			if post_str != "null" {
				post_str = "POST"
			} else {
				post_str = ""
			}
			tmp_url, url_err := url.ParseRequestURI(url_str)
			tmp_post, post_err := url.ParseQuery(post_str)
			tmp_req, req_err := http.NewRequest(method_str, tmp_url.String(), strings.NewReader(tmp_post.Encode()))
			if req_err == nil && post_err == nil && url_err == nil {
				for v, k := range headers.(map[string]interface{}) {
					tmp_req.Header.Add(v, k.(string))
				}
				response, err := netClient.Do(tmp_req)
				if err == nil {
					return response
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			return nil
		}
	})

	JSVM.Set("panic", func(v interface{}) {
		panic(fmt.Sprintf("%v", v))
	})
}

func runExtension(file string) {
	script, _ := ioutil.ReadFile(file)
	_, err := JSVM.Run(string(script))
	if err != nil {
		fmt.Printf("JS Error: %v\n", err)
		badConfig()
	}
}
