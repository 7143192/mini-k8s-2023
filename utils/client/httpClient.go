package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	rootURL string
	port    string
}

func NewClient(url, port string) *Client {
	return &Client{
		rootURL: url,
		port:    port,
	}
}

func (c *Client) DelRequest(uri string) (int, string, error) {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:%s%s", c.rootURL, c.port, uri), nil)
	if err != nil {
		return -1, "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return -1, "", err
	}
	result := make(map[string]string)
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return -1, "", err
	}
	if response.StatusCode == http.StatusOK {
		return http.StatusOK, result["INFO"], nil
	} else {
		return http.StatusBadRequest, result["ERROR"], nil
	}
}

func (c *Client) PostRequest(body *bytes.Reader, uri string) (string, error) {
	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s%s", c.rootURL, c.port, uri), body)
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	result := make(map[string]string, 0)
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("[ERROR]: %s\n", result["ERROR"]), nil
	} else {
		return fmt.Sprintf("[INFO]: %s\n", result["INFO"]), nil
	}
}

func (c *Client) GetRequest(uri string) (int, []byte, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s%s", c.rootURL, c.port, uri), nil)
	//TODO: We may enable the get-request to contain paras just like below
	//q := req.URL.Query()
	//q.Add("api_key", "key_from_environment_or_flag")
	//q.Add("another_thing", "foo & bar")
	//req.URL.RawQuery = q.Encode()
	//
	if err != nil {
		return -1, []byte(""), err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return -1, []byte(""), err
	}
	result := make(map[string][]byte)
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return -1, []byte(""), err
	}
	if response.StatusCode == http.StatusOK {
		return http.StatusOK, result["INFO"], nil
	} else {
		return http.StatusBadRequest, result["ERROR"], nil
	}
}

func (c *Client) GetRequestStr(uri string) (int, string, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s%s", c.rootURL, c.port, uri), nil)
	//TODO: We may enable the get-request to contain paras just like below
	//q := req.URL.Query()
	//q.Add("api_key", "key_from_environment_or_flag")
	//q.Add("another_thing", "foo & bar")
	//req.URL.RawQuery = q.Encode()
	//
	if err != nil {
		return -1, "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return -1, "", err
	}
	result := make(map[string]string)
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return -1, "", err
	}
	if response.StatusCode == http.StatusOK {
		return http.StatusOK, result["INFO"], nil
	} else {
		return http.StatusBadRequest, result["ERROR"], nil
	}
}
