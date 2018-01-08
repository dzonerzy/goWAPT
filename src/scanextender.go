package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
	"golang.org/x/net/html"
)

func registerPlugin(name string, smalldescription string, scantype int) otto.Value {
	var plugin ScannerPlugin
	plugin.name = name
	plugin.smallDescription = smalldescription
	plugin.scanType = scantype
	ep, err := JSScanVM.Get("test")
	if err != nil {
		panic("Plugin must define a test method")
	}
	plugin_callback := func(x interface{}) otto.Value {
		res, err := ep.Call(otto.NullValue(), x)
		if err != nil {
			panic(fmt.Sprintf("[-] Erorr running plugin '%s' : %v\n", name, err))
		} else {
			v, _ := JSScanVM.ToValue(res)
			return v
		}
	}
	plugin.entryPoint = PluginEntryPoint(plugin_callback)
	ScannerPlugins = append(ScannerPlugins, plugin)
	v, _ := JSScanVM.ToValue(name)
	return v
}

func makePassedTest(description string, request string, response string, pattern string, param Parameter, confidence int, severity int, plugin_name string) otto.Value {
	test := TestResult{id: testIncID, description: description, request: request,
		response: response, pattern: pattern, parameter: param,
		confidence: confidence, severity: severity, plugin: plugin_name}
	test.request_response = request + "\n\n" + response
	testIncID++
	v, _ := JSScanVM.ToValue(test)
	return v
	//ScannerResults = append(ScannerResults, test)
}

func httpToString(obj interface{}) string {
	switch v := obj.(type) {
	case *http.Request:
		req, _ := httputil.DumpRequest(v, false)
		body, _ := ioutil.ReadAll(v.Body)
		rdrq1 := ioutil.NopCloser(bytes.NewBuffer(body))
		req = append(req, body...)
		v.Body = rdrq1
		return string(req)
	case *http.Response:
		req, _ := httputil.DumpResponse(v, false)
		body, _ := ioutil.ReadAll(v.Body)
		rdrq1 := ioutil.NopCloser(bytes.NewBuffer(body))
		req = append(req, body...)
		v.Body = rdrq1
		return string(req)
	default:
		return ""
	}
}

func getParameter(r *http.Request, name string) otto.Value {
	p := make(Parameter)
	v := r.URL.Query().Get(name)
	if v == "" {
		v := r.Header.Get(name)
		if v == "" {
			reqb, _ := ioutil.ReadAll(r.Body)
			rdrq1 := ioutil.NopCloser(bytes.NewBuffer(reqb))
			q, _ := url.ParseQuery(string(reqb))
			v := q.Get(name)
			r.Body = rdrq1
			if v == "" {
				return otto.NullValue()
			} else {
				p["request"] = r
				p["name"] = name
				p["orgValue"] = ""
				p["curValue"] = v
				p["position"] = PARAM_POSITION_BODY
				val, _ := JSScanVM.ToValue(p)
				return val
			}
		} else {
			p["request"] = r
			p["name"] = name
			p["orgValue"] = ""
			p["curValue"] = v
			p["position"] = PARAM_POSITION_HEADER
			val, _ := JSScanVM.ToValue(p)
			return val
		}
	} else {
		p["request"] = r
		p["name"] = name
		p["orgValue"] = ""
		p["curValue"] = v
		p["position"] = PARAM_POSITION_URL
		val, _ := JSScanVM.ToValue(p)
		return val
	}
}

