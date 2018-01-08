/*

Author: DZONERZY

CSRF-Attack

This plugin try to detect CSRF vulnerable requests

*/

csrf = Scanner.registerPlugin("CSRF-Attack", "Try to detect CSRF vulnerable requests", ScanType.RISK_MID_INVASIVE)

function requestToBody(req){
  response = Http.sendRequest(req)
  response_string = Utils.httpToString(response)
  response_body = String(response_string).substring(String(response_string).indexOf('\r\n\r\n')+4)
  return [response, response_body]
}

function test(base_request){
  vulnerable = true
  vulnerabilities = []
  response = requestToBody(base_request)
  all_parameters = Utils.getAllParameters(base_request)
  forLoop:
  for(var i=0; i<all_parameters.length; i++){
    if(all_parameters[i].name != "Cookie"){
      org_val = String(all_parameters[i].curValue)
      new_val = org_val.slice(0, -4) + "FAKE"
      Utils.setParameter(all_parameters[i], new_val)
      response_tmp = Http.sendRequest(base_request)
      response_string_tmp = Utils.httpToString(response_tmp)
      response_body_tmp = String(response_string_tmp).substring(String(response_string_tmp).indexOf('\r\n\r\n')+4)
      if(response_body_tmp.length == response[1].length) {
        vulnerable = vulnerable & true
        // What will happen if we remove the csrf token parameter ???
        Utils.setParameter(all_parameters[i], org_val)
      }else{
        vulnerable = vulnerable & false
        Utils.setParameter(all_parameters[i], org_val)
      }
    }
  }
  if(vulnerable){
    vuln = Scanner.makePassedTest(
      "Request is not protected agains CSRF Attack",
      Utils.httpToString(base_request),
      response_string_tmp,
      "Response body never changed during parameter modification",
      {"name": "<None>"},
      Vuln.CONFIDENCE_POSSIBLE,
      Severity.MEDIUM,
      csrf
    )
    vulnerabilities.push(vuln)
  }
  return vulnerabilities
}
