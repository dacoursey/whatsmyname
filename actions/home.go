package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

type siteData struct {
	License []string `json:"license"`
	Authors []string `json:"authors"`
	Sites   []site   `json:"sites"`
}

// We need this here so the overridden unmarshal can get to it.
var input string

/////
// Handlers
/////

// HomeHandler is a default handler to serve up a home page.
func HomeHandler(c buffalo.Context) error {
	//First we need to get the master JSON
	sd, err := grab()

	if err != nil {
		return c.Error(500, errors.New("unable to load source material"))
	}

	c.Set("count", len(sd.Sites))

	return c.Render(200, r.HTML("index.html"))
}

func grab() (sd siteData, err error) {
	// We will use this to track requests.
	//uuid, err := newUUID()

	// This is the master JSON file for site data.
	// We need to pull the site data from the source.
	master := "https://raw.githubusercontent.com/WebBreacher/WhatsMyName/master/web_accounts_list.json"
	grabClient := http.Client{
		Timeout: time.Second * 2, // Kill it after 2 seconds.
	}

	// Build the http request for GET'ing the data.
	req, err := http.NewRequest(http.MethodGet, master, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Send the request and store the result in a http response.
	res, getErr := grabClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	// Read the body and put it into a byte array.
	data, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	// Unmarshal the JSON into our object.
	var sList siteData
	if err := json.Unmarshal(data, &sList); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return sList, err
}

/////
// Helpers
/////

// This is our custom unmarshaler that skips anything with square brackets
func (sd *siteData) UnmarshalJSON(data []byte) error {
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
	*sd = siteData{
		License: tmp.License,
		Authors: tmp.Authors,
		Sites:   sites,
	}
	return nil
}
