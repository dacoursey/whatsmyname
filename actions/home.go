package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
)

type site struct {
	Name     string   `json:"name"`
	CheckURI string   `json:"check_uri"`
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

// We need global access to this object.
var sl siteList

// We need this here so the overridden unmarshal can get to it.
var input string

/////
// Handlers
/////

// HomeHandler manages loading the application root.
func HomeHandler(c buffalo.Context) error {
	//First we need to get the master JSON
	sl, err := getSiteList()

	if err != nil {
		return c.Error(500, errors.New("unable to load source material"))
	}

	//If this is uncommented it works properly.
	// names := make(map[string]string)

	// for _, n := range sl.Sites {
	// 	names[n.Name] = n.CheckURI
	// }

	c.Set("names", getSiteNames())
	c.Set("count", len(sl.Sites))

	for _, x := range sl.Sites {
		fmt.Printf(x.Name + " | ")
	}

	return c.Render(200, r.HTML("index.html"))
}

// FetchResults handles testing of all sites for the given input string.
func FetchResults(c buffalo.Context) error {
	i, _ := c.Value("input").(string)
	fmt.Printf("Request input string: " + i + "\n")

	for _, x := range sl.Sites {
		fmt.Printf(x.Name)
	}

	c.Set("names", getSiteNames())
	c.Set("count", 0)
	c.Set("now", time.Now())
	return c.Render(200, r.JavaScript("traffic.js"))
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

	fmt.Println(len(sl.Sites))
	for _, n := range sl.Sites {
		names[n.Name] = n.CheckURI
		fmt.Println(names[n.Name])
	}

	return names
}

func checkSite(s site) (present bool, err error) {
	present = false

	grabClient := http.Client{
		Timeout: time.Second * 2, // Kill it after 2 seconds.
	}

	// Build the http request for GET'ing the data.
	req, err := http.NewRequest(http.MethodGet, s.CheckURI, nil)
	if err != nil {
		return false, err
	}

	// Send the request and store the result in a http response.
	res, err := grabClient.Do(req)
	if err != nil {
		return false, err
	}

	// Read the body and put it into a byte array.
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(data), s.ExString) {
		//wat do?
	}

	return present, err
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

		// Modify the original CheckURI with our input value.
		s.CheckURI = strings.Replace(s.CheckURI, "{account}", input, -1)

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

// func TrafficCop(c buffalo.Context) error {
// 	time.Sleep(500 * time.Millisecond)

// 	p, _ := c.Value("badge").(string)

// 	switch p {
// 	case "success", "warning":
// 		c.Set("badge", p)
// 	default:
// 		c.Set("badge", "danger")
// 	}

// 	c.Set("count", 0)
// 	c.Set("now", time.Now())
// 	return c.Render(200, r.JavaScript("traffic.js"))
// }
