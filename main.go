package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

// ItemList is the collection of items
type ItemList struct {
	Heading string `json:"heading"`
	Trails  []Item `json:"trails"`
}

// Item is the basic article data model
type Item struct {
	URL        string `json:"url"`
	LinkText   string `json:"linkText"`
	ShowByline string `json:"showByline"`
	Byline     string `json:"byline"`
	Image      string `json:"image"`
	IsLiveblog string `json:"isLiveBlog"`
}

// CAPIItem is the CAPI iten model
type CAPIItem struct {
	ID string `json:"id"`
}

// CAPIResponse is the main CAPI response model
type CAPIResponse struct {
	Response struct {
		Results []CAPIItem `json:"mostViewed"`
	} `json:"response"`
}

func main() {
	c := cache.New(5*time.Minute, 10*time.Minute)

	http.HandleFunc("/most-viewed/", mostViewedHandler(c))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mostViewedHandler(c *cache.Cache) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var items CAPIResponse
		var err error

		path := strings.TrimPrefix(r.URL.Path, "/most-viewed/")

		switch path {
		case "uk", "us", "au":
			items, err = cachedGet(path, c)
		default:
			items, err = capiGet(path)
		}

		if err != nil {
			errorResponse(w, err)
			return
		}

		respJSON := items.asItemList().asJSON()
		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
		return
	}
}

func cachedGet(path string, c *cache.Cache) (CAPIResponse, error) {
	if items, found := c.Get(path); found {
		return items.(CAPIResponse), nil
	}

	// get from CAPI, set cache and return
	items, err := capiGet(path)

	if err != nil {
		return items, errors.Wrap(err, "CAPI GET failed")
	}

	c.Set(path, items, cache.DefaultExpiration)
	return items, nil
}

func capiGet(path string) (CAPIResponse, error) {
	var response CAPIResponse
	APIKey := "test"

	url := fmt.Sprintf("https://content.guardianapis.com/%s?show-most-viewed=true&api-key=%s", path, APIKey)

	resp, err := http.Get(url)
	if err != nil {
		return response, errors.Wrap(err, "GET failed")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, errors.Wrap(err, "Unable to read response body")
	}

	err = json.Unmarshal(body, &response) // TODO fixme
	if err != nil {
		return response, errors.Wrap(err, "Unable to unmarshal response body")
	}

	return response, err
}

func (resp CAPIResponse) asItemList() ItemList {
	var items []Item

	for _, capiItem := range resp.Response.Results {
		item := Item{
			URL:        capiItem.ID,
			LinkText:   "foo",
			ShowByline: "foo",
			Byline:     "foo",
			Image:      "foo",
			IsLiveblog: "foo",
		}

		items = append(items, item)
	}

	return ItemList{
		Heading: "Placeholder heading",
		Trails:  items,
	}
}

func (il ItemList) asJSON() []byte {
	respJSON, err := json.Marshal(il)
	if err != nil {
		log.Fatalf("Unable to marshal item list (should never happen), %s", err)
	}

	return respJSON
}

func errorResponse(w http.ResponseWriter, err error) {
	log.Printf("%s", err)
	w.WriteHeader(http.StatusInternalServerError)
}
