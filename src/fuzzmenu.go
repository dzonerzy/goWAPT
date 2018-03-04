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
var wassearching bool = false
var order = 5

type ByRequestNumber []Result

func (c ByRequestNumber) Len() int           { return len(c) }
func (c ByRequestNumber) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByRequestNumber) Less(i, j int) bool { return c[i].request_number < c[j].request_number }

type ByTagsASC []Result

func (c ByTagsASC) Len() int           { return len(c) }
func (c ByTagsASC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByTagsASC) Less(i, j int) bool { return c[i].stat.tags < c[j].stat.tags }

type ByTagsDESC []Result

func (c ByTagsDESC) Len() int           { return len(c) }
func (c ByTagsDESC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByTagsDESC) Less(i, j int) bool { return c[i].stat.tags > c[j].stat.tags }

type ByHTTPCodeASC []Result

func (c ByHTTPCodeASC) Len() int           { return len(c) }
func (c ByHTTPCodeASC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByHTTPCodeASC) Less(i, j int) bool { return c[i].stat.code < c[j].stat.code }

type ByHTTPCodeDESC []Result

func (c ByHTTPCodeDESC) Len() int           { return len(c) }
func (c ByHTTPCodeDESC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByHTTPCodeDESC) Less(i, j int) bool { return c[i].stat.code > c[j].stat.code }

type ByWordsASC []Result

func (c ByWordsASC) Len() int           { return len(c) }
func (c ByWordsASC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByWordsASC) Less(i, j int) bool { return c[i].stat.words < c[j].stat.words }

type ByWordsDESC []Result

func (c ByWordsDESC) Len() int           { return len(c) }
func (c ByWordsDESC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByWordsDESC) Less(i, j int) bool { return c[i].stat.words > c[j].stat.words }

type ByCharsASC []Result

func (c ByCharsASC) Len() int           { return len(c) }
func (c ByCharsASC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCharsASC) Less(i, j int) bool { return c[i].stat.chars < c[j].stat.chars }

type ByCharsDESC []Result

func (c ByCharsDESC) Len() int           { return len(c) }
func (c ByCharsDESC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCharsDESC) Less(i, j int) bool { return c[i].stat.chars > c[j].stat.chars }

type ByLinesASC []Result

func (c ByLinesASC) Len() int           { return len(c) }
func (c ByLinesASC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByLinesASC) Less(i, j int) bool { return c[i].stat.lines < c[j].stat.lines }

type ByLinesDESC []Result

func (c ByLinesDESC) Len() int           { return len(c) }
func (c ByLinesDESC) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByLinesDESC) Less(i, j int) bool { return c[i].stat.lines > c[j].stat.lines }

type ByID []TestResult

func (c ByID) Len() int           { return len(c) }
func (c ByID) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByID) Less(i, j int) bool { return c[i].id < c[j].id }

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
			if fuzz_menu_is_fuzz {
				drawFooter("^7[^1Q^7] Quit ^7[^1S^7] Start/Stop ^7[^1↑^7] Navigate up ^7[^1↓^7] Navigate down ^7[^1I^7] Inspect ^7[^1O^7] Order ^7[^1G^7] Goto")
			} else {
				drawFooter("^7[^1Q^7] Quit ^7[^1↑^7] Navigate up ^7[^1↓^7] Navigate down ^7[^1I^7] Inspect ^7[^1O^7] Order ^7[^1G^7] Goto")
			}
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
	if fuzz_menu_is_fuzz {
		order = (order + 1) % 11
		switch order {
		case 0:
			sort.Sort(ByRequestNumber(results))
		case 1:
			sort.Sort(ByTagsASC(results))
		case 2:
			sort.Sort(ByTagsDESC(results))
		case 3:
			sort.Sort(ByHTTPCodeASC(results))
		case 4:
			sort.Sort(ByHTTPCodeDESC(results))
		case 5:
			sort.Sort(ByWordsASC(results))
		case 6:
			sort.Sort(ByWordsDESC(results))
		case 7:
			sort.Sort(ByCharsASC(results))
		case 8:
			sort.Sort(ByCharsDESC(results))
		case 9:
			sort.Sort(ByLinesASC(results))
		case 10:
			sort.Sort(ByLinesDESC(results))
		}
	} else {
		sort.Sort(ByID(ScannerResults))
	}
	return nil
}

