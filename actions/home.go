package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
)

type site struct {
	Name     string   `json:"name"`
	CheckURI string   `json:"check_uri"`
	RealURI  string   `json:"real_uri"`
	ExCode   string   `json:"account_existence_code"`
	ExString string   `json:"account_existence_string"`
	MiCode   string   `json:"account_missing_code"`
	MiString string   `json:"account_missing_string"`
	Known    []string `json:"known_accounts"`
	Cat      string   `json:"category"`
	Valid    bool     `json:"valid"`
	Comments []string `json:"comments"`
}

type siteList struct {
	License []string `json:"license"`
	Authors []string `json:"authors"`
	Sites   []site   `json:"sites"`
}

type siteResult struct {
	name string
	url  string
}

type siteResponse struct {
	ret int
	err error
}

// We need global access to these.
var sl siteList
var sitesPresent map[string]string
var sitesMissing map[string]string
var sitesUnknown map[string]string

// We need this here so the overridden unmarshal can get to it.
var input string

/////
// Handlers
/////

// HomeHandler manages loading the application root.
func HomeHandler(c buffalo.Context) error {
	//First we need to get the master JSON
	s, err := getSiteList()
	if err != nil {
		// handle err
	}

	sl = s

	if err != nil {
		return c.Error(500, errors.New("unable to load source material"))
	}

	return c.Render(200, r.HTML("index.html"))
}

// FetchResults handles testing of all sites for the given input string.
func FetchResults(c buffalo.Context) error {
	// Basic context vars for page data.
	count := len(sl.Sites)
	c.Set("count", count)

	sitesPresent = make(map[string]string)
	sitesMissing = make(map[string]string)
	sitesUnknown = make(map[string]string)

	// Grab our input string for insertion in target URL.
	checkVal, _ := c.Value("input").(string)
	fmt.Printf("Request input string: " + checkVal + "\n")

	chans := make([]chan siteResponse, count)
	for i := range chans {
		chans[i] = make(chan siteResponse)
	}

	for i, s := range sl.Sites {
		// We need to sub in our input value into the actual URL
		realURL := strings.Replace(s.CheckURI, "{account}", checkVal, -1)
		sl.Sites[i].RealURI = realURL

		// Launch our burst of requests.
		go checkSiteConcurrent(realURL, s.ExString, s.MiString, chans[i])
	}

	for i := range chans {
		checkResponse(sl.Sites[i].RealURI, chans[i])
	}

	for _, s := range sitesPresent {
		fmt.Printf("site: %s\n", s)
	}

	fmt.Printf("\n------------------------FINISHED CHECKING------------------------\n")
	fmt.Printf("Present: %d\n", len(sitesPresent))
	fmt.Printf("Missing: %d\n", len(sitesMissing))
	fmt.Printf("Unknown: %d\n", len(sitesUnknown))

	// Set these for the HTML vars.
	c.Set("present", sitesPresent)
	c.Set("missing", sitesMissing)
	c.Set("unknown", sitesUnknown)

	return c.Render(200, r.JavaScript("results.js"))
}

/////
// Helpers
/////

func getSiteList() (sd siteList, err error) {
	// We will use this to track requests.
	//uuid, err := newUUID()
	var l siteList

	// This is the master JSON file for site data.
	// We need to pull the site data from the source.
	master := "https://raw.githubusercontent.com/WebBreacher/WhatsMyName/master/web_accounts_list.json"

	grabClient := http.Client{
		Timeout: time.Second * 2, // Kill it after 2 seconds.
	}

	// Build the http request for GET'ing the data.
	req, err := http.NewRequest(http.MethodGet, master, nil)
	if err != nil {
		return l, err
	}

	// Send the request and store the result in a http response.
	res, err := grabClient.Do(req)
	if err != nil {
		return l, err
	}

	// Read the body and put it into a byte array.
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return l, err
	}

	// Unmarshal the JSON into our object.
	if err := json.Unmarshal(data, &l); err != nil {
		return l, err
	}

	return l, err
}

func getSiteNames() (names map[string]string) {
	names = make(map[string]string)

	for _, n := range sl.Sites {
		names[n.Name] = n.CheckURI
	}

	return names
}

// TODO: Not used anymore, remove when concurrent is working.
func checkSite(url string, exists string, missing string) (resp siteResponse) {
	ret := 0
	var r siteResponse

	grabClient := http.Client{
		Timeout: time.Second * 5, // Kill it after 5 seconds.
	}

	// Build the http request for GET'ing the data.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		r = siteResponse{ret, err}
		return r
	}

	// Send the request and store the result in a http response.
	res, err := grabClient.Do(req)
	if err != nil {
		r = siteResponse{ret, err}
		return r
	}

	// Read the body and put it into a byte array.
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r = siteResponse{ret, err}
		return r
	}

	if strings.Contains(string(data), exists) {
		ret = 1
	}

	if strings.Contains(string(data), missing) {
		ret = -1
	}

	r = siteResponse{ret, err}
	return r
}

func checkSiteConcurrent(url string, exists string, missing string, resp chan siteResponse) {
	ret := 0
	var r siteResponse

	grabClient := http.Client{
		Timeout: time.Second * 5, // Kill it after 5 seconds.
	}

	// Build the http request for GET'ing the data.
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	// Send the request and store the result in a http response.
	res, getErr := grabClient.Do(req)
	if getErr != nil {
		r = siteResponse{ret, getErr}
		resp <- r
		return
	}

	// Read the body and put it into a byte array.
	data, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		r = siteResponse{ret, getErr}
		resp <- r
	}

	if strings.Contains(string(data), exists) {
		ret = 1
	}

	if strings.Contains(string(data), missing) {
		ret = -1
	}

	r = siteResponse{ret, nil}
	resp <- r
}

func checkResponse(fullURL string, resp chan siteResponse) {
	r := <-resp

	// uncomment if you need to see the errors in the log.
	// if r.err != nil {
	// 	fmt.Printf("\nError loading: %s - Return value: %d\n", fullURL, r.ret)
	// 	fmt.Printf("Error message: %s\n", r.err)
	// }

	checkVal := r.ret

	u, _ := url.Parse(fullURL)
	//sr := siteResult{u.Hostname(), fullURL}

	switch checkVal {
	case 1:
		sitesPresent[u.Hostname()] = fullURL
		fmt.Printf("*")
	case -1:
		sitesMissing[u.Hostname()] = fullURL
		fmt.Printf("-")
	case 0:
		sitesUnknown[u.Hostname()] = fullURL
		fmt.Printf("?")
	default:
		// Something went horribly wrong. :(
	}
}

// This is our custom unmarshaler that skips anything with square brackets
func (sd *siteList) UnmarshalJSON(data []byte) error {
	var tmp struct {
		License  []string        `json:"license"`
		Authors  []string        `json:"authors"`
		RawSites json.RawMessage `json:"sites"`
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	var sites []site
	dec := json.NewDecoder(bytes.NewReader(tmp.RawSites))
	// Discard the initial '['.
	dec.Token()
	for dec.More() {
		var s site
		if err := dec.Decode(&s); err != nil {
			continue
		}

		sites = append(sites, s)
	}
	// Discard the final ']'.
	dec.Token()
	*sd = siteList{
		License: tmp.License,
		Authors: tmp.Authors,
		Sites:   sites,
	}
	return nil
}
