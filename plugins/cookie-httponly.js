/*

Author: DZONERZY

Cookie-httpOnly

This plugin try to find non-HttpOnly cookie

*/

cookie = Scanner.registerPlugin("Cookie-httpOnly", "Try to check for non httpOnly", ScanType.RISK_NOT_INVASIVE)


function test(base_request){
  vulnerabilities = []
  cookieval = Utils.getParameter(base_request,"Cookie")
  if(cookieval != null) {
    if(String(cookieval.curVal).indexOf(" httpOnly") == -1){
      vuln = Scanner.makePassedTest(
        "Cookies are not protected agains XMLHTTP",
        Utils.httpToString(base_request),
        "",
        cookieval.curValue,
        cookieval,
        Vuln.CONFIDENCE_FIRM,
        Severity.LOW,
        cookie
      )
      vulnerabilities.push(vuln)
    }
  }
  return vulnerabilities
}
