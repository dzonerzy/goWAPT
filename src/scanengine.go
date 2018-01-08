package main

import (
	"net/http"
)

var stopScannerEngine chan bool
var stopScanEngine chan bool
var endScanDispose chan bool
var stopPlugin chan bool
var scanFinishedChan chan bool
var scanStopPlugin chan bool
var testChannel chan []TestResult
var scanStoppedViaStop bool = false
var scanEndDispose chan bool

func startScanEngine(base_request *http.Request, plugins *[]ScannerPlugin) {
	stopScannerEngine = make(chan bool)
	stopScanEngine = make(chan bool)
	go runScanEngine(base_request, plugins)
	for {
		select {
		case res := <-stopScannerEngine:
			if res == true {
				stopScanEngine <- true
				return
			}
		}
	}
}

func scanDispose(res *[]TestResult, ch chan []TestResult, until int, finish chan bool, end chan bool) {
	for i := 1; i <= until; i++ {
		select {
		case r := <-ch:
			percentage = (100 * i) / until
			for _, element := range r {
				*res = append(*res, element)
				videoUpdateChan <- true
			}
			videoUpdateChan <- true
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

func scanWaitTillEnd(finished chan bool) {
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

func doPlugin(ch chan []TestResult, stopch chan bool, base_request *http.Request, plugin ScannerPlugin) {
	var done bool = false
	for !done {
		select {
		case s := <-stopch:
			if s == true {
				return
			}
		default:
			r := plugin.entryPoint(base_request)
			val, _ := r.Export()
			switch v := val.(type) {
			case []TestResult:
				ch <- v
			default:
				ch <- []TestResult{}
			}
			done = true
		}
	}
}

func runScanEngine(base_request *http.Request, plugins *[]ScannerPlugin) {
	testChannel = make(chan []TestResult)
	scanFinishedChan = make(chan bool)
	scanStopPlugin = make(chan bool)
	scanEndDispose = make(chan bool)
	videoUpdateChan = make(chan bool)
	max := len(*plugins)
	go scanDispose(&ScannerResults, testChannel, max, scanFinishedChan, scanEndDispose)
scanEngineLoop:
	for _, plugin := range *plugins {
		select {
		case r := <-stopEngine:
			if r == true {
				scanStoppedViaStop = true
				scanEndDispose <- true
				break scanEngineLoop
			}
		default:
			r := plugin.entryPoint(base_request)
			val, _ := r.Export()
			switch v := val.(type) {
			case []TestResult:
				testChannel <- v
			default:
				testChannel <- []TestResult{}
			}
		}
	}
	scanWaitTillEnd(scanFinishedChan)
	if !scanStoppedViaStop {
		stopScannerEngine <- true
	}
	started = false
	percentage = (100 * max) / max
	videoUpdateChan <- true
	return
}
