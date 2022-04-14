package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//	map[access_token:IKpL6k6mpUzpwgHjdkthCVdSqj412bRy6JtnzNyTMPw created_at:1.649969596e+09 expires_in:86400 refresh_token:maShLmt7_ukrarX6LNLjPFnI5wBtXWMupN2EnWHdS34 scope:user_rates comments topics token_type:Bearer]
func ShikiGetToken(target interface{}) error {
	spaceClient := &http.Client{
		Timeout: time.Second * 2,
	}
	data := url.Values{}
	data.Set("grant_type", `authorization_code`)
	data.Add("client_id", `AQwAdPILflg8RVN5XbDqloaGgGrnCQhMkqP5rCLzU2k`)
	data.Add("client_secret", "yrNkdwMlE1w5O6kwzVBh8bXpOKISGVl4wrIVQumeUeI")
	data.Add("code", "AK3F_MTYQKHs0rkH9psSL9LExVFXcQ30d2uG7HDQo10")
	data.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")

	req, err := http.NewRequest(http.MethodPost, `https://shikimori.one/oauth/token`, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "funi-funi")
	res, err := spaceClient.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}

func ShikiRefreshToken(refresh string, target interface{}) error {
	spaceClient := &http.Client{
		Timeout: time.Second * 2,
	}
	data := url.Values{}
	data.Set("grant_type", `refresh_token`)
	data.Add("client_id", `AQwAdPILflg8RVN5XbDqloaGgGrnCQhMkqP5rCLzU2k`)
	data.Add("client_secret", "yrNkdwMlE1w5O6kwzVBh8bXpOKISGVl4wrIVQumeUeI")
	data.Add("refresh_token", refresh)

	req, err := http.NewRequest(http.MethodPost, `https://shikimori.one/oauth/token`, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "funi-funi")
	res, err := spaceClient.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}

func ShikiGetTopics(access string, target interface{}) error {
	spaceClient := &http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequest(http.MethodGet, `https://shikimori.one/api/topics?limit=1&forum=news&page=1`, nil)
	if err != nil {
		return err
	}

	//req.Header.Set("User-Agent", "funi-funi")
	//req.Header.Set("Authorization", `Bearer `+access)
	res, err := spaceClient.Do(req)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}
	return nil
}
