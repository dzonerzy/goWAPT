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
* result.code => HTTP Response status
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
		/*
		* Send an HTTP request in a synchronous way
		*
		* This API accept 4 parameters:
		* method => GET | POST | HEAD | PUT | PATCH | UPDATE
		* url => The url of the HTTP service
		* post_data => The content of request bodyBytes
		* headers => A javascript dictionary {headerName => headerValue}
		*
		* The response object may be null or undefined or an http.Response from golang
		*/
		var response = sendRequestSync("GET", "http://example.com/", null, {"Fake": "Header"})
	}
})
