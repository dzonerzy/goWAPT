package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
)

var srvCloser io.Closer
var received bool = false
var globalConfig *Configuration

func badConfig() {
	os.Exit(-1)
}

func formatRequest(r *http.Request) string {
	var request string
	r.URL.Host = ""
	r.URL.Scheme = ""
	url := fmt.Sprintf("%v %v %v\n", r.Method, r.URL, r.Proto)
	request += url
	request += fmt.Sprintf("Host: %v\n", r.Host)
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request += fmt.Sprintf("%v: %v\n", name, h)
		}
	}
	if r.Method == "POST" {
		b, _ := ioutil.ReadAll(r.Body)
		request += "\n"
		request += string(b)
	} else {
		request += "\n"
	}
	return request
}

func NonProxyHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme == "https" {
		globalConfig.ssl = true
	} else {
		globalConfig.ssl = false
	}
	data := formatRequest(req)
	if !strings.Contains(data, "FUZZ") {
		fmt.Println("Error: When using proxy mode a keyword FUZZ must be specified inside request.")
		badConfig()
	} else {
		globalConfig.templateData = data
	}
	http.Error(w, "Request received by GOWAPT (Transparent)", 200)
	received = true
}

func handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	if req.URL.Scheme == "https" {
		globalConfig.ssl = true
	} else {
		globalConfig.ssl = false
	}
	data := formatRequest(req)
	if !strings.Contains(data, "FUZZ") {
		fmt.Println("Error: When using proxy mode a keyword FUZZ must be specified inside request.")
		badConfig()
	} else {
		globalConfig.templateData = data
	}
	return req, goproxy.NewResponse(req,
		goproxy.ContentTypeText, http.StatusOK,
		"Request received by GOWAPT")
}

func handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	received = true
	return resp
}

