package hmcpcm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	//	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
)

const timeout = 30

// NewSession initialize a Session struct
func NewSession(url string, user string, password string) (*Session, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &Session{client: &http.Client{Transport: tr, Jar: jar, Timeout: time.Second * timeout}, User: user, Password: password, url: url}, nil
}

// SetLog set logger the session
func (s *Session) SetLog(l *logrus.Logger) {
	s.slog = l
}

// SetDebugLog set filename for log
func (s *Session) SetDebugLog(filename string) {

	if s.dlog != nil {
		return
	}

	l, err := GetDebugLogger(filename)
	if err != nil {
		s.Errorf("ERROR on create HMC API debug file %s ", err)
	}
	s.dlog = l

}

//Release session
func (s *Session) Release() {
	//do nothing right now
	// We should do DELETE to release the session https://www.ibm.com/support/knowledgecenter/POWER8/p8ehl/apis/Logon.htm
}

// SetSamples SetCurrent Samples
func (s *Session) SetSamples(samples int) {
	s.samples = samples
}

// DoLogon performs the login to the HMC
func (s *Session) DoLogon() error {

	authurl := s.url + "/rest/api/web/Logon"

	// template for login request
	logintemplate := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
  <LogonRequest xmlns="http://www.ibm.com/xmlns/systems/power/firmware/web/mc/2012_10/" schemaVersion="V1_1_0">
    <Metadata>
      <Atom/>
    </Metadata>
    <UserID kb="CUR" kxe="false">{{.User}}</UserID>
    <Password kb="CUR" kxe="false">{{.Password}}</Password>
  </LogonRequest>`

	tmpl := template.New("logintemplate")
	tmpl.Parse(logintemplate)
	authrequest := new(bytes.Buffer)
	err := tmpl.Execute(authrequest, s)
	if err != nil {
		return fmt.Errorf("Error on template execution error:%s", err)
	}

	request, err := http.NewRequest("PUT", authurl, authrequest)
	if err != nil {
		return fmt.Errorf("error on HTTP Request: %v", err)
	}

	// set request headers
	request.Header.Set("Content-Type", "application/vnd.ibm.powervm.web+xml; type=LogonRequest")
	request.Header.Set("Accept", "application/vnd.ibm.powervm.web+xml; type=LogonResponse")
	request.Header.Set("X-Audit-Memento", "hmctest")

	s.PrintHTTPRequest(request)
	now := time.Now()
	response, err := s.client.Do(request)
	duration := time.Since(now)
	if err != nil {
		return fmt.Errorf("HMC error sending auth request: %v", err)
	}
	defer response.Body.Close()
	s.PrintHTTPResponse(response, duration)
	s.Infof("HTTPREQ PUT: %s [%s]", authurl, duration.String())
	contents, _ := ioutil.ReadAll(response.Body)
	s.PrintHTTPContent(contents)
	if response.StatusCode != 200 {
		return fmt.Errorf("HMC authentication error: %s", response.Status)
	}

	return nil
}

func (s *Session) httpGet(link string) ([]byte, error) {
	request, _ := http.NewRequest("GET", link, nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	s.PrintHTTPRequest(request)
	now := time.Now()
	response, err := s.client.Do(request)
	duration := time.Since(now)
	if err != nil {
		return nil, err
	}
	s.PrintHTTPResponse(response, duration)
	s.Infof("HTTPREQ GET %s [%s]", link, duration.String())
	defer response.Body.Close()
	contents, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, readErr
	}
	if response.StatusCode != 200 {
		s.Errorf("Error getting LPAR informations. status code: %d", response.StatusCode)
	}
	return contents, nil
}

func (s *Session) getPCMLink(link string) (string, error) {

	var feed Feed

	contents, readErr := s.httpGet(link)
	if readErr != nil {
		s.Errorf("ERROR  on read Body: %s", readErr)
		return "", readErr
	}

	s.PrintHTTPContentXML(contents)
	unmarshalErr := xml.Unmarshal(contents, &feed)

	if unmarshalErr != nil {
		s.Errorf("ERROR  on Unmarshall Body: %s", unmarshalErr)
		return "", unmarshalErr
	}

	for _, entry := range feed.Entries {
		if len(entry.Category.Term) == 0 {
			continue
		}
		if entry.Link.Type == "application/json" {
			return entry.Link.Href, nil
		}
	}
	return "", fmt.Errorf("not JSON formated links found on %s", link)
}

// GetSysPCMData get PCM for sytems
func (s *Session) GetSysPCMData(system *ManagedSystem) (PCMData, error) {
	s.Infof("Gathering data for system %s | %s", system.SystemName, system.UUID)
	sysurl := s.url + "/rest/api/pcm/ManagedSystem/" + system.UUID + "/ProcessedMetrics?NoOfSamples=" + strconv.Itoa(s.samples)
	//Get JSON Link
	link, err := s.getPCMLink(sysurl)
	if err != nil {
		return PCMData{}, err
	}
	s.Debugf("Got System LINK %#+v", link)
	return s.GetPCMData(link)
}

// GetLparPCMData get PCM for Lpar systems
func (s *Session) GetLparPCMData(system *ManagedSystem, lpar *LogicalPartition) (PCMData, error) {
	s.Infof("Gathering data for system[%s] LPAR [%s] ", lpar.PartitionName, lpar.PartitionName)
	lparurl := s.url + "/rest/api/pcm/ManagedSystem/" + system.UUID + "/LogicalPartition/" + lpar.PartitionUUID + "/ProcessedMetrics?NoOfSamples=" + strconv.Itoa(s.samples)
	//Get JSON Link
	link, err := s.getPCMLink(lparurl)
	if err != nil {
		return PCMData{}, err
	}
	s.Debugf("Got System LINK %#+v", link)
	return s.GetPCMData(link)
}

// GetPCMData retreives the PCM data in JSON format and returns them stored in an PCMData struct
func (s *Session) GetPCMData(rawurl string) (PCMData, error) {
	var data PCMData
	u, _ := url.Parse(rawurl)
	pcmurl := s.url + u.Path

	contents, err := s.httpGet(pcmurl)
	if err != nil {
		return data, err
	}

	s.PrintHTTPContentJSON(contents)

	jsonErr := json.Unmarshal(contents, &data)

	if jsonErr != nil {
		s.Errorf("ERROR on Json Unmarshall: %s", jsonErr)
	}
	return data, jsonErr

}

// GetViosInfo returns a list of the managed systems retrieved from the atom feed
func (s *Session) GetViosInfo(link string) (*VirtualIOServer, error) {

	contents, readErr := s.httpGet(link)
	if readErr != nil {
		return nil, readErr
	}

	s.PrintHTTPContentXML(contents)

	var entry ViosEntry

	newErr := xml.Unmarshal(contents, &entry)
	if newErr != nil {
		return nil, newErr
	}
	s.Debugf("LPAR ENTRY %#+v", entry)
	return &entry.Contents[0].Vios[0], nil
}

// GetLparInfo returns a list of the managed systems retrieved from the atom feed
func (s *Session) GetLparInfo(link string) (*LogicalPartition, error) {

	contents, readErr := s.httpGet(link)
	if readErr != nil {
		return nil, readErr
	}

	s.PrintHTTPContentXML(contents)

	var entry LparEntry

	newErr := xml.Unmarshal(contents, &entry)
	if newErr != nil {
		return nil, newErr
	}
	s.Debugf("LPAR ENTRY %#+v", entry)
	return &entry.Contents[0].Lpar[0], nil
}

// GetManagedSystems returns a list of the managed systems retrieved from the atom feed
func (s *Session) GetManagedSystems() (feed *Feed, err error) {
	mgdurl := s.url + "/rest/api/uom/ManagedSystem"

	contents, readErr := s.httpGet(mgdurl)
	if readErr != nil {
		return nil, readErr
	}

	s.PrintHTTPContentXML(contents)

	newErr := xml.Unmarshal(contents, &feed)

	if newErr != nil {
		return feed, newErr
	}

	for _, entry := range feed.Entries {

		entry.Contents[0].System[0].UUID = entry.ID
		entry.Contents[0].System[0].Lpars = make(map[string]*LogicalPartition)
		entry.Contents[0].System[0].Vios = make(map[string]*VirtualIOServer)
		for _, link := range entry.Contents[0].System[0].AssociatedLogicalPartitions.Links {
			lpe, err := s.GetLparInfo(link.Href)
			if err != nil {
				s.Errorf("Error on get LPAR info for link [%s] error: %s ", link, err)
			}
			s.Debugf("LPAR : %+v", lpe)
			entry.Contents[0].System[0].Lpars[lpe.PartitionUUID] = lpe
		}
		for _, link := range entry.Contents[0].System[0].AssociatedVirtualIOServers.Links {
			lpe, err := s.GetViosInfo(link.Href)
			if err != nil {
				s.Errorf("Error on get LPAR info for link [%s] error: %s ", link, err)
			}
			s.Debugf("VIOS : %+v", lpe)
			entry.Contents[0].System[0].Vios[lpe.PartitionUUID] = lpe
		}
	}

	return
}
