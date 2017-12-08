package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

var tmp_search_pos int = 0
var cur int = -1
var r_cur int = 0
var started = false
var stopFuzzMenu chan bool
var draw_item int = DRAW_STATS
var search_write_pos = 1
var search_term []string
var goto_req []string
var request_search_pos = -1

type ByRequestNumber []Result

func (c ByRequestNumber) Len() int           { return len(c) }
func (c ByRequestNumber) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByRequestNumber) Less(i, j int) bool { return c[i].request_number < c[j].request_number }

func fuzzHandleEvents(ev termbox.Event) interface{} {
	switch ev.Type {
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventKey:
		if haveCallbackDefined("fuzz", int(ev.Ch)|int(ev.Key)) {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			fuzzInitTerminal()
			r := checkCallback("fuzz", int(ev.Ch)|int(ev.Key))
			termbox.Flush()
			return r
		}
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		fuzzInitTerminal()
		termbox.Flush()
	case termbox.EventResize:
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		fuzzInitTerminal()
		termbox.Flush()
	default:
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		fuzzInitTerminal()
		termbox.Flush()
	}
	return nil
}

func fuzzInitTerminal() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if _, _, correctSize := checkSize(); correctSize {
		drawHeader("^7Go Web Application Penetration Test made with ^1❤  ^7by DZONERZY")
		if draw_item == DRAW_STATS {
			drawFooter("^7[^1Q^7] Quit ^7[^1S^7] Start/Stop ^7[^1↑^7] Navigate up ^7[^1↓^7] Navigate down ^7[^1I^7] Inspect ^7[^1O^7] Order ^7[^1G^7] Goto")
		} else if draw_item == DRAW_REQUEST {
			drawFooter("^7[^1Q^7] Quit ^7[^1B^7] Back ^7[^1S^7] Search ^7[^1N^7] Find Next")
		} else if draw_item == DRAW_SEARCH {
			drawFooter("^7[^1ENTER^7] Find ^7[^1B^7] Back")
		}
	} else {
		drawCenterAlign(termbox.ColorRed, termbox.ColorDefault, "Please resize screen")
	}
	termbox.Flush()
}

func fuzzQuit(int) interface{} {
	cleanAndExit()
	return nil
}

func fuzzOrder(x int) interface{} {
	sort.Sort(ByRequestNumber(results))
	return nil
}

func fuzzStartStop(x int) interface{} {
	if !started {
		go startFuzzEngine(&config, &results)
		cur = 0
		start = 0
		percentage = 0
		results = []Result{}
		started = true
	} else {
		stopFuzzingEngine <- true
		started = false
	}
	return nil
}

func fuzzIncCursor(x int) interface{} {
	cur++
	return nil
}

func fuzzDecCursor(x int) interface{} {
	cur--
	return nil
}

func fuzzBack(x int) interface{} {
	draw_item = DRAW_STATS
	resetCallbacks("fuzz")
	fuzzInitHotkeys()
	fuzzInitTerminal()
	return nil
}

func fuzzInspectRequest(x int) interface{} {
	if cur > -1 {
		draw_item = DRAW_REQUEST
		resetCallbacks("fuzz")
		inspectInitHotKeys()
		fuzzInitTerminal()
	}
	return nil
}

func fuzzInspectIncCursor(x int) interface{} {
	r_cur++
	return nil
}

func fuzzInspectDecCursor(x int) interface{} {
	r_cur--
	return nil
}

func inspectHandleKey(x int) interface{} {
	w, h := termbox.Size()
	if x == int('b') || x == int('B') {
		draw_item = DRAW_REQUEST
		resetCallbacks("fuzz")
		inspectInitHotKeys()
		fuzzInitTerminal()
	} else {
		if x == 127 {
			if len(search_term) > 0 {
				search_term = search_term[:len(search_term)-1]
			}
		} else if x != 13 {
			if len(search_term) < w-2 {
				search_term = append(search_term, fmt.Sprintf("%c", x))
			}
		} else {
			found := false
			term := strings.Join(search_term, "")
			if term != "" {
				r := splitToFill(results[cur].request_response, w)
			searchLoop:
				for i := r_cur + 1; i < len(r); i++ {
					if strings.Contains(strings.ToLower(r[i]), strings.ToLower(term)) {
						r_cur = i
						r_start = r_cur / (h - 5)
						found = true
						break searchLoop
					}

				}
				if !found {
					r_cur = 0
					r_start = r_cur / (h - 5)
				}
			}
			draw_item = DRAW_REQUEST
			resetCallbacks("fuzz")
			inspectInitHotKeys()
			fuzzInitTerminal()
			if !found && term != "" {
				drawCenterHorizontal(h-2, termbox.ColorRed, fmt.Sprintf("^7Term '%s' not found, restarting from top.", term))
			}
		}
		termbox.Flush()
	}
	return nil
}

func inspectSearchNext(x int) interface{} {
	w, h := termbox.Size()
	found := false
	term := strings.Join(search_term, "")
	if term != "" {
		r := splitToFill(results[cur].request_response, w)
	searchLoop2:
		for i := r_cur + 1; i < len(r); i++ {
			if strings.Contains(strings.ToLower(r[i]), strings.ToLower(term)) {
				r_cur = i
				r_start = r_cur / (h - 5)
				found = true
				break searchLoop2
			}
		}
		if !found {
			r_cur = 0
			r_start = r_cur / (h - 5)
			drawCenterHorizontal(h-2, termbox.ColorRed, fmt.Sprintf("^7Term '%s' not found, restarting from top.", term))
		}
	}
	return nil
}

