# GOWPT - Go Web Application Penetration Test

GOWPT is the younger brother of [wfuzz](https://github.com/xmendez/wfuzz) a swiss army knife of WAPT, it allow pentester to perform huge activity with no stress at all, just configure it and it's just a matter of clicks.

## How to install

To install `gowpt` just type:
```bash
make
sudo make install
```

## Usage

From the `-h` menu

```bash
Usage of ./gowpt:
  -a string
    	Basic authentication (user:password)
  -c string
    	A list of cookies
  -d string
    	POST data for request
  -e string
    	A list of comma separated encoders (default "plain")
  -f string
    	Filter the results
  -fuzz
    	Use the built-in fuzzer
  -p string
    	Use upstream proxy
  -ssl
    	Use SSL
  -t string
    	Template for request
  -threads int
    	Number of threads (default 10)
  -u string
    	URL to fuzz
  -w string
    	Wordlist file
```

**Examples**

Scan http://www.example.com and filter all `200 OK` requests

	gowpt -u "http://www.example.com/FUZZ" -w wordlist/general/common.txt -f "code == 200"
    
Scan http://www.example.com fuzzing `vuln` GET parameter looking for XSS (assume it had 200 tag with a legit request)

	gowpt -u "http://www.example.com/?vuln=FUZZ" -w wordlist/Injections/XSS.txt -f "tags > 200"
    
Scan http://www.example.com fuzzing `vuln` POST parameter looking for XSS (assume it had 200 tag with a legit request)

	gowpt -u "http://www.example.com/" -d "vuln=FUZZ" -w wordlist/Injections/XSS.txt -f "tags > 200"
    
Scan auth protected http://www.example.com and filter all `200 OK` requests

	gowpt -u "http://www.example.com/FUZZ" -w wordlist/general/common.txt -f "code == 200" -a "user:password"
    
## Wordlists

Wordlists comes from [wfuzz](https://github.com/xmendez/wfuzz) project! so thanks much guys!

## Look&Feel

[![asciicast](https://asciinema.org/a/151130.png)](https://asciinema.org/a/151130)

## Encoders

Below the list of encoders available

- **url** (URL encode)
- **urlurl** (Double URL encode)
- **html** (HTML encode)
- **htmlhex** (HTML hex encode)
- **unicode** (Unicode encode)
- **hex** (Hex encode)
- **md5hash** (MD5 hash)
- **sha1hash** (SHA1 hash)
- **sha2hash** (SHA2 hash)
- **b64** (Base64 encode)
- **b32** (Base32 encode)
- **plain** (No encoding)

## Filters

You can apply filters on the following variables

- **tags** (Number of tags)
- **lines** (Number of lines of response body)
- **words** (Number of words of response body)
- **length** (Size of response body)
- **code** (HTTP status code)
- **chars** (Number of chars of response body)

## License

`gowpt` is released under the GPL 3.0 license and it's copyleft of Daniele 'dzonerzy' Linguaglossa
