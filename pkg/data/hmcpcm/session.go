package hmcpcm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
)

//
// XML parsing structures
//

// Feed base struct of Atom feed
type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
}

// Entry is the atom feed section containing the links to PCM data and the Category
type Entry struct {
	XMLName xml.Name `xml:"entry"`
	ID      string   `xml:"id"`
	Link    struct {
		Href string `xml:"href,attr"`
	} `xml:"link,omitempty"`
	Contents []Content `xml:"content"`
	Category struct {
		Term string `xml:"term,attr"`
	} `xml:"category,omitempty"`
}

// Content feed struct containing all managed systems
type Content struct {
	XMLName xml.Name        `xml:"content"`
	System  []ManagedSystem `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ ManagedSystem"`
}

// ManagedSystem struct contains a managed system and his associated partitions
type ManagedSystem struct {
	XMLName                     xml.Name `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ ManagedSystem"`
	SystemName                  string
	AssociatedLogicalPartitions Partitions `xml:"http://www.ibm.com/xmlns/systems/power/firmware/uom/mc/2012_10/ AssociatedLogicalPartitions"`
}

// Partitions contains links to the partition informations
type Partitions struct {
	Links []Link `xml:"link,omitempty"`
}

// Link the link itself is stored in the attribute href
type Link struct {
	Href string `xml:"href,attr"`
}

// System struct store system Name and UUID
type System struct {
	Name string
	UUID string
}

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

// PCMLinks store a system and associated partitions links to PCM data
type PCMLinks struct {
	System     string
	Partitions []string
}

// GetSystemPCMLinks encapsulation function
func (s *Session) GetSystemPCMLinks(uuid string) (PCMLinks, error) {
	var pcmURL string
	if s.samples > 0 {
		pcmURL = fmt.Sprintf("%s/rest/api/pcm/ManagedSystem/%s/ProcessedMetrics?NoOfSamples=%d", s.url, uuid, s.samples)
	} else {
		pcmURL = s.url + "/rest/api/pcm/ManagedSystem/" + uuid + "/ProcessedMetrics"
	}
	return s.getPCMLinks(pcmURL)
}

// GetPartitionPCMLinks encapsulation function
func (s *Session) GetPartitionPCMLinks(link string) (PCMLinks, error) {
	var pcmURL string
	if s.samples > 0 {
		pcmURL = fmt.Sprintf("%s%s?NoOfSamples=%d", s.url, link, s.samples)
	} else {
		pcmURL = s.url + link
	}
	return s.getPCMLinks(pcmURL)
}

func (s *Session) getPCMLinks(link string) (PCMLinks, error) {

	var pcmlinks PCMLinks
	request, _ := http.NewRequest("GET", link, nil)

	request.Header.Set("Accept", "*/*;q=0.8")

	s.PrintHTTPRequest(request)
	now := time.Now()
	response, requestErr := s.client.Do(request)
	duration := time.Since(now)
	if requestErr != nil {
		return pcmlinks, requestErr
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		errorMessage := fmt.Sprintf("Error getting PCM informations. status code: %d", response.StatusCode)
		statusErr := errors.New(errorMessage)
		return pcmlinks, statusErr
	}
	s.PrintHTTPResponse(response, duration)
	s.Infof("HTTPREQ GET %s [%s]", link, duration.String())

	var feed Feed
	contents, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		s.Errorf("ERROR  on read Body: %s", readErr)
		return pcmlinks, readErr
	}
	s.PrintHTTPContentXML(contents)
	unmarshalErr := xml.Unmarshal(contents, &feed)

	if unmarshalErr != nil {
		s.Errorf("ERROR  on Unmarshall Body: %s", unmarshalErr)
		return pcmlinks, unmarshalErr
	}

	for _, entry := range feed.Entries {
		if len(entry.Category.Term) == 0 {
			continue
		}
		if entry.Category.Term == "ManagedSystem" {
			pcmlinks.System = entry.Link.Href
		}

		if entry.Category.Term == "LogicalPartition" {
			pcmlinks.Partitions = append(pcmlinks.Partitions, entry.Link.Href)
		}
	}

	return pcmlinks, nil
}

// GetPCMData encapsulation function
func (s *Session) GetPCMData(link string) (PCMData, error) {
	return s.getPCMData(link)
}

// get PCMData retreives the PCM data in JSON format and returns them stored in an PCMData struct
func (s *Session) getPCMData(rawurl string) (PCMData, error) {
	var data PCMData
	u, _ := url.Parse(rawurl)
	pcmurl := s.url + u.Path

	request, _ := http.NewRequest("GET", pcmurl, nil)

	s.PrintHTTPRequest(request)
	now := time.Now()
	response, err := s.client.Do(request)
	duration := time.Since(now)
	if err != nil {
		return data, err
	}
	defer response.Body.Close()
	s.PrintHTTPResponse(response, duration)
	s.Infof("HTTPREQ GET %s [%s]", pcmurl, duration.String())
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return data, err
	}

	s.PrintHTTPContentJSON(contents)

	if response.StatusCode != 200 {
		s.Errorf("Error getting PCM Data informations. status code: %d", response.StatusCode)
	}

	jsonErr := json.Unmarshal(contents, &data)

	if jsonErr != nil {
		s.Errorf("ERROR on Json Unmarshall: %s", jsonErr)
	}
	return data, jsonErr

}

// GetManagedSystems encapsulation function
func (s *Session) GetManagedSystems() ([]System, error) {
	return s.getManagedSystems()
}

// getManagedSystems returns a list of the managed systems retrieved from the atom feed
func (s *Session) getManagedSystems() (systems []System, err error) {
	mgdurl := s.url + "/rest/api/uom/ManagedSystem"
	request, _ := http.NewRequest("GET", mgdurl, nil)

	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	s.PrintHTTPRequest(request)
	now := time.Now()
	response, err := s.client.Do(request)
	duration := time.Since(now)
	if err != nil {
		return nil, err
	}
	s.PrintHTTPResponse(response, duration)
	s.Infof("HTTPREQ GET %s [%s]", mgdurl, duration.String())
	defer response.Body.Close()
	contents, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return systems, readErr
	}

	if response.StatusCode != 200 {
		s.Errorf("Error getting LPAR informations. status code: %d", response.StatusCode)
	}

	s.PrintHTTPContentXML(contents)

	var feed Feed
	newErr := xml.Unmarshal(contents, &feed)

	if newErr != nil {
		return systems, newErr
	}
	for _, entry := range feed.Entries {

		for _, content := range entry.Contents {
			for _, system := range content.System {
				systems = append(systems, System{Name: system.SystemName, UUID: entry.ID})
			}
		}
	}

	return
}