func deleteParameter(param otto.Value) otto.Value {
	pi, _ := param.Export()
	parameter := pi.(Parameter)
	if parameter["position"] == PARAM_POSITION_URL {
		q := parameter["request"].(*http.Request).URL.Query()
		q.Del(parameter["name"].(string))
		parameter["request"].(*http.Request).URL.RawQuery = q.Encode()
	} else if parameter["position"] == PARAM_POSITION_HEADER {
		parameter["request"].(*http.Request).Header.Del(parameter["name"].(string))
	} else {
		reqb, _ := ioutil.ReadAll(parameter["request"].(*http.Request).Body)
		q, _ := url.ParseQuery(string(reqb))
		q.Del(parameter["name"].(string))
		rdrq1 := ioutil.NopCloser(bytes.NewBuffer([]byte(q.Encode())))
		parameter["request"].(*http.Request).Body = rdrq1
	}
	old_param := getParameter(parameter["request"].(*http.Request), parameter["name"].(string))
	if old_param == otto.NullValue() {
		return otto.TrueValue()
	} else {
		return otto.FalseValue()
	}
}

func setParameter(param otto.Value, value string) otto.Value {
	pi, _ := param.Export()
	parameter := pi.(Parameter)
	parameter["orgValue"] = parameter["curValue"]
	if parameter["position"] == PARAM_POSITION_URL {
		q := parameter["request"].(*http.Request).URL.Query()
		q.Set(parameter["name"].(string), value)
		parameter["request"].(*http.Request).URL.RawQuery = q.Encode()
	} else if parameter["position"] == PARAM_POSITION_HEADER {
		parameter["request"].(*http.Request).Header.Set(parameter["name"].(string), value)
	} else {
		reqb, _ := ioutil.ReadAll(parameter["request"].(*http.Request).Body)
		q, _ := url.ParseQuery(string(reqb))
		q.Set(parameter["name"].(string), value)
		rdrq1 := ioutil.NopCloser(bytes.NewBuffer([]byte(q.Encode())))
		parameter["request"].(*http.Request).Body = rdrq1
	}
	new_param := getParameter(parameter["request"].(*http.Request), parameter["name"].(string))
	new_param_i, _ := new_param.Export()
	new_parameter := new_param_i.(Parameter)
	if new_parameter["curValue"].(string) == value {
		parameter["curValue"] = value
		return otto.TrueValue()
	} else {
		return otto.FalseValue()
	}
}

func getAllParameters(r *http.Request) otto.Value {
	var params []Parameter
	for key, values := range r.URL.Query() {
		for _, value := range values {
			p := make(Parameter)
			p["request"] = r
			p["name"] = key
			p["orgValue"] = ""
			p["curValue"] = value
			p["position"] = PARAM_POSITION_URL
			params = append(params, p)
		}
	}
	for key, values := range r.Header {
		for _, value := range values {
			p := make(Parameter)
			p["request"] = r
			p["name"] = key
			p["orgValue"] = ""
			p["curValue"] = value
			p["position"] = PARAM_POSITION_HEADER
			params = append(params, p)
		}
	}
	reqb, _ := ioutil.ReadAll(r.Body)
	rdrq1 := ioutil.NopCloser(bytes.NewBuffer(reqb))
	q, _ := url.ParseQuery(string(reqb))
	r.Body = rdrq1
	for key, values := range q {
		for _, value := range values {
			p := make(Parameter)
			p["request"] = r
			p["name"] = key
			p["orgValue"] = ""
			p["curValue"] = value
			p["position"] = PARAM_POSITION_BODY
			params = append(params, p)
		}
	}
	val, _ := JSScanVM.ToValue(params)
	return val
}

func findOffsetInResponse(r *http.Response, value string) otto.Value {
	s := httpToString(r)
	offset_start := strings.Index(s, value)
	offset_end := offset_start + len(value)
	var offsets []int
	offsets = append(offsets, offset_start)
	offsets = append(offsets, offset_end)
	v, _ := JSScanVM.ToValue(offsets)
	return v
}

func sendRequest(r *http.Request) otto.Value {
	r.Header.Set("Content-Encoding", "identity")
	var done = false
	var err error
	var response *http.Response
	for !done {
		response, err = netClient.Do(r)
		if err == nil {
			var reader io.ReadCloser
			switch response.Header.Get("Content-Encoding") {
			case "gzip":
				reader, _ = gzip.NewReader(response.Body)
				response.Header.Set("Content-Encoding", "identity")
				defer reader.Close()
			default:
				reader = response.Body
			}
			bodyBytes, _ := ioutil.ReadAll(reader)
			rdr1 := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			response.Body = rdr1
			done = true
		}
	}
	v, _ := JSScanVM.ToValue(response)
	return v
}

