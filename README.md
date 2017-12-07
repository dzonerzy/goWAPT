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

```
Usage of ./gowpt:
  -H value
    	A list of additional headers
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
  -x string
    	Extension file example.js
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
 
Scan http://www.example.com adding header `Hello: world` and filter all `200 OK` requests

	gowpt -u "http://www.example.com/FUZZ" -w wordlist/general/common.txt -f "code == 200" -H "Hello: world"
    
Scan http://www.example.com using basic auth with user/pass `guest:guest`

	gowpt -u "http://www.example.com/FUZZ" -w wordlist/general/common.txt -a "guest:guest"
    
Scan http://www.example.com adding an extension

	gowpt -u "http://www.example.com/FUZZ" -w wordlist/general/common.txt -x myextension.js
    
## Extension

Extension are an easy way to extend gowpt features, a JavaScript VM is the responsable for loading and executing extension files.

### JS Api

Below a list of currently implemented API

| Method | Number of params | Type of params |
|:------------------:|:----------------:|:------------------------------------------------------------------------:|
| addCustomEncoder | 2 | Param1 -> EncoderName (string)<br>Param2 -> EncoderLogic (function) |
| Panic | 1 | Param1 -> PanicText (string) |
| dumpResponse | 2 | Param1 -> ResponseObject (http.Response)<br> Param2 -> Path (string) |
| setHTTPInterceptor | 1 | Param1 -> HTTPCallback (function) * |

**\*** **PS: When using <u>setHTTPInterceptor</u> the callback method receive 3 paramters:**

- A request/response object
- A result object
- A flag object that indicate whenever the first object is a request or a response

More info on the example extension below:

**example.js**

```js
/*
* Create a custom encoder called helloworld
*
* This encore just add the string "_helloworld" to every payload
* coming from the wordlist
*/
addCustomEncoder("helloworld", myenc);
/*
* Define the callback method for the helloworld encoder
*/
function myenc(data) {
	return data + "_helloword";
}
/*
* Create an HTTP interceptor
*
* The interceptor will hook every request / response
* is possible to modify request before send it, anyway the respose item
* it's just shadow copy of the one received from the server so no modification
* are possible
*
*
* request_response is an object which may contains both http.Request
* or http.Response , to know which on is contained check is_request flag
*
* REMEMBER! request_response is an http.* object so you must interact with
* this one just like you would do in golang!
*
* dumpResponse is a built-in function which dump full request-response to
* disk.
* result is an object filled with stats about the response it contains some fields
*
* result.tags => Number of tags in the response
* result.code => HTTP Response stauts
* result.words => Number of words in the response
* result.lines => Number of lines in the response
* result.chars => Number of chars in the response
* result.request => Full dump of the request
* result.response => Full dump of the response
* result.response => The injected payload
*
*/
setHTTPInterceptor(function(request_response, result, is_request){
	if(is_request){
		request_response.Header.Set("Hello", "world")
	}else{
		dumpResponse(request_response, "/tmp/dump.txt")
	}
})

```

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
