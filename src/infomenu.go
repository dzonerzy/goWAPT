package main

import (
	"github.com/nsf/termbox-go"
)

func infoHandleEvents(ev termbox.Event) interface{} {
	switch ev.Type {
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventKey:
		if haveCallbackDefined("info", int(ev.Ch)) {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			infoInitTerminal()
			return checkCallback("info", int(ev.Ch))
			//termbox.Flush()
		}
	case termbox.EventResize:
		infoInitTerminal()
	default:
		infoInitTerminal()
	}
	return nil
}

func infoInitTerminal() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if _, _, correctSize := checkSize(); correctSize {
		drawHeader("^7Go Web Application Penetration Test made with ^1❤  ^7by DZONERZY")
		drawCenterCenterAlign(termbox.ColorWhite, termbox.ColorDefault, "Go Web Application Pentration Test\n"+
			"made by DZONERZY\n"+
			"GOWAPT is an active WebApp fuzzer\n"+
			"it can be used to check for common/uncommon vulnerabilities\n"+
			"GOWAPT it's more then a scanner, it may help you secure you application\n"+
			"finding and exploitig web application vulnerabilities\n"+
			"GOWAPT is written in Go (Golang) so it's extremely fast and relaiable\n\n"+
			"For info and bug send mail to danielelinguaglossa@gmail.com")
		drawFooter("^7[^1Q^7] Quit ^7[^1B^7] Back")
	} else {
		drawCenterAlign(termbox.ColorRed, termbox.ColorDefault, "Please resize screen")
	}
	termbox.Flush()
}

func infoQuit(int) interface{} {
	cleanAndExit()
	return nil
}

func infoBack(int) interface{} {
	return true
}

func infoInitHotkeys() {
	addCallbackMenu("info", int('q'), callbackMethod(infoQuit))
	addCallbackMenu("info", int('Q'), callbackMethod(infoQuit))
	addCallbackMenu("info", int('b'), callbackMethod(infoBack))
	addCallbackMenu("info", int('B'), callbackMethod(infoBack))
}

func infoMenu() {
	infoInitHotkeys()
	infoInitTerminal()
loop:
	for {
		event := termbox.PollEvent()
		ret := infoHandleEvents(event)
		if ret != nil {
			break loop
		}
	}
}