func gotoHandleKey(x int) interface{} {
	w, h := termbox.Size()
	if x == int('b') || x == int('B') {
		draw_item = DRAW_STATS
		resetCallbacks("fuzz")
		fuzzInitHotkeys()
		fuzzInitTerminal()
	} else {
		if x == 127 {
			if len(goto_req) > 0 {
				goto_req = goto_req[:len(goto_req)-1]
			}
		} else if x != 13 {
			if len(goto_req) < w-2 {
				goto_req = append(goto_req, fmt.Sprintf("%c", x))
			}
		} else {
			goto_str := strings.Join(goto_req, "")
			n, _ := strconv.Atoi(goto_str)
			for pos, element := range results {
				if element.request_number == n {
					cur = pos
					start = cur / (h - 5)
				}
			}
			draw_item = DRAW_STATS
			resetCallbacks("fuzz")
			fuzzInitHotkeys()
			fuzzInitTerminal()
		}
		termbox.Flush()
	}
	return nil
}

func fuzzGoTo(x int) interface{} {
	draw_item = DRAW_GOTOREQ
	resetCallbacks("fuzz")
	gotoInitHotKeys()
	fuzzInitTerminal()
	return nil
}

func gotoInitHotKeys() {
	addCallbackMenu("fuzz", CALLBACK_EVERY_KEY, callbackMethod(gotoHandleKey))
}

func inspectSearch(x int) interface{} {
	draw_item = DRAW_SEARCH
	resetCallbacks("fuzz")
	searchInitHotKeys()
	fuzzInitTerminal()
	return nil
}

func searchInitHotKeys() {
	addCallbackMenu("fuzz", CALLBACK_EVERY_KEY, callbackMethod(inspectHandleKey))
}

func inspectInitHotKeys() {
	_, h := termbox.Size()
	r_cur = 0
	r_start = r_cur / (h - 5)
	addCallbackMenu("fuzz", int('q'), callbackMethod(fuzzQuit))
	addCallbackMenu("fuzz", int('Q'), callbackMethod(fuzzQuit))
	addCallbackMenu("fuzz", int('b'), callbackMethod(fuzzBack))
	addCallbackMenu("fuzz", int('B'), callbackMethod(fuzzBack))
	addCallbackMenu("fuzz", int('s'), callbackMethod(inspectSearch))
	addCallbackMenu("fuzz", int('S'), callbackMethod(inspectSearch))
	addCallbackMenu("fuzz", int('n'), callbackMethod(inspectSearchNext))
	addCallbackMenu("fuzz", int('N'), callbackMethod(inspectSearchNext))
	addCallbackMenu("fuzz", int(termbox.KeyArrowDown), callbackMethod(fuzzInspectIncCursor))
	addCallbackMenu("fuzz", int(termbox.KeyArrowUp), callbackMethod(fuzzInspectDecCursor))
}

func fuzzInitHotkeys() {
	addCallbackMenu("fuzz", int('q'), callbackMethod(fuzzQuit))
	addCallbackMenu("fuzz", int('Q'), callbackMethod(fuzzQuit))
	addCallbackMenu("fuzz", int('I'), callbackMethod(fuzzInspectRequest))
	addCallbackMenu("fuzz", int('i'), callbackMethod(fuzzInspectRequest))
	addCallbackMenu("fuzz", int('O'), callbackMethod(fuzzOrder))
	addCallbackMenu("fuzz", int('o'), callbackMethod(fuzzOrder))
	addCallbackMenu("fuzz", int('G'), callbackMethod(fuzzGoTo))
	addCallbackMenu("fuzz", int('g'), callbackMethod(fuzzGoTo))
	addCallbackMenu("fuzz", int('s'), callbackMethod(fuzzStartStop))
	addCallbackMenu("fuzz", int('S'), callbackMethod(fuzzStartStop))
	addCallbackMenu("fuzz", int(termbox.KeyArrowDown), callbackMethod(fuzzIncCursor))
	addCallbackMenu("fuzz", int(termbox.KeyArrowUp), callbackMethod(fuzzDecCursor))
}

func updateStats() {
	if w, h, correctSize := checkSize(); correctSize {
		if draw_item == DRAW_STATS {
			drawStats(results, cur)
			drawPercent(percentage)
			if started {
				drawMultiColor(w-16, h-1, "^7STATUS^8: ^5Running", 0xf2)
			} else {
				drawMultiColor(w-16, h-1, "^7STATUS^8: ^1Stopped", 0xf2)
			}
			termbox.Flush()

		} else if draw_item == DRAW_REQUEST {
			drawRequest(results[cur].request_response, r_cur)
			termbox.Flush()
		} else if draw_item == DRAW_SEARCH {
			drawSearchBox(&search_term)
			termbox.Flush()
		} else if draw_item == DRAW_GOTOREQ {
			drawGoTo(&goto_req)
			termbox.Flush()
		}
	}
}

func fuzzMenu() {
	fuzzInitHotkeys()
	fuzzInitTerminal()

	event_queue := make(chan termbox.Event)
	go func() {
		for {
			event_queue <- termbox.PollEvent()
		}
	}()

loop:
	for {
		select {
		case ev := <-event_queue:
			ret := fuzzHandleEvents(ev)
			if ret != nil {
				break loop
			}
		case <-videoUpdateChan:
			updateStats()
		default:
			updateStats()
		}
	}
}