func requestFromString(r string, ssl bool) otto.Value {
	tmp_req, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(r)))
	tmp_req.URL.Scheme = ""
	tmp_req.URL.Host = ""
	var full_url string = ""
	if ssl {
		full_url = fmt.Sprintf("%s%s%s", "https://", tmp_req.Host, tmp_req.URL.String())
	} else {
		full_url = fmt.Sprintf("%s%s%s", "http://", tmp_req.Host, tmp_req.URL.String())
	}
	full_url_parsed, _ := url.ParseRequestURI(full_url)
	body, _ := ioutil.ReadAll(tmp_req.Body)
	reqBody := string(body)
	req, _ := http.NewRequest(tmp_req.Method, full_url_parsed.String(), strings.NewReader(reqBody))
	req.Header = tmp_req.Header
	v, _ := JSScanVM.ToValue(req)
	return v
}

func Match(data string, regex string) otto.Value {
	if matched, _ := regexp.MatchString(regex, data); matched {
		return otto.TrueValue()
	} else {
		return otto.FalseValue()
	}
}

func matchGroup(data string, regex string) otto.Value {
	re, err := regexp.Compile(regex)
	if err != nil {
		return otto.NullValue()
	}
	res := re.FindAllStringSubmatch(data, -1)
	v, _ := JSScanVM.ToValue(res)
	return v
}

func countResponseTags(r *http.Response) otto.Value {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
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
	r.Body = rdr2
	v, _ := JSScanVM.ToValue(tags)
	return v
}

func addParameter(name string, value string, position int, r *http.Request) otto.Value {
	if position == PARAM_POSITION_URL {
		values := r.URL.Query()
		values.Add(name, value)
		r.URL.RawQuery = values.Encode()
	} else if position == PARAM_POSITION_HEADER {
		r.Header.Add(name, value)
	} else {
		reqb, _ := ioutil.ReadAll(r.Body)
		q, _ := url.ParseQuery(string(reqb))
		q.Set(name, value)
		rdrq1 := ioutil.NopCloser(bytes.NewBuffer([]byte(q.Encode())))
		r.Body = rdrq1
	}
	p := getParameter(r, name)
	pv, _ := p.Export()
	param := pv.(Parameter)
	if param["name"].(string) == value {
		return otto.TrueValue()
	} else {
		return otto.FalseValue()
	}
}

func initScanEngine() {
	if config.upstream_proxy != "" {
		netTransport.Proxy = http.ProxyURL(config.upstream_url)
	}
	scanner := make(map[string]interface{})
	scanner["registerPlugin"] = registerPlugin
	scanner["makePassedTest"] = makePassedTest
	scantype := make(map[string]int)
	scantype["RISK_INVASIVE"] = RISK_INVASIVE
	scantype["RISK_NOT_INVASIVE"] = RISK_NOT_INVASIVE
	scantype["RISK_MID_INVASIVE"] = RISK_MID_INVASIVE
	utils := make(map[string]interface{})
	utils["httpToString"] = httpToString
	utils["setParameter"] = setParameter
	utils["getParameter"] = getParameter
	utils["getAllParameters"] = getAllParameters
	utils["deleteParameter"] = deleteParameter
	utils["addParameter"] = addParameter
	httputils := make(map[string]interface{})
	httputils["sendRequest"] = sendRequest
	httputils["requestFromString"] = requestFromString
	httputils["findOffsetInResponse"] = findOffsetInResponse
	regex := make(map[string]interface{})
	regex["Match"] = Match
	regex["matchGroup"] = matchGroup
	html := make(map[string]interface{})
	html["countResponseTags"] = countResponseTags
	confidentiality := make(map[string]int)
	confidentiality["CONFIDENCE_CERTAIN"] = CONFIDENCE_CERTAIN
	confidentiality["CONFIDENCE_POSSIBLE"] = CONFIDENCE_POSSIBLE
	confidentiality["CONFIDENCE_FIRM"] = CONFIDENCE_FIRM
	severity := make(map[string]int)
	severity["CRITICAL"] = CRITICAL
	severity["HIGH"] = HIGH
	severity["MEDIUM"] = MEDIUM
	severity["LOW"] = LOW
	severity["INFO"] = INFO
	param := make(map[string]int)
	param["POSITION_URL"] = PARAM_POSITION_URL
	param["POSITION_BODY"] = PARAM_POSITION_BODY
	param["POSITION_HEADER"] = PARAM_POSITION_HEADER
	JSScanVM.Set("ScanType", scantype)
	JSScanVM.Set("Scanner", scanner)
	JSScanVM.Set("Utils", utils)
	JSScanVM.Set("Http", httputils)
	JSScanVM.Set("Regex", regex)
	JSScanVM.Set("Html", html)
	JSScanVM.Set("Vuln", confidentiality)
	JSScanVM.Set("Severity", severity)
	JSScanVM.Set("Param", param)
}

