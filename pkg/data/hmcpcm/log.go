package hmcpcm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
)

// Debugf info
func (s *Session) Debugf(expr string, vars ...interface{}) {
	expr2 := "HMC HTTP Session [" + s.url + "] " + expr
	s.slog.Debugf(expr2, vars...)
}

// Infof info
func (s *Session) Infof(expr string, vars ...interface{}) {
	expr2 := "HMC HTTP Session [" + s.url + "] " + expr
	s.slog.Infof(expr2, vars...)
}

// Errorf info
func (s *Session) Errorf(expr string, vars ...interface{}) {
	expr2 := "HMC HTTP Session [" + s.url + "] " + expr
	s.slog.Errorf(expr2, vars...)
}

// Warnf log warn data
func (s *Session) Warnf(expr string, vars ...interface{}) {
	expr2 := "HMC HTTP Session [" + s.url + "] " + expr
	s.slog.Warnf(expr2, vars...)
}

// PrintHTTPContent  warn data
func (s *Session) PrintHTTPContent(contents []byte) {
	if s.Debug {
		s.dlog.Print("= HTTPCONT ==========================================================")
		s.dlog.Print(string(contents[:]))
	}
}

// PrintHTTPContentXML print format XML
func (s *Session) PrintHTTPContentXML(contents []byte) {
	s.PrintHTTPContent(contents)
}

// PrintHTTPContentJSON print format json
func (s *Session) PrintHTTPContentJSON(contents []byte) {
	if s.Debug {
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, contents, "", "\t")
		if err != nil {
			s.slog.Println("JSON parse error: ", err)
		}
		s.PrintHTTPContent(prettyJSON.Bytes())
	}
}

// PrintHTTPRequest print pretty  HTTP request
func (s *Session) PrintHTTPRequest(request *http.Request) {
	if s.Debug {
		requestDump, err := httputil.DumpRequest(request, true)
		if err != nil {
			s.dlog.Printf("Error on dump %s", err)
		}
		s.dlog.Print("= HTTPREQ ============================================================")
		s.dlog.Print(string(requestDump))
	}
}

// PrintHTTPResponse print pretty  HTTP response
func (s *Session) PrintHTTPResponse(response *http.Response) {
	if s.Debug {
		responseDump, err := httputil.DumpResponse(response, true)
		if err != nil {
			s.dlog.Printf("Error on dump response %s", err)
		}
		s.dlog.Print("= HTTPRESP ============================================================")
		s.dlog.Print(string(responseDump))
	}
}
