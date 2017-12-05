package main

import (
	"net/http"
	"net/url"
)

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
