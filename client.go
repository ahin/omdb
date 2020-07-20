package omdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// DefaultURL is the default request URL for API requests.
	DefaultURL = "http://www.omdbapi.com/"
)

//Client is a omdb client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

//NewClient creates a new omdb Client.
func NewClient(key string, client *http.Client) *Client {
	return &Client{
		apiKey:     key,
		httpClient: client,
	}
}

//requestOmdbAPI will call the OMDB API
func (c *Client) requestOmdbAPI(params url.Values) (*http.Response, error) {

	if c.httpClient == nil {
		return nil, errors.New("http.Client is not provided")
	}
	if c.apiKey == "" {
		return nil, errors.New("Missing OMDB API Key")
	}
	params.Set("apikey", c.apiKey)

	url, err := url.Parse(DefaultURL)
	if err != nil {
		return nil, err
	}

	url.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		return nil, fmt.Errorf("http Status = %d", res.StatusCode)
	}

	return res, nil
}

//SearchByImdbID performs an API search for a specified movie or series or episode by
//the specific imdb id. Although OMDB API allows passing other parameters like Year, SearchType etc
//but they are ignored here as search is done on a unique id.
func (c *Client) SearchByImdbID(q QueryData) (interface{}, error) {

	if q.ImdbID == "" {
		return nil, errors.New("Missing ImdbID in query")
	}

	params := url.Values{}
	params.Add("i", q.ImdbID)

	res, err := c.requestOmdbAPI(params)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	envelope := &resultEnvelope{}
	err = json.Unmarshal(data, envelope)
	if err != nil {
		return nil, err
	}

	if envelope.Response == "False" {
		return nil, errors.New("Error from OMDB API: " + envelope.Error)
	}

	var val interface{}

	switch envelope.Type {

	case "movie":
		movie := MovieResult{}
		err = json.Unmarshal(data, &movie)
		if err != nil {
			return nil, err
		}
		val = movie

	case "series":
		series := SeriesResult{}
		err = json.Unmarshal(data, &series)
		if err != nil {
			return nil, err
		}
		val = series

	case "episode":
		episode := EpisodeResult{}
		err = json.Unmarshal(data, &episode)
		if err != nil {
			return nil, err
		}
		val = episode
	}

	return val, nil
}

//SearchByTitle performs an API search for a specified movie or series or episode by
//the specific title.
//Currently OMDB API seems to be returning only movie information, on a title
//based search. Year parameter has to be greater tha nor equal to 1888 (trivia: which
//is the world's earliest surviving motion-picture film?)
func (c *Client) SearchByTitle(q QueryData) (interface{}, error) {

	params := url.Values{}

	if q.Title == "" {
		return nil, errors.New("omdb: Title is missing")
	}
	params.Add("t", q.Title)

	if q.SearchType != "" && q.SearchType != "movie" && q.SearchType != "series" && q.SearchType != "episode" {
		return nil, errors.New("omdb: Searchtype should be either blank or one of following: movie, series, episode")
	}
	if q.SearchType != "" {
		params.Add("type", q.SearchType)
	}

	if q.Year != "" {
		i, err := strconv.Atoi(q.Year)
		if err != nil {
			return nil, errors.New("omdb: Year should be either blank or a valid number")
		}
		if i < 1888 {
			return nil, errors.New("omdb: Year should be either blank or greater than 1887")
		}
	}
	if q.Year != "" {
		params.Add("y", q.Year)
	}

	if q.Plot != "" && q.Plot != "short" && q.Plot != "full" {
		return nil, errors.New("omdb: Plot should be either blank or one of following: short, full")
	}
	if q.Plot != "" {
		params.Add("plot", q.Plot)
	}

	res, err := c.requestOmdbAPI(params)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	envelope := resultEnvelope{}
	err = json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, err
	}

	if envelope.Response == "False" {
		return nil, errors.New("omdb: Error from OMDB API: " + envelope.Error)
	}

	var val interface{}

	switch envelope.Type {

	case "movie":
		movie := MovieResult{}
		err = json.Unmarshal(data, &movie)
		if err != nil {
			return nil, err
		}
		val = movie

	case "series":
		series := SeriesResult{}
		err = json.Unmarshal(data, &series)
		if err != nil {
			return nil, err
		}
		val = series

	case "episode":
		episode := EpisodeResult{}
		err = json.Unmarshal(data, &episode)
		if err != nil {
			return nil, err
		}
		val = episode
	}

	return val, nil
}

//SearchByText performs an API search based on given text and return a SearchResponse
//struct.
func (c *Client) SearchByText(q QueryData) (*SearchResponse, error) {

	params := url.Values{}

	if q.Title == "" {
		return nil, errors.New("omdb: Text to search (Title) is missing")
	}
	params.Add("s", q.Title)

	if q.SearchType != "" && q.SearchType != "movie" && q.SearchType != "series" && q.SearchType != "episode" {
		return nil, errors.New("omdb: Searchtype should be either blank or one of following: movie, series, episode")
	}
	if q.SearchType != "" {
		params.Add("type", q.SearchType)
	}

	if q.Year != "" {
		i, err := strconv.Atoi(q.Year)
		if err != nil {
			return nil, errors.New("omdb: Year omdb: is either blank or a valid number")
		}
		if i < 1888 {
			return nil, errors.New("omdb: Year should be either blank or greater than 1887")
		}
	}
	if q.Year != "" {
		params.Add("y", q.Year)
	}

	if q.Page != "" {
		i, err := strconv.Atoi(q.Page)
		if err != nil {
			return nil, errors.New("omdb: Page should be either blank or a valid number")
		}
		if i < 1 || i > 100 {
			return nil, errors.New("omdb: Page should be either blank or between 1 to 100 (inclusive of both)")
		}
	}
	if q.Page != "" {
		params.Add("page", q.Page)
	}

	res, err := c.requestOmdbAPI(params)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	searchresponse := SearchResponse{}
	err = json.Unmarshal(data, &searchresponse)
	if err != nil {
		return nil, err
	}

	if searchresponse.Response == "False" {
		return nil, errors.New("omdb: Error from OMDB API: " + searchresponse.Error)
	}

	return &searchresponse, nil
}
