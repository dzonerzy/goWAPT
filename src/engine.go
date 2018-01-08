package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var stopFuzzingEngine chan bool
var stopEngine chan bool
var videoUpdateChan chan bool
var netTransport = &http.Transport{
	MaxIdleConns:        200,
	MaxIdleConnsPerHost: 200,
	Dial: (&net.Dialer{
		Timeout: 1 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 1 * time.Second,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	DisableCompression:  true,
}
var netClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout:   time.Second * 5,
	Transport: netTransport,
}
var req *http.Request
var extenderChan chan []interface{}
var extenderStopChan chan bool
var nextRequest chan bool
var concurrencyChan chan int
var finishedChan chan bool
var quitConcurrency chan bool
var endDispose chan bool
var stopRequest chan bool
var stoppedViaStop bool = false
var concurrency int
var extenderFuncCompleted chan bool

func startFuzzEngine(cfg *Configuration, res *[]Result) {
	stopFuzzingEngine = make(chan bool)
	stopEngine = make(chan bool)
	go runEngine(cfg, res)

	for {
		select {
		case res := <-stopFuzzingEngine:
			if res == true {
				stopEngine <- true
				return
			}
		}
	}
}

func extenderPooler() {
	for {
		select {
		case <-extenderStopChan:
			return
		case r := <-extenderChan:
			HTTPInterceptor(r[0], r[1], r[2].(bool))
			extenderFuncCompleted <- true
		}
	}
}

func Request(ch chan Result, client *http.Client, req *http.Request, payload string, req_num int, stopch chan bool, reqBodyString string, concChan chan int) {
	var err error
	var response *http.Response
	var done bool = false
	var rdrq1 io.ReadCloser
	var rdrq2 io.ReadCloser
	for !done {
		select {
		case s := <-stopch:
			if s == true {
				return
			}
		default:
			if len(reqBodyString) > 0 {
				reqb, _ := ioutil.ReadAll(req.Body)
				rdrq1 = ioutil.NopCloser(bytes.NewBuffer(reqb))
				rdrq2 = ioutil.NopCloser(bytes.NewBuffer(reqb))
				req.Body = rdrq1
			}
			response, err = client.Do(req)
			if err == nil {
				result := Result{}
				defer response.Body.Close()
				result.payload = payload
				result.stat = &Stats{}
				result.stat.code = response.StatusCode

				var reader io.ReadCloser

				switch response.Header.Get("Content-Encoding") {
				case "gzip":
					reader, err = gzip.NewReader(response.Body)
					response.Header.Set("Content-Encoding", "identity")
					defer reader.Close()
				default:
					reader = response.Body
				}

				bodyBytes, _ := ioutil.ReadAll(reader)
				rdr1 := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
				rdr2 := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
				if len(reqBodyString) > 0 {
					response.Request.Body = rdrq2
				}
				response.Body = rdr2
				bodyString := string(bodyBytes)
				htmlreader := html.NewTokenizer(rdr1)
				tags := 0
			tagLoop:
				for {
					tt := htmlreader.Next()
					switch tt {
					case html.ErrorToken:
						break tagLoop
					case html.StartTagToken, html.EndTagToken:
						tags++
					}
				}
				result.stat.chars = len(bodyString)
				result.stat.words = len(strings.Split(bodyString, " "))
				result.stat.lines = len(strings.Split(bodyString, "\n"))
				result.stat.tags = tags
				cl, _ := strconv.Atoi(response.Header.Get("Content-Length"))
				result.stat.length = cl
				s_req, _ := httputil.DumpRequest(req, false)
				s_res, _ := httputil.DumpResponse(response, false)
				result.request_number = req_num
				result.request = string(s_req) + reqBodyString + "\n\n"
				result.response = string(s_res) + bodyString
				result.request_response = result.request + result.response
				if haveHTTPInterceptor {

					tmp_res := map[string]interface{}{"tags": result.stat.tags,
						"code": result.stat.code, "words": result.stat.words,
						"chars": result.stat.chars, "lines": result.stat.lines,
						"payload": result.payload, "request": result.request,
						"response": result.response}
					tmp_response := *response
					r := []interface{}{&tmp_response, tmp_res, false}
					extenderChan <- r
					select {
					case <-extenderFuncCompleted:

					}
				}
				ch <- result
				nextRequest <- true
				done = true
			} else {
				time.Sleep(time.Millisecond * 500)
			}
		}

	}
}

func Dispose(res *[]Result, ch chan Result, until int, finish chan bool, end chan bool) {
	for i := 0; i < until; i++ {
		select {
		case r := <-ch:
			initFilters(config.filter, r.stat)
			if checkFilter() {
				*res = append(*res, r)
				videoUpdateChan <- true
			}
		case k := <-end:
			if k == true {
				finish <- true
				return
			}
		}
	}
	finish <- true
	return
}

func waitTillEnd(finished chan bool) {
endLoop:
	for {
		select {
		case r := <-finished:
			if r == true {
				break endLoop
			}
		}
	}
}

func runEngine(cfg *Configuration, res *[]Result) {
	extenderFuncCompleted = make(chan bool)
	stoppedViaStop = false
	stopRequest = make(chan bool)
	endDispose = make(chan bool)
	quitConcurrency = make(chan bool)
	finishedChan = make(chan bool)
	concurrencyChan = make(chan int)
	nextRequest = make(chan bool)
	extenderChan = make(chan []interface{})
	extenderStopChan = make(chan bool)
	if haveHTTPInterceptor {
		go extenderPooler()
	}
	if cfg.upstream_proxy != "" {
		netTransport.Proxy = http.ProxyURL(cfg.upstream_url)
	}
	videoUpdateChan = make(chan bool)
	rand.Seed(time.Now().Unix())
	var fuzzed_data string = ""
	var tmp_url string = ""
	var reqBody string = ""
	concurrency = 0
	wordlist, _ := ioutil.ReadFile(cfg.wordlist)
	str_wordlist := string(wordlist)
	slice_wordlist := strings.Split(str_wordlist, "\n")
	max := len(slice_wordlist)
	n_encoders := len(cfg.encoderList)
	if cfg.usefuzzer {
		max *= 10
	}
	max_wordlist_encoders := max * n_encoders
	current_request := 0
	channel := make(chan Result)
	go Dispose(res, channel, max_wordlist_encoders, finishedChan, endDispose)
engineLoop:
	for k := 0; k < n_encoders; k++ {
		for i := 0; i < max; i++ {
			select {
			case r := <-stopEngine:
				if r == true {
					stoppedViaStop = true
					endDispose <- true
					break engineLoop
				}
			case val := <-concurrencyChan:
				concurrency = val
			default:
				current_request++

				//rnd_encoder_num := rand.Intn(len(cfg.encoderList))
				if cfg.usefuzzer {
					if slice_wordlist[i%len(slice_wordlist)] != "" {
						fuzzed_data = cfg.encoderList[k](fuzz(slice_wordlist[i%len(slice_wordlist)]))
					} else {
						fuzzed_data = cfg.encoderList[k](fuzz("/'\"><img>\" or 1 = 1/*/../../etc/passwd"))
					}
				} else {
					fuzzed_data = cfg.encoderList[k](slice_wordlist[i])
				}
				if cfg.url != "" {
					tmp_url = strings.Replace(cfg.url, "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD)
					fuzzurl, _ := url.ParseRequestURI(tmp_url)
					if cfg.postdata == "" {
						req, _ = http.NewRequest("GET", fuzzurl.String(), nil)
					} else {
						var data url.Values
						reqBody = strings.Replace(cfg.postdata, "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD)
						data, _ = url.ParseQuery(reqBody)
						req, _ = http.NewRequest("POST", fuzzurl.String(), strings.NewReader(data.Encode()))
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
					}
					if cfg.cookies != "" {
						req.Header.Add("Cookie", strings.Replace(cfg.cookies, "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD))
					}
					if len(cfg.headers) > 0 {
						for _, header := range cfg.headers {
							header_vals := strings.SplitN(header, ":", 2)
							if len(header_vals) == 2 {
								header_fuzz := strings.Replace(header_vals[1], "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD)
								req.Header.Add(strings.Replace(header_vals[0], ":", "", 1), header_fuzz)
							}
						}
					}
					if cfg.ssl {
						req.URL.Scheme = "https"
					} else {
						req.URL.Scheme = "http"
					}
				} else if cfg.templateData != "" {
					tmp_data := strings.Replace(cfg.templateData, "FUZZ", url.QueryEscape(fuzzed_data), 1)
					tmp_req, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(tmp_data)))
					tmp_req.URL.Scheme = ""
					tmp_req.URL.Host = ""
					var full_url string = ""
					if cfg.ssl {
						full_url = fmt.Sprintf("%s%s%s", "https://", tmp_req.Host, tmp_req.URL.String())
					} else {
						full_url = fmt.Sprintf("%s%s%s", "http://", tmp_req.Host, tmp_req.URL.String())
					}
					full_url_parsed, _ := url.ParseRequestURI(full_url)
					body, _ := ioutil.ReadAll(tmp_req.Body)
					reqBody = string(body)
					req, _ = http.NewRequest(tmp_req.Method, full_url_parsed.String(), strings.NewReader(reqBody))
					req.Header = tmp_req.Header
				}
				if cfg.auth != "" {
					req.Header.Set("Authentication", cfg.auth)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
				req.Header.Set("Content-Encoding", "identity")
				req.Header.Set("Connection", "keep-alive")
				if haveHTTPInterceptor {
					r := []interface{}{req, nil, true}
					extenderChan <- r
					select {
					case <-extenderFuncCompleted:
					}
				}
				percentage = (100 * current_request) / max_wordlist_encoders
				if concurrency < max_concurrency {
					concurrency += 1
					go Request(channel, netClient, req, fuzzed_data, current_request, stopRequest, reqBody, concurrencyChan)
				} else {
					select {
					case r := <-nextRequest:
						if r == true {
							go Request(channel, netClient, req, fuzzed_data, current_request, stopRequest, reqBody, concurrencyChan)
						}
					}
				}

			}
		}
	}
	waitTillEnd(finishedChan)
	if haveHTTPInterceptor {
		extenderStopChan <- true
	}
	if !stoppedViaStop {
		stopFuzzingEngine <- true
	}
	started = false
	percentage = (100 * max_wordlist_encoders) / max_wordlist_encoders
	videoUpdateChan <- true
	return
}