func loadPluginsDirectory(dir string) []string {
	files, err := filepath.Glob(dir + string(os.PathSeparator) + "*.js")
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func runScannerPlugin(filepath string) {
	script, _ := ioutil.ReadFile(filepath)
	_, err := JSScanVM.Run(string(script))
	if err != nil {
		fmt.Printf("[-] Unable to run '%s' JS Error: %v\n", filepath, err)
	}
}

func initScanner(cfg *Configuration) {
	var req *http.Request
	if cfg.templateData != "" {
		tmp_req, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(cfg.templateData)))
		tmp_req.URL.Scheme = ""
		tmp_req.URL.Host = ""
		var full_url string = ""
		if cfg.ssl {
			full_url = fmt.Sprintf("%s%s%s", "https://", tmp_req.Host, tmp_req.URL.String())
		} else {
			full_url = fmt.Sprintf("%s%s%s", "http://", tmp_req.Host, tmp_req.URL.String())
		}
		full_url_parsed, _ := url.ParseRequestURI(full_url)
		req, _ = http.NewRequest(tmp_req.Method, full_url_parsed.String(), nil)
		req.Header = tmp_req.Header
		req.Body = tmp_req.Body
	}
	if cfg.auth != "" {
		req.Header.Set("Authentication", cfg.auth)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")
	req.Header.Set("Content-Encoding", "identity")
	req.Header.Set("Connection", "keep-alive")
	cfg.base_request = req
	levelText := make(map[int]string)
	levelText[2] = "Invasive"
	levelText[4] = "Not Invasive"
	levelText[8] = "Mid Invasive"
	confidence_text[CONFIDENCE_CERTAIN] = "Certain"
	confidence_text[CONFIDENCE_FIRM] = "Firm"
	confidence_text[CONFIDENCE_POSSIBLE] = "Possible"
	severty_text[CRITICAL] = "Critical"
	severty_text[HIGH] = "High"
	severty_text[MEDIUM] = "Medium"
	severty_text[LOW] = "Low"
	severty_text[INFO] = "Info"
	initScanEngine()
	var plugins []string = loadPluginsDirectory(cfg.plugin_dir)
	for _, plugin_path := range plugins {
		runScannerPlugin(plugin_path)
	}
	for _, plugin := range ScannerPlugins {
		fmt.Printf("[+] Loaded %s plugin %s !\n", levelText[plugin.scanType], plugin.name)
	}
	fmt.Printf("[*] Loaded %d plugins.\n", len(ScannerPlugins))
	if len(ScannerPlugins) > 0 {
		fmt.Println("[!] Ready to go, press the Enter Key to continue!")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	} else {
		fmt.Println("[-] Not enough plugin to scan the request.")
		badConfig()
	}

}
