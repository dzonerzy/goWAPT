package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

type Color struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

var formatColors = map[string]Color{}
var start int
var r_start int

func initPrints() {
	formatColors["^1"] = Color{fg: termbox.ColorRed, bg: termbox.ColorDefault}
	formatColors["^2"] = Color{fg: termbox.ColorBlue, bg: termbox.ColorDefault}
	formatColors["^3"] = Color{fg: termbox.ColorMagenta, bg: termbox.ColorDefault}
	formatColors["^4"] = Color{fg: termbox.ColorCyan, bg: termbox.ColorDefault}
	formatColors["^5"] = Color{fg: termbox.ColorGreen, bg: termbox.ColorDefault}
	formatColors["^6"] = Color{fg: termbox.ColorYellow, bg: termbox.ColorDefault}
	formatColors["^7"] = Color{fg: termbox.ColorWhite, bg: termbox.ColorDefault}
	formatColors["^8"] = Color{fg: termbox.ColorBlack, bg: termbox.ColorDefault}
	formatColors["^9"] = Color{fg: termbox.ColorDefault, bg: termbox.ColorDefault}
}

func drawCell(x, y int, fg, bg termbox.Attribute, msg string) (int, int) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
	return x, y
}

func drawColor(x, y int, msg string, bg termbox.Attribute, args ...interface{}) (int, int) {
	r, _ := regexp.Compile("(\\^[0-9]+)")
	c := r.FindString(msg)
	msg = r.ReplaceAllString(msg, "")
	color := formatColors[c]
	return drawFormat(x, y, color.fg, bg, msg, args...)
}

func drawMultiColor(x, y int, msg string, bg termbox.Attribute, args ...interface{}) {
	arg_index := 0
	tokens := strings.Split(msg, "^")
	for i, token := range tokens {
		match, _ := regexp.MatchString("%[sdvfcr]+", token)
		if match {
			arg_count := strings.Count(token, "%")
			if i != 0 {
				token = "^" + token
			}
			x, _ = drawColor(x, y, token, bg, args[arg_index:arg_index+arg_count]...)
			arg_index += arg_count
		} else {
			if i != 0 {
				token = "^" + token
			}
			x, _ = drawColor(x, y, token, bg)
		}
	}
}

func drawFormat(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) (int, int) {
	s := fmt.Sprintf(format, args...)
	return drawCell(x, y, fg, bg, s)
}

func drawCenterAlign(fg, bg termbox.Attribute, format string, args ...interface{}) {
	w, h := termbox.Size()
	s := fmt.Sprintf(format, args...)
	drawFormat((w/2)-(len(s)/2), (h/2)-1, fg, bg, s)
}

func drawCenterHorizontal(y int, bg termbox.Attribute, format string, args ...interface{}) {
	r, _ := regexp.Compile("(\\^[0-9]+)")
	w, _ := termbox.Size()
	s := fmt.Sprintf(format, args...)
	msg := r.ReplaceAllString(s, "")
	drawMultiColor((w/2)-(len(msg)/2), y, s, bg)
}

func drawCenterVertical(x int, bg termbox.Attribute, format string, args ...interface{}) {
	_, h := termbox.Size()
	s := fmt.Sprintf(format, args...)
	drawMultiColor(x, (h/2)-1, s, bg)
}

func drawCenterCenterAlign(fg, bg termbox.Attribute, format string, args ...interface{}) {
	w, h := termbox.Size()
	s := fmt.Sprintf(format, args...)
	tokens := strings.Split(s, "\n")
	y_index := (len(tokens) / 2) + 1
	for _, token := range tokens {
		drawFormat((w/2)-(len(token)/2), (h/2)-y_index, fg, bg, token)
		y_index--
	}
}

func drawHeader(msg string) {
	w, h := termbox.Size()
	r := w
	drawFormat(0, h-h, termbox.ColorWhite|termbox.AttrBold, 0xf2, strings.Repeat(" ", r))
	reg, _ := regexp.Compile("(\\^[0-9]+)")
	cleaned := reg.ReplaceAllString(msg, "")
	drawMultiColor((w/2)-(len(cleaned)/2), h-h, msg, 0xf2)
}

func drawFooter(msg string) {
	w, h := termbox.Size()
	r := w
	drawFormat(0, h-1, termbox.ColorWhite|termbox.AttrBold, 0xf2, strings.Repeat(" ", r))
	drawMultiColor(0, h-1, msg, 0xf2)
}

