/*

Author: DZONERZY

CORS-Policy

This plugin try to check for bad CORS policy

*/

cors = Scanner.registerPlugin("CORS-Policy", "Try to check for for bad CORS policy", ScanType.RISK_NOT_INVASIVE)


function test(base_request){
  vulnerabilities = []
  Utils.addParameter("Origin","http://www.evil.com", Param.POSITION_HEADER, base_request)
  response = Http.sendRequest(base_request)
  policy = response.Header.Get("Access-Control-Allow-Origin")
  if(policy == "*" || policy.indexOf("http://www.evil.com") !== -1) {
    vuln = Scanner.makePassedTest(
      "Cross domain request are possible due to bad CORS policy",
      Utils.httpToString(base_request),
      Utils.httpToString(response),
      "Access-Control-Allow-Origin: "+policy,
      {"name": "CORS"},
      Vuln.CONFIDENCE_CERTAIN,
      Severity.HIGH,
      cors
    )
    vulnerabilities.push(vuln)
  }
  return vulnerabilities
}
