package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/robertkrimen/otto"
)

type Encoder func(string) string

type Result struct {
	payload          string
	request          string
	response         string
	request_response string
	request_number   int
	stat             *Stats
}

var config Configuration
var stats Stats
var results []Result
var percentage int
var max_concurrency int = 0

const (
	POSITION_URL    = 1
	POSITION_DATA   = 2
	POSITION_COOKIE = 3
	POSITION_HEADER = 4
)

type Configuration struct {
	url              string
	template         string
	templateData     string
	postdata         string
	ssl              bool
	from_proxy       bool
	wordlist         string
	usefuzzer        bool
	filter           string
	threads          int
	encoders         string
	encoderList      []Encoder
	cookies          string
	request          http.Request
	keyword_position int
	upstream_proxy   string
	upstream_url     *url.URL
	auth             string
	extension        string
	headers          []string
	scanner          bool
	plugin_dir       string
	base_request     *http.Request
}

type Stats struct {
	code   int
	length int
	words  int
	lines  int
	chars  int
	tags   int
}

const (
	variable = 1
	op       = 3
	data     = 6
)

const (
	DRAW_STATS   = 0
	DRAW_REQUEST = 1
	DRAW_SEARCH  = 2
	DRAW_GOTOREQ = 3
)

const (
	CALLBACK_EVERY_KEY = 1337
)

var MAX_FUZZ_KEYWORD = 5

var JSVM *otto.Otto = otto.New()

var JSScanVM *otto.Otto = otto.New()

type JSHTTPInterceptor func(interface{}, interface{}, bool)

var HTTPInterceptor JSHTTPInterceptor

var haveHTTPInterceptor bool = false

type Headers []string

func (i *Headers) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *Headers) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var headers Headers

var ScannerPlugins []ScannerPlugin

var ScannerResults []TestResult

var testIncID int = 1

type PluginEntryPoint func(interface{}) otto.Value

const (
	RISK_INVASIVE     = 2
	RISK_NOT_INVASIVE = 4
	RISK_MID_INVASIVE = 8
)

const (
	PARAM_POSITION_BODY   = 1
	PARAM_POSITION_HEADER = 2
	PARAM_POSITION_URL    = 3
)

const (
	CONFIDENCE_CERTAIN  = 32
	CONFIDENCE_POSSIBLE = 64
	CONFIDENCE_FIRM     = 128
)

const (
	CRITICAL = 10
	HIGH     = 8
	MEDIUM   = 5
	LOW      = 3
	INFO     = 1
)

var confidence_text = make(map[int]string)

var severty_text = make(map[int]string)

type ScannerPlugin struct {
	smallDescription string
	scanType         int
	name             string
	entryPoint       PluginEntryPoint
}

type Parameter map[string]interface{}

type TestResult struct {
	id               int
	testPassed       bool
	description      string
	request          string
	response         string
	request_response string
	pattern          string
	parameter        Parameter
	confidence       int
	plugin           string
	severity         int
}

var fuzz_menu_is_fuzz bool = true

var levelText = make(map[int]string)