func drawPercent(percent int) {
	w, h := termbox.Size()
	p := fmt.Sprintf("%d％", percent)
	min := 1
	max := w - 2
	nhash := ((max * percent) / 100)
	drawMultiColor(0, h-2, "^7[", termbox.ColorDefault)
	drawMultiColor(w-1, h-2, "^7]", termbox.ColorDefault)
	drawMultiColor(min, h-2, strings.Repeat("^1#", nhash), termbox.ColorDefault)
	drawMultiColor((w/2)-(len(p)/2), h-2, p, termbox.ColorDefault, 1)
}

func drawWidthHeight() {
	w, h := termbox.Size()
	drawFormat(1, 1, termbox.ColorGreen, termbox.ColorBlue, "Width: %d Height: %d", w, h)
}

func drawScanStats(r []TestResult, r_pos int) {
	w, h := termbox.Size()
	drawFormat(0, 1, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlue, strings.Repeat(" ", w))
	drawMultiColor(0, 1, "^8[^7ID^8]        ^8[^7Severity^8]        ^8[^7Plugin^8]       ^8[^7Confidence^8]^7   ^8[^7Parameter^8]^7    ^8[^7Pattern^8]^7", termbox.ColorBlue)
	if len(r) > 0 {
		if r_pos < 0 {
			cur = 0
			r_pos = 0
		}

		if r_pos > len(r)-1 {
			cur = len(r) - 1
			r_pos = len(r) - 1
		}

		r_pos = r_pos % len(r)

		if r_pos-start >= h-5 {
			start = r_pos - (h - 5)
		} else if r_pos == start-1 && start > 0 {
			start -= 1
		}

	forloop:
		for row := 2; row < h-2; row++ {
			if start+(row-2) > len(r)-1 {
				break forloop
			}
			val := r
			space0 := strings.Repeat(" ", 12-len(strconv.Itoa(val[start+(row-2)].id)))
			space1 := strings.Repeat(" ", 15-len(severty_text[val[start+(row-2)].severity]))
			space2 := strings.Repeat(" ", 18-len(val[start+(row-2)].plugin))
			space3 := strings.Repeat(" ", 15-len(confidence_text[val[start+(row-2)].confidence]))
			space4 := strings.Repeat(" ", 14-len(val[start+(row-2)].parameter["name"].(string)))
			s := fmt.Sprintf(" %d%s%s%s%s%s%s%s%s%s%s", val[start+(row-2)].id, space0, severty_text[val[start+(row-2)].severity], space1, val[start+(row-2)].plugin, space2, confidence_text[val[start+(row-2)].confidence], space3, val[start+(row-2)].parameter["name"], space4, val[start+(row-2)].pattern)
			how_many := w - len(s)
			if how_many < 0 {
				how_many = 0
			}
			if r_pos-start == row-2 {
				drawColor(0, row, "^8%s%s", termbox.ColorGreen, s, strings.Repeat(" ", how_many))
			} else {
				drawColor(0, row, "^4%s%s", termbox.ColorDefault, s, strings.Repeat(" ", how_many))
			}
		}
	}
}

