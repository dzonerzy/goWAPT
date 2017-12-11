package main

import (
	"github.com/nsf/termbox-go"
)

func handleEvents(ev termbox.Event, cfg Configuration) {
	switch ev.Type {
	case termbox.EventError:
		panic(ev.Err)
	case termbox.EventKey:
		if haveCallbackDefined("main", int(ev.Ch)) {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			checkCallback("main", int(ev.Ch))
			initTerminal(cfg)
			termbox.Flush()
		}
	case termbox.EventResize:
		initTerminal(cfg)
	default:
		initTerminal(cfg)
	}
}

func initTerminal(cfg Configuration) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if _, _, correctSize := checkSize(); correctSize {
		drawCenterHorizontal(2, termbox.ColorDefault, "^1CONFIGURATION")

		drawMultiColor(1, 5, "^4URL ^6: ^7%s", termbox.ColorDefault, formatConfigOption(cfg.url))
		drawMultiColor(1, 6, "^4SSL ^6: ^7%v", termbox.ColorDefault, cfg.ssl)
		drawMultiColor(1, 7, "^4Cookies ^6: ^7%v", termbox.ColorDefault, formatConfigOption(cfg.cookies))
		drawMultiColor(1, 8, "^4Encoders ^6: ^7%v", termbox.ColorDefault, cfg.encoders)
		drawMultiColor(1, 9, "^4Post Data ^6: ^7%v", termbox.ColorDefault, formatConfigOption(cfg.postdata))
		drawMultiColor(1, 10, "^4Template ^6: ^7%v", termbox.ColorDefault, formatConfigOption(cfg.template))
		drawMultiColor(1, 11, "^4Wordlist ^6: ^7%v", termbox.ColorDefault, formatConfigOption(cfg.wordlist))
		drawMultiColor(1, 12, "^4Use fuzzer ^6: ^7%v", termbox.ColorDefault, cfg.usefuzzer)
		drawMultiColor(1, 13, "^4Threads ^6: ^7%v", termbox.ColorDefault, cfg.threads)
		drawMultiColor(1, 14, "^4Filter ^6: ^7%v", termbox.ColorDefault, formatConfigOption(cfg.filter))
		drawHeader("^7Go Web Application Penetration Test made with ^1❤  ^7by DZONERZY")
		drawFooter("^7[^1Q^7] Quit ^7[^1S^7] Start Attack ^7[^1I^7] Info")
	} else {
		drawCenterAlign(termbox.ColorRed, termbox.ColorDefault, "Please resize screen")
	}
	termbox.Flush()
}

func quit(int) interface{} {
	cleanAndExit()
	return nil
}

func info(int) interface{} {
	infoMenu()
	return nil
}

func startFuzz(int) interface{} {
	started = true
	go startFuzzEngine(&config, &results)
	fuzzMenu()
	return nil
}

func initHotkeys() {
	addCallbackMenu("main", int('q'), callbackMethod(quit))
	addCallbackMenu("main", int('Q'), callbackMethod(quit))
	addCallbackMenu("main", int('i'), callbackMethod(info))
	addCallbackMenu("main", int('I'), callbackMethod(info))
	addCallbackMenu("main", int('S'), callbackMethod(startFuzz))
	addCallbackMenu("main", int('s'), callbackMethod(startFuzz))
}

func mainMenu(cfg Configuration) {
	initHotkeys()
	initPrints()
	initTerminal(cfg)
	for {
		event := termbox.PollEvent()
		handleEvents(event, cfg)
	}
}
