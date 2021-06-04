package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func HttpPost(url, params string, contentType string, retMap interface{}) error {
	if contentType == "" {
		contentType = "application/json"
	}

	resp, err := http.Post(url,
		contentType,
		strings.NewReader(params))

	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, retMap)
	if err != nil {
		return err
	}

	return nil
}

func HttpGet(url string, retMap interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, retMap)
	if err != nil {
		return err
	}

	return nil
}
