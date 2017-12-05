package main

import (
	"bufio"
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
	"sync"
	"time"

	"golang.org/x/net/html"
)

var stopFuzzingEngine chan bool
var stopEngine chan bool
var videoUpdateChan chan bool
var concurrency int
var netTransport = &http.Transport{
	MaxIdleConnsPerHost: 20,
	Dial: (&net.Dialer{
		Timeout: 2 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 2 * time.Second,
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
var wg sync.WaitGroup

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

func Request(wg *sync.WaitGroup, ch chan Result, client *http.Client, req *http.Request, payload string, video chan bool, req_num int, stopch chan bool, reqBodyString string) {
	var err error
	var response *http.Response
	var done bool = false
	for !done {
		select {
		case s := <-stopch:
			if s == true {
				return
			}
		default:
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
					defer reader.Close()
				default:
					reader = response.Body
				}

				bodyBytes, _ := ioutil.ReadAll(reader)
				bodyString := string(bodyBytes)
				htmlreader := html.NewTokenizer(strings.NewReader(bodyString))
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
				wg.Done()
				ch <- result
				video <- true
				done = true
			} else {
				time.Sleep(time.Millisecond * 1000)
			}
		}

	}
}

func Dispose(res *[]Result, ch chan Result, until int) {
	for i := 0; i < until; i++ {
		select {
		case r := <-ch:
			initFilters(config.filter, r.stat)
			if checkFilter() {
				*res = append(*res, r)
				videoUpdateChan <- true
			}
		}
	}
}

func runEngine(cfg *Configuration, res *[]Result) {
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
	if cfg.usefuzzer {
		max *= 10
	}
	channel := make(chan Result)
	for i := 0; i < max; i++ {
		select {
		case r := <-stopEngine:
			if r == true {
				return
			}
		default:
			if slice_wordlist[i%len(slice_wordlist)] != "" {
				rnd_encoder_num := rand.Intn(len(cfg.encoderList))
				if cfg.usefuzzer {
					fuzzed_data = cfg.encoderList[rnd_encoder_num](fuzz(slice_wordlist[i%len(slice_wordlist)]))
				} else {
					fuzzed_data = cfg.encoderList[rnd_encoder_num](slice_wordlist[i])
				}
				wg.Add(1)
				concurrency++
				if cfg.url != "" {
					tmp_url = strings.Replace(cfg.url, "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD)
					fuzzurl, _ := url.ParseRequestURI(tmp_url)
					if cfg.postdata == "" {
						req, _ = http.NewRequest("GET", fuzzurl.String(), nil)
					} else {
						var data url.Values
						reqBody = strings.Replace(cfg.postdata, "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD)
						data, _ = url.ParseQuery(cfg.postdata)
						req, _ = http.NewRequest("POST", fuzzurl.String(), strings.NewReader(data.Encode()))
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
					}
					if cfg.cookies != "" {
						req.Header.Add("Cookie", strings.Replace(cfg.cookies, "FUZZ", url.QueryEscape(fuzzed_data), MAX_FUZZ_KEYWORD))
					}
					if cfg.ssl {
						req.URL.Scheme = "https"
					} else {
						req.URL.Scheme = "http"
					}
				} else if cfg.templateData != "" {
					tmp_data := strings.Replace(cfg.templateData, "FUZZ", url.QueryEscape(fuzzed_data), 1)
					tmp_req, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(tmp_data)))
					var full_url string = ""
					if cfg.ssl {
						full_url = fmt.Sprintf("%s%s%s", "https://", tmp_req.Host, tmp_req.URL.String())
					} else {
						full_url = fmt.Sprintf("%s%s%s", "http://", tmp_req.Host, tmp_req.URL.String())
					}
					full_url_parsed, _ := url.ParseRequestURI(full_url)
					req, _ = http.NewRequest(tmp_req.Method, full_url_parsed.String(), strings.NewReader(tmp_req.Form.Encode()))
					req.Header = tmp_req.Header
				}
				if cfg.auth != "" {
					req.Header.Set("Authentication", cfg.auth)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
				req.Header.Set("Content-Encoding", "identity")
				req.Header.Set("Connection", "keep-alive")
				go Request(&wg, channel, netClient, req, fuzzed_data, videoUpdateChan, i, stopEngine, reqBody)
				percentage = (100 * i) / max
				if concurrency > max_concurrency {
					wg.Wait()
					Dispose(res, channel, concurrency)
					concurrency = 0
				}
			}
		}
	}
	percentage = (100 * max) / max
	Dispose(res, channel, concurrency)
	started = false
	stopFuzzingEngine <- true
	return
}
