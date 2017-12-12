package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

func checkSize() (int, int, bool) {
	width, height := termbox.Size()
	if (width < 105) || (height < 23) {
		return width, height, false
	}
	return width, height, true
}

func cleanAndExit() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Close()
	os.Exit(0)
}

type callbackMethod func(int) interface{}

type Callback struct {
	cmd    int
	action callbackMethod
}

var callbacks = map[string][]Callback{}

func addCallbackMenu(menuName string, charcmd int, callback callbackMethod) {
	callbacks[menuName] = append(callbacks[menuName], Callback{cmd: charcmd, action: callback})
}

func haveCallbackDefined(menuName string, charcmd int) bool {
	for _, callback := range callbacks[menuName] {
		if callback.cmd == charcmd || callback.cmd == CALLBACK_EVERY_KEY {
			return true
		}
	}
	return false
}

func checkCallback(menuName string, cmd int) interface{} {
	for _, callback := range callbacks[menuName] {
		if callback.cmd == cmd || callback.cmd == CALLBACK_EVERY_KEY {
			return callback.action(cmd)
		}
	}
	return nil
}

func formatConfigOption(s string) string {
	if s == "" {
		return "<None>"
	} else {
		return s
	}
}

func resetCallbacks(menuName string) {
	callbacks[menuName] = []Callback{}
}

func splitToFill(s string, n int) []string {
	var pieces []string
	var tmp string
	var count int = 0
	for _, element := range s {
		count++
		tmp += string(element)
		if strings.Contains(tmp, "\n") {
			count = 0
			pieces = append(pieces, tmp)
			tmp = ""
		}
		if count == n {
			count = 0
			pieces = append(pieces, tmp)
			tmp = ""
		}
	}
	pieces = append(pieces, tmp)
	return pieces
}

func response2String(response *http.Response) []byte {
	var reqBodyBytes []byte
	respBodyBytes, _ := ioutil.ReadAll(response.Body)
	cl, _ := strconv.Atoi(response.Request.Header.Get("Content-Length"))
	if cl > 0 {
		reqBodyBytes, _ = ioutil.ReadAll(response.Request.Body)
	} else {
		reqBodyBytes = []byte{}
	}
	s_req, _ := httputil.DumpRequest(response.Request, true)
	s_res, _ := httputil.DumpResponse(response, false)
	var full []byte
	full = append(full, s_req...)
	full = append(full, reqBodyBytes...)
	full = append(full, []byte{0x0a, 0x0a, 0x0a}...)
	full = append(full, s_res...)
	full = append(full, respBodyBytes...)
	return full
}
