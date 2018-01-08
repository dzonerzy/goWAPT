/*

Author: DZONERZY

XSS-Reflect

This plugin try to discover reflected parameters and then try to exploit them by
escaping tags/scripts and by inserting new tags, the check is performed on tag
counts.

Register the plugin inside goWAPT, the returned value is the plugin name
Parameters:
  name        (string)  Plugin name
  description (string)  Plugin small description
  scanType    (int)     The scan type (invasive, mid-invasive, not-invasive)
*/
xss_reflect = Scanner.registerPlugin("XSS-Reflect", "Try to check for reflected XSS Injection", ScanType.RISK_INVASIVE)

function plain(x) {
    return x
}

function urlencode(x){
  return encodeURI(x);
}

function urlurlencode(x){
  return urlencode(urlencode(x));
}

var attacks = [
  '"><img>',
  '";</script><img>',
  "'><img>",
  "';</script><img>",
  "*/ --></script><img>",
  "><img>",
  "</title><img>",
  "</style><img>"
]

var encoders = [
  plain,
  urlencode,
  urlurlencode,
]

var reflect_string =  "JUSTTOKNOWIFITSREFLECTED"

/*

test function is the callback called by goWAPT when the plugin is executed, it
accept an argument, the base request configured via --from-proxy (it's an *http.Request)

*/
function test(base_request){
  vulnerabilities = []
  base_response = Http.sendRequest(base_request)
  if(String(base_response.Header.Get("Content-Type")).indexOf("text/html")!==-1){
    base_tags = Html.countResponseTags(base_response)
    all_parameters = Utils.getAllParameters(base_request)
    for(var i=0; i<all_parameters.length; i++) {
      if(all_parameters[i].position == 1 || all_parameters[i].position == 3){
        Utils.setParameter(all_parameters[i], reflect_string)
        response = Http.sendRequest(base_request)
        offset = Http.findOffsetInResponse(response, reflect_string)
        Utils.setParameter(all_parameters[i], all_parameters[i].orgValue)
        if(offset[0] > -1){
          org = all_parameters[i].curValue
          attackLoop:
          for(var k=0; k<attacks.length; k++){
            for(var enc=0; enc<encoders.length; enc++) {
              Utils.setParameter(all_parameters[i], encoders[enc](attacks[k]))
              all_parameters[i].orgValue = org
              response = Http.sendRequest(base_request)
              tags = Html.countResponseTags(response)
              if(tags > base_tags){
                response = Utils.httpToString(response)
                /*
                Create a vulnerability object
                Parameters:
                  description (string)    The vulnerability description
                  request     (string)    The modified request
                  response    (string)    The response
                  pattern     (string)    The pattern that caused the test to pass
                  parameter   (Parameter) The vulnerable parameter
                  confidence  (int)       The confidence level (certain, firm, possible)
                  severity    (int)       The severity of the vulnerability
                  pluginName  (string)    The name of the plugin that found the vulnerability
                */
                vuln = Scanner.makePassedTest(
                  "A reflected XSS was found",
                  Utils.httpToString(base_request),
                  response,
                  encodeURI(encoders[enc](attacks[k])),
                  all_parameters[i],
                  Vuln.CONFIDENCE_POSSIBLE,
                  Severity.CRITICAL,
                  xss_reflect
                )
                vulnerabilities.push(vuln)
                break attackLoop
              }
            }
          }
          Utils.setParameter(all_parameters[i], org)
        }
      }
    }
  }
  /*
  Return an array of vulnerabilities
  */
  return vulnerabilities
}
