package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func badConfig() {
	os.Exit(-1)
}

func checkConfig(c *Configuration) Configuration {
	if c.url == "" && c.template == "" {
		fmt.Println("Error: At least an url or a template must be specified.")
		badConfig()
	}

	if c.url != "" && c.template != "" {
		fmt.Println("Error: You must define only one option.")
		badConfig()
	}
	if c.template != "" && c.postdata != "" {
		fmt.Println("Error: Template option should include POST data in template.")
		badConfig()
	}

	if c.filter != "" {
		r, _ := regexp.Compile("(?P<line>[a-z]+)(\\s{0,100})(?P<op>(<|>|\\|\\||==|!=))(\\s{0,100})(?P<data>[-\\d]+)")
		if !r.Match([]byte(c.filter)) {
			fmt.Println("Error: Bad filter expression.")
			badConfig()
		}
	}

	if c.upstream_proxy != "" {
		proxy_url, err := url.ParseRequestURI(c.upstream_proxy)

		if err != nil {
			fmt.Println("Error: Please use a valid upstream proxy url.")
			badConfig()
		}
		c.upstream_url = proxy_url
	}

	if c.url != "" && c.ssl == false {
		if strings.Contains(c.url, "http://") {
			c.ssl = false
		} else if strings.Contains(c.url, "https://") {
			c.ssl = true
		}
	}

	if c.url != "" && !strings.HasPrefix(c.url, "http") {
		fmt.Println("Error: Invalid protocol specified!")
		badConfig()
	}

	if c.template != "" && c.url == "" {
		if _, err := os.Stat(c.template); os.IsNotExist(err) {
			fmt.Println("Error: Template file does not exist.")
			badConfig()
		} else {
			_, err := os.Open(c.template)
			if err != nil {
				fmt.Println("Error: You don't have read permission on template file.")
				badConfig()
			}
		}
	}

	if c.url != "" {
		have_keyword := false
		have_keyword = have_keyword || strings.Contains(c.url, "FUZZ")
		have_keyword = have_keyword || strings.Contains(c.postdata, "FUZZ")
		have_keyword = have_keyword || strings.Contains(c.cookies, "FUZZ")
		for _, v := range c.headers {
			have_keyword = have_keyword || strings.Contains(v, "FUZZ")
		}
		if !have_keyword {
			fmt.Println("Error: Neither url nor post data nor cookie nor headers contains FUZZ keyword.")
			badConfig()
		}
	}

	if c.template != "" {
		data, _ := ioutil.ReadFile(c.template)
		if !strings.Contains(string(data), "FUZZ") {
			fmt.Println("Error: When using template a keyword FUZZ must be specified.")
			badConfig()
		} else {
			c.templateData = string(data) + "\r\n"
		}
	}

	if c.usefuzzer && c.wordlist == "" {
		fmt.Println("Error: Fuzzer mode require a wordlist.")
		badConfig()
	}

	if !c.usefuzzer && c.wordlist == "" {
		fmt.Println("Error: Please specify a wordlist")
		badConfig()
	}

	fmt.Println(c.wordlist)
	if c.wordlist != "" {
		if _, err := os.Stat(c.wordlist); os.IsNotExist(err) {
			fmt.Println("Error: Wordlist file does not exist.")
			badConfig()
		} else {
			_, err := os.Open(c.wordlist)
			if err != nil {
				fmt.Println("Error: You don't have read permission on wordlist file.")
				badConfig()
			}
		}
	}

	initEncoders()

	if c.extension != "" {
		if _, err := os.Stat(c.extension); os.IsNotExist(err) {
			fmt.Println("Error: Extension file does not exist.")
			badConfig()
		} else {
			_, err := os.Open(c.extension)
			if err != nil {
				fmt.Println("Error: You don't have read permission on extension file.")
				badConfig()
			}
		}
		initExtender()
		runExtension(c.extension)
	}

	for _, enc := range strings.Split(c.encoders, ",") {
		enc = strings.Trim(enc, " ")
		if !strings.Contains(enc, "@") {
			if _, ok := encoders[enc]; ok {
				c.encoderList = append(c.encoderList, encoders[enc])
			} else {
				fmt.Println(fmt.Sprintf("Error: Unable to find encoder named '%s'.", enc))
				badConfig()
			}
		} else {
			encs := strings.Split(enc, "@")
			if len(encs) == 2 {
				if _, oke1 := encoders[encs[0]]; oke1 {
					if _, oke2 := encoders[encs[1]]; oke2 {
						f := func(x string) string {
							return encoders[encs[1]](encoders[encs[0]](x))
						}
						c.encoderList = append(c.encoderList, Encoder(f))
					}
				} else {
					fmt.Println(fmt.Sprintf("Error: Unable to find encoder named '%v'.", enc))
					badConfig()
				}
			} else {
				fmt.Println("Error: Maximum encoder chaining is 2 functions")
				badConfig()
			}

		}
	}

	if c.auth != "" {
		c.auth = "Basic " + b64(c.auth)
	}

	max_concurrency = c.threads

	return *c
}
