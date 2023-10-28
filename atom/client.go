package atom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Client wrapped *http.Client and some methods for accessing atom feed are added
type Client struct {
	*http.Client
}

// GetFeed gets the blog feed
func (c *Client) GetFeed(url string) (*Feed, error) {
	resp, err := c.http("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return Parse(resp.Body)
}

// GetEntry gets the blog entry
func (c *Client) GetEntry(url string) (*Entry, error) {
	resp, err := c.http("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return ParseEntry(resp.Body)
}

// PutEntry puts the blog entry
func (c *Client) PutEntry(url string, e *Entry) (*Entry, error) {
	body := new(bytes.Buffer)

	body.WriteString(xml.Header)
	err := xml.NewEncoder(body).Encode(e)
	if err != nil {
		return nil, err
	}

	resp, err := c.http("PUT", url, body)
	if err != nil {
		return nil, err
	}

	newEntry, err := ParseEntry(resp.Body)
	if err != nil {
		return nil, err
	}

	return newEntry, nil
}

// PostEntry posts the blog entry
func (c *Client) PostEntry(url string, e *Entry) (*Entry, error) {
	body, err := entryBody(e)
	if err != nil {
		return nil, err
	}

	resp, err := c.http("POST", url, body)
	if err != nil {
		return nil, err
	}

	newEntry, err := ParseEntry(resp.Body)
	if err != nil {
		return nil, err
	}

	return newEntry, nil
}

func entryBody(e *Entry) (*bytes.Buffer, error) {
	body := new(bytes.Buffer)

	body.WriteString(xml.Header)
	err := xml.NewEncoder(body).Encode(e)
	if err != nil {
		return nil, err
	}

	return body, nil
}

var blogsyncDebug = os.Getenv("BLOGSYNC_DEBUG") != ""

type traceDump struct {
	RequestBody  string `yaml:",omitempty"`
	ResponseBody string
	Method       string
	URL          string
	Code         int
}

func (c *Client) http(method, url string, body io.Reader) (*http.Response, error) {
	td := traceDump{}
	if blogsyncDebug && body != nil {
		bb, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		td.RequestBody = string(bb)
		body = strings.NewReader(td.RequestBody)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		bytes, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("got [%s]: %q", resp.Status, string(bytes))
	}

	if blogsyncDebug {
		bb, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		td.ResponseBody = string(bb)
		td.Method = method
		td.URL = url
		td.Code = resp.StatusCode
		fmt.Printf("%+v\n", td)
		resp.Body = io.NopCloser(strings.NewReader(td.ResponseBody))
	}
	return resp, nil
}
