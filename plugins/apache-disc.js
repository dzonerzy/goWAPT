/*

Author: DZONERZY

Apache-Disclosure

This plugin try to find non standard Apache Server Header

*/

apache = Scanner.registerPlugin("Apache-Disc", "Try to check for non standard Apache Server Header", ScanType.RISK_NOT_INVASIVE)


function test(base_request){
  vulnerabilities = []
  response = Http.sendRequest(base_request)
  server = response.Header.Get("Server")
  if(server != "") {
    if(String(server).indexOf("Apache") !== -1 && String(server).length > 10){
      vuln = Scanner.makePassedTest(
        "Apache diclose information via Server Header",
        Utils.httpToString(base_request),
        Utils.httpToString(response),
        server,
        {"name": "Server"},
        Vuln.CONFIDENCE_FIRM,
        Severity.INFO,
        apache
      )
      vulnerabilities.push(vuln)
    }
  }
  return vulnerabilities
}