func drawStats(r []Result, r_pos int, order int) {
	w, h := termbox.Size()
	drawFormat(0, 1, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlue, strings.Repeat(" ", w))
	switch order {
	case 0:
		drawMultiColor(0, 1, "^8[^3Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 1:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^3Tags↑^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 2:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^3Tags↓^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 3:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^3HTTP Code↑^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 4:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^3HTTP Code↓^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 5:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^3Words↑^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 6:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^3Words↓^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 7:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^3Chars↑^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 8:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^3Chars↓^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 9:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^3Lines↑^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	case 10:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^3Lines↓^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	default:
		drawMultiColor(0, 1, "^8[^7Req. No.^8]    ^8[^7Tags^8]         ^8[^7HTTP Code^8]^7    ^8[^7Words^8]^7        ^8[^7Chars^8]^7        ^8[^7Lines^8]^7        ^8[^7Payload^8]^7", termbox.ColorBlue)
	}
	if len(r) > 0 {
		if r_pos < 0 {
			cur = 0
			r_pos = 0
		}

		if r_pos > len(r)-1 {
			cur = len(r) - 1
			r_pos = len(r) - 1
		}

		r_pos = r_pos % len(r)

		if r_pos-start >= h-5 {
			start = r_pos - (h - 5)
		} else if r_pos == start-1 && start > 0 {
			start -= 1
		}

	forloop:
		for row := 2; row < h-2; row++ {
			if start+(row-2) > len(r)-1 {
				break forloop
			}
			val := r
			space0 := strings.Repeat(" ", 15-len(strconv.Itoa(val[start+(row-2)].request_number)))
			space1 := strings.Repeat(" ", 15-len(strconv.Itoa(val[start+(row-2)].stat.tags)))
			space2 := strings.Repeat(" ", 15-len(strconv.Itoa(val[start+(row-2)].stat.code)))
			space3 := strings.Repeat(" ", 14-len(strconv.Itoa(val[start+(row-2)].stat.words)))
			space4 := strings.Repeat(" ", 16-len(strconv.Itoa(val[start+(row-2)].stat.chars)))
			space5 := strings.Repeat(" ", 14-len(strconv.Itoa(val[start+(row-2)].stat.lines)))
			s := fmt.Sprintf(" %d%s%d%s%d%s%d%s%d%s%d%s%s", val[start+(row-2)].request_number, space0, val[start+(row-2)].stat.tags, space1, val[start+(row-2)].stat.code, space2, val[start+(row-2)].stat.words, space3, val[start+(row-2)].stat.chars, space4, val[start+(row-2)].stat.lines, space5, val[start+(row-2)].payload)
			how_many := w - len(s)
			if how_many < 0 {
				how_many = 0
			}
			if r_pos-start == row-2 {
				drawColor(0, row, "^8%s%s", termbox.ColorGreen, s, strings.Repeat(" ", how_many))
			} else {
				drawColor(0, row, "^4%s%s", termbox.ColorDefault, s, strings.Repeat(" ", how_many))
			}
		}
	}
}

func drawRequest(req_res string, r_pos int) {
	w, h := termbox.Size()

	r := splitToFill(req_res, w)

	if len(r) > 0 {
		if r_pos < 0 {
			r_cur = 0
			r_pos = 0
		}

		if r_pos > len(r)-1 {
			r_cur = len(r) - 1
			r_pos = len(r) - 1
		}

		r_pos = r_pos % len(r)

		if r_pos-r_start >= h-5 {
			r_start = r_pos - (h - 5)
		} else if r_pos == r_start-1 && r_start > 0 {
			r_start -= 1
		}

	forloop:
		for row := 2; row < h-2; row++ {
			if r_start+(row-2) > len(r)-1 {
				break forloop
			}
			s := fmt.Sprintf("%s", r[r_start+(row-2)])
			how_many := w - len(s)
			if how_many < 0 {
				how_many = 0
			}
			if r_pos-r_start == row-2 {
				drawColor(0, row, "^8%s%s", termbox.ColorGreen, s, strings.Repeat(" ", how_many))
			} else {
				drawColor(0, row, "^4%s%s", termbox.ColorDefault, s, strings.Repeat(" ", how_many))
			}
		}
	}
}

func drawSearchBox(searchterm *[]string) {
	w, h := termbox.Size()
	drawMultiColor(0, (h/2)-5, strings.Repeat(" ", w), termbox.ColorYellow)
	drawCenterHorizontal((h/2)-5, termbox.ColorYellow, "^8Search Term")
	drawMultiColor(0, (h/2)-4, " ", termbox.ColorYellow)
	drawMultiColor(1, (h/2)-4, strings.Repeat(" ", w-1), 0xf2)
	drawMultiColor(w-1, (h/2)-4, " ", termbox.ColorYellow)
	drawMultiColor(0, (h/2)-3, strings.Repeat(" ", w), termbox.ColorYellow)
	for pos, element := range *searchterm {
		drawColor(pos+1, (h/2)-4, element, 0xf2)
	}
}

func drawGoTo(goto_r *[]string) {
	w, h := termbox.Size()
	drawMultiColor(0, (h/2)-5, strings.Repeat(" ", w), termbox.ColorYellow)
	drawCenterHorizontal((h/2)-5, termbox.ColorYellow, "^8Go to request")
	drawMultiColor(0, (h/2)-4, " ", termbox.ColorYellow)
	drawMultiColor(1, (h/2)-4, strings.Repeat(" ", w-1), 0xf2)
	drawMultiColor(w-1, (h/2)-4, " ", termbox.ColorYellow)
	drawMultiColor(0, (h/2)-3, strings.Repeat(" ", w), termbox.ColorYellow)
	for pos, element := range *goto_r {
		drawColor(pos+1, (h/2)-4, element, 0xf2)
	}
}
