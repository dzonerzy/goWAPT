package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

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
			} else {
				return otto.TrueValue()
			}
		default:
			return otto.FalseValue()
		}
	})

	JSVM.Set("panic", func(v string) {
		panic(v)
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