func fuzzStartStop(x int) interface{} {
	if !started {
		if fuzz_menu_is_fuzz {
			go startFuzzEngine(&config, &results)
		}
		cur = 0
		start = 0
		percentage = 0
		if fuzz_menu_is_fuzz {
			results = []Result{}
		} else {
			ScannerResults = []TestResult{}
		}
		started = true
	} else {
		if fuzz_menu_is_fuzz {
			stopFuzzingEngine <- true
		} else {
			stopScannerEngine <- true
		}
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
				c := fmt.Sprintf("%c", x)
				if c == "%" {
					search_term = append(search_term, "%%")
				} else {
					search_term = append(search_term, fmt.Sprintf("%c", x))
				}
			}
		} else {
			var r []string
			found := false
			term := strings.Join(search_term, "")
			if term != "" {
				if fuzz_menu_is_fuzz {
					r = splitToFill(results[cur].request_response, w)
				} else {
					r = splitToFill(ScannerResults[cur].request_response, w)
				}
			searchLoop:
				for i := r_cur + 1; i < len(r); i++ {
					if strings.Contains(strings.ToLower(r[i]), strings.ToLower(term)) {
						r_cur = i
						r_start = i / (h - 5)
						found = true
						break searchLoop
					}
				}
				if !found {
					r_cur = 0
					r_start = r_cur / (h - 5)
				}
			}
			wassearching = true
			draw_item = DRAW_REQUEST
			resetCallbacks("fuzz")
			inspectInitHotKeys()
			fuzzInitTerminal()
			if !found && term != "" {
				msg := fmt.Sprintf("Term '%s' not found, restarting from top.", term)
				drawCell((w/2)-(len(msg)/2), h-2, termbox.ColorWhite, termbox.ColorRed, msg)
			}
		}
		termbox.Flush()
	}
	return nil
}

func inspectSearchNext(x int) interface{} {
	w, h := termbox.Size()
	var r []string
	found := false
	term := strings.Join(search_term, "")
	if term != "" {
		if fuzz_menu_is_fuzz {
			r = splitToFill(results[cur].request_response, w)
		} else {
			r = splitToFill(ScannerResults[cur].request_response, w)
		}
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
			msg := fmt.Sprintf("Term '%s' not found, restarting from top.", term)
			drawCell((w/2)-(len(msg)/2), h-2, termbox.ColorWhite, termbox.ColorRed, msg)
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
				c := fmt.Sprintf("%c", x)
				if c == "%" {
					goto_req = append(goto_req, "%%")
				} else {
					goto_req = append(goto_req, fmt.Sprintf("%c", x))
				}
			}
		} else {
			goto_str := strings.Join(goto_req, "")
			n, _ := strconv.Atoi(goto_str)
			if fuzz_menu_is_fuzz {
				for pos, element := range results {
					if element.request_number == n {
						cur = pos
						start = cur / (h - 5)
					}
				}
			} else {
				for pos, element := range ScannerResults {
					if element.id == n {
						cur = pos
						start = cur / (h - 5)
					}
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
	if !wassearching {
		_, h := termbox.Size()
		r_cur = 0
		r_start = r_cur / (h - 5)
	} else {
		wassearching = !wassearching
	}
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
	if fuzz_menu_is_fuzz {
		addCallbackMenu("fuzz", int('s'), callbackMethod(fuzzStartStop))
		addCallbackMenu("fuzz", int('S'), callbackMethod(fuzzStartStop))
	}
	addCallbackMenu("fuzz", int(termbox.KeyArrowDown), callbackMethod(fuzzIncCursor))
	addCallbackMenu("fuzz", int(termbox.KeyArrowUp), callbackMethod(fuzzDecCursor))
}

func updateStats(drawpercentage bool) {
	if w, h, correctSize := checkSize(); correctSize {
		if draw_item == DRAW_STATS {
			if fuzz_menu_is_fuzz {
				drawStats(results, cur, order)
			} else {
				drawScanStats(ScannerResults, cur)
			}
			if drawpercentage {
				drawPercent(percentage)
			}
			if started {
				drawMultiColor(w-16, h-1, "^7STATUS^8: ^5Running", 0xf2)
			} else {
				drawMultiColor(w-16, h-1, "^7STATUS^8: ^1Stopped", 0xf2)
			}
			termbox.Flush()

		} else if draw_item == DRAW_REQUEST {
			if fuzz_menu_is_fuzz {
				drawRequest(results[cur].request_response, r_cur)
			} else {
				drawRequest(ScannerResults[cur].request_response, r_cur)
			}
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
			updateStats(true)
		default:
			updateStats(true)
		}
	}
}
