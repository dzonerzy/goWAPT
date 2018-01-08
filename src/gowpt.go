package main

import (
	"flag"

	"github.com/nsf/termbox-go"
)

func parseCli() Configuration {
	var url = flag.String("u", "", "URL to fuzz")
	var template = flag.String("t", "", "Template for request")
	var postdata = flag.String("d", "", "POST data for request")
	var ssl = flag.Bool("ssl", false, "Use SSL")
	var wordlist = flag.String("w", "", "Wordlist file")
	var usefuzzer = flag.Bool("fuzz", false, "Use the built-in fuzzer")
	var filter = flag.String("f", "", "Filter the results")
	var threads = flag.Int("threads", 10, "Number of threads")
	var encoders = flag.String("e", "plain", "A list of comma separated encoders")
	var cookies = flag.String("c", "", "A list of cookies")
	var upstream = flag.String("p", "", "Use upstream proxy")
	var auth = flag.String("a", "", "Basic authentication (user:password)")
	var extension = flag.String("x", "", "Extension file example.js")
	var viaproxy = flag.Bool("from-proxy", false, "Get the request via a proxy server")
	var scanner = flag.Bool("scanner", false, "Run in scanning mode")
	var plugin_dir = flag.String("plugin-dir", "", "Directory containing all scanning module")
	flag.Var(&headers, "H", "A list of additional headers")
	flag.Parse()
	config = Configuration{url: *url,
		template: *template, postdata: *postdata,
		ssl: *ssl, wordlist: *wordlist, usefuzzer: *usefuzzer,
		filter: *filter, threads: *threads, encoders: *encoders,
		cookies: *cookies, upstream_proxy: *upstream, auth: *auth,
		extension: *extension, headers: headers, from_proxy: *viaproxy,
		scanner: *scanner, plugin_dir: *plugin_dir}
	config = checkConfig(&config)
	return config
}

func main() {
	cfg := parseCli()
	if !cfg.scanner {
		err := termbox.Init()
		if err != nil {
			panic(err)
		}
		defer termbox.Close()
		termbox.SetInputMode(termbox.InputEsc)
		termbox.SetOutputMode(termbox.Output256)
		mainMenu(cfg)
	} else {
		initScanner(&cfg)
		err := termbox.Init()
		if err != nil {
			panic(err)
		}
		defer termbox.Close()
		termbox.SetInputMode(termbox.InputEsc)
		termbox.SetOutputMode(termbox.Output256)
		initPrints()
		started = true
		go startScanEngine(cfg.base_request, &ScannerPlugins)
		fuzz_menu_is_fuzz = false
		fuzzMenu()
	}
}