func ListenAndServeWithClose(addr string, handler http.Handler) (sc io.Closer, err error) {
	var listener net.Listener
	srv := &http.Server{Addr: addr, Handler: handler}
	if addr == "" {
		addr = ":http"
	}
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() {
		srv.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)})
	}()
	return listener, nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func checkConfig(c *Configuration) Configuration {
	globalConfig = c
	if c.url == "" && c.template == "" && c.from_proxy == false {
		fmt.Println("Error: At least an url or a template must be specified.")
		badConfig()
	}

	if c.url != "" && c.template != "" && c.from_proxy == true {
		fmt.Println("Error: You must define only one option.")
		badConfig()
	}
	if c.template != "" && c.postdata != "" {
		fmt.Println("Error: Template option should include POST data in template.")
		badConfig()
	}

	if c.filter != "" {
		r, _ := regexp.Compile("(?P<line>[a-z]+)(\\s{0,100})(?P<op>(<|>|\\|\\||==|!=))(\\s{0,100})(?P<data>[-\\d]+)")
		if !r.Match([]byte(c.filter)) {
			fmt.Println("Error: Bad filter expression.")
			badConfig()
		}
	}

	if c.upstream_proxy != "" {
		proxy_url, err := url.ParseRequestURI(c.upstream_proxy)

		if err != nil {
			fmt.Println("Error: Please use a valid upstream proxy url.")
			badConfig()
		}
		c.upstream_url = proxy_url
	}

	if c.url != "" && c.ssl == false {
		if strings.Contains(c.url, "http://") {
			c.ssl = false
		} else if strings.Contains(c.url, "https://") {
			c.ssl = true
		}
	}

	if c.url != "" && !strings.HasPrefix(c.url, "http") {
		fmt.Println("Error: Invalid protocol specified!")
		badConfig()
	}

	if c.template != "" && c.url == "" {
		if _, err := os.Stat(c.template); os.IsNotExist(err) {
			fmt.Println("Error: Template file does not exist.")
			badConfig()
		} else {
			_, err := os.Open(c.template)
			if err != nil {
				fmt.Println("Error: You don't have read permission on template file.")
				badConfig()
			}
		}
	}

	if c.url != "" {
		have_keyword := false
		have_keyword = have_keyword || strings.Contains(c.url, "FUZZ")
		have_keyword = have_keyword || strings.Contains(c.postdata, "FUZZ")
		have_keyword = have_keyword || strings.Contains(c.cookies, "FUZZ")
		for _, v := range c.headers {
			have_keyword = have_keyword || strings.Contains(v, "FUZZ")
		}
		if !have_keyword {
			fmt.Println("Error: Neither url nor post data nor cookie nor headers contains FUZZ keyword.")
			badConfig()
		}
	}

	if c.template != "" {
		data, _ := ioutil.ReadFile(c.template)
		if !strings.Contains(string(data), "FUZZ") {
			fmt.Println("Error: When using template a keyword FUZZ must be specified.")
			badConfig()
		} else {
			c.templateData = string(data) + "\r\n"
		}
	}

	if c.usefuzzer && c.wordlist == "" {
		fmt.Println("Error: Fuzzer mode require a wordlist.")
		badConfig()
	}

	if !c.usefuzzer && c.wordlist == "" {
		fmt.Println("Error: Please specify a wordlist")
		badConfig()
	}

	if c.wordlist != "" {
		if _, err := os.Stat(c.wordlist); os.IsNotExist(err) {
			fmt.Println("Error: Wordlist file does not exist.")
			badConfig()
		} else {
			_, err := os.Open(c.wordlist)
			if err != nil {
				fmt.Println("Error: You don't have read permission on wordlist file.")
				badConfig()
			}
		}
	}

	initEncoders()

	if c.extension != "" {
		if _, err := os.Stat(c.extension); os.IsNotExist(err) {
			fmt.Println("Error: Extension file does not exist.")
			badConfig()
		} else {
			_, err := os.Open(c.extension)
			if err != nil {
				fmt.Println("Error: You don't have read permission on extension file.")
				badConfig()
			}
		}
		initExtender()
		runExtension(c.extension)
	}

	for _, enc := range strings.Split(c.encoders, ",") {
		enc = strings.Trim(enc, " ")
		if !strings.Contains(enc, "@") {
			if _, ok := encoders[enc]; ok {
				c.encoderList = append(c.encoderList, encoders[enc])
			} else {
				fmt.Println(fmt.Sprintf("Error: Unable to find encoder named '%s'.", enc))
				badConfig()
			}
		} else {
			encs := strings.Split(enc, "@")
			if len(encs) == 2 {
				if _, oke1 := encoders[encs[0]]; oke1 {
					if _, oke2 := encoders[encs[1]]; oke2 {
						f := func(x string) string {
							return encoders[encs[1]](encoders[encs[0]](x))
						}
						c.encoderList = append(c.encoderList, Encoder(f))
					}
				} else {
					fmt.Println(fmt.Sprintf("Error: Unable to find encoder named '%v'.", enc))
					badConfig()
				}
			} else {
				fmt.Println("Error: Maximum encoder chaining is 2 functions")
				badConfig()
			}

		}
	}

	if c.auth != "" {
		c.auth = "Basic " + b64(c.auth)
	}

	if c.from_proxy && c.url == "" && c.template == "" {
		proxy := goproxy.NewProxyHttpServer()
		proxy.NonproxyHandler = http.HandlerFunc(NonProxyHandler)
		proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).HandleConnect(goproxy.AlwaysMitm)
		proxy.OnRequest().DoFunc(handleRequest)
		proxy.OnResponse().DoFunc(handleResponse)
		fmt.Println("[*] Receiving requests on [127.0.0.1:31337]")
		srvCloser, _ = ListenAndServeWithClose(":31337", proxy)
		for !received {
		}
		time.Sleep(time.Second * 1)
		srvCloser.Close()
		c.template = "** FROM PROXY **"
	}

	max_concurrency = c.threads

	return *c
}
