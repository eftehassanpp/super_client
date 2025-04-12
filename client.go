package super_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

	"github.com/eftehassanpp/super_client/utils"
)

var Chrome133 = profiles.Chrome_133
var Firefox135 = profiles.Firefox_135

func isJSONResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json")
}

type SuperResponse struct {
	URL         *url.URL
	Status      string // e.g. "200 OK"
	StatusCode  int    // e.g. 200
	Body        []byte //original response body
	Text        string //response body converted to string
	Headers     http.Header
	Cookies     []*http.Cookie
	Error       error
	Unmarshaled map[string]any
}

func (sr *SuperResponse) GetCookie(name string) *http.Cookie {
	for _, cookie := range sr.Cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

type SuperClient struct {
	Client         tls_client.HttpClient
	Jar            tls_client.CookieJar
	Profile        profiles.ClientProfile
	DefaultHeaders *map[string]string
	HeaderOrder    *[]string
	Timeout        int
	BaseProxy      string
	SecCH          string
	Platform       string
	UserAgent      string
	AcceptLang     string
}

type ClientConfig struct {
	HeaderOrder *[]string
	BaseProxy   string
	Timeout     int
	SecCH       string
	Platform    string
	UserAgent   string
	AcceptLang  string
	Profile     *profiles.ClientProfile
}

func (sc *SuperClient) Init(conf *ClientConfig) error {
	sc.Jar = tls_client.NewCookieJar()
	if conf.Timeout != 0 {
		sc.Timeout = conf.Timeout
	}
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(sc.Timeout),
		tls_client.WithClientProfile(sc.Profile),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(sc.Jar),
		tls_client.WithRandomTLSExtensionOrder(),
	}

	var err error
	sc.Client, err = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return err
	}
	sc.HeaderOrder = &RegularHOrder //setting default header order
	// updating config value to client
	if conf.UserAgent != "" {
		sc.UserAgent = conf.UserAgent
	}
	if conf.HeaderOrder != nil {
		sc.HeaderOrder = conf.HeaderOrder
	}
	if conf.AcceptLang != "" {
		sc.AcceptLang = conf.AcceptLang
	}
	if conf.Profile != nil {
		sc.Profile = *conf.Profile
	}
	if conf.BaseProxy != "" {
		sc.BaseProxy = conf.BaseProxy
		err = sc.SetNewProxy()
		if err != nil {
			return err
		}
	}
	if sc.Profile.GetClientHelloId().Client == "Firefox" {
		if conf.UserAgent == "" {
			sc.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0"
		}
		sc.DefaultHeaders = &map[string]string{
			"upgrade-insecure-requests": "1",
			"user-agent":                sc.UserAgent,
			"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"sec-fetch-site":            "none",
			"sec-fetch-mode":            "navigate",
			"sec-fetch-user":            "?1",
			"sec-fetch-dest":            "document",
			"accept-encoding":           "gzip, deflate, br, zstd",
			"accept-language":           sc.AcceptLang,
		}
	} else {
		if conf.SecCH != "" {
			sc.SecCH = conf.SecCH
		}
		if conf.Platform != "" {
			sc.Platform = conf.Platform
		}
		sc.DefaultHeaders = &map[string]string{
			"sec-ch-ua":                 sc.SecCH,
			"sec-ch-ua-mobile":          "?0",
			"sec-ch-ua-platform":        sc.Platform,
			"upgrade-insecure-requests": "1",
			"user-agent":                sc.UserAgent,
			"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"sec-fetch-site":            "none",
			"sec-fetch-mode":            "navigate",
			"sec-fetch-user":            "?1",
			"sec-fetch-dest":            "document",
			"accept-encoding":           "gzip, deflate, br, zstd",
			"accept-language":           sc.AcceptLang,
			"priority":                  "u=0, i",
		}
	}

	return nil
}
func (sc *SuperClient) SetNewProxy() error {
	randomProxySession := utils.GenerateRandomString(12)
	randomProxy := fmt.Sprintf(sc.BaseProxy, randomProxySession)
	err := sc.Client.SetProxy(randomProxy)
	return err
}
func (sc *SuperClient) SetProxy(proxyUrl string) error {
	err := sc.Client.SetProxy(proxyUrl)
	return err
}

func (sc *SuperClient) SetFollowRedirect(followRedirect bool) {
	sc.Client.SetFollowRedirect(followRedirect)
}

// Cookies returns the cookies for the SuperClient
func (sc *SuperClient) GetAllCookies() map[string][]*http.Cookie {
	return sc.Jar.GetAllCookies()
}
func (sc *SuperClient) Cookies(host string) []*http.Cookie {
	return sc.Jar.Cookies(&url.URL{Host: host})
}

// GetCookie returns a cookie by name and host
func (sc *SuperClient) GetCookie(name string, host string) *http.Cookie {
	cookies := sc.Jar.Cookies(&url.URL{Host: host})
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func (sc *SuperClient) SetCookie(name string, value string, host string) {
	cookie := &http.Cookie{
		Name:   name,
		Value:  value,
		Domain: host,
		Path:   "/",
	}
	cookies := []*http.Cookie{cookie}
	sc.Jar.SetCookies(&url.URL{Host: host}, cookies)
}
func (sc *SuperClient) SetCookies(cookies []*http.Cookie, host string) {
	sc.Jar.SetCookies(&url.URL{Host: host}, cookies)
}
func (sc *SuperClient) SetHeaderOrder(headerOrder *[]string) {
	sc.HeaderOrder = headerOrder
}

// CurrentProxy returns the current proxy for the SuperClient
func (sc *SuperClient) CurrentProxy() string {
	return sc.Client.GetProxy()
}

func (sc *SuperClient) MakeRequest(req *http.Request, headers map[string]string) (*http.Response, error) {
	for key, value := range *sc.DefaultHeaders {
		req.Header[key] = []string{value}
	}
	for key, value := range headers {
		req.Header[key] = []string{value}
	}
	for _, item := range *sc.HeaderOrder {
		req.Header.Add(http.HeaderOrderKey, item)
	}
	resp, err := sc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	sc.Jar.SetCookies(resp.Request.URL, resp.Cookies())
	return resp, nil
}

func (sc *SuperClient) Get(url string, params map[string]string, headers map[string]string) (response *SuperResponse) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return &SuperResponse{Error: fmt.Errorf("error creating GET request for URL %s: %v", url, err)}
	}
	if params != nil {
		query := req.URL.Query()
		for key, value := range params {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}
	resp, err := sc.MakeRequest(req, headers)
	if err != nil {
		return &SuperResponse{Error: fmt.Errorf("error excecuting GET request for URL %s: %v", url, err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SuperResponse{Error: fmt.Errorf("error reading response body: %v", err)}
	}
	response = &SuperResponse{
		URL:        resp.Request.URL,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
		Text:       string(body),
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
		Error:      nil,
	}
	if isJSONResponse(resp) {
		json.Unmarshal([]byte(response.Text), &response.Unmarshaled)
	}
	return response
}

func (sc *SuperClient) Post(url string, params map[string]string, headers map[string]string, data *string, jsonData any) (response *SuperResponse) {
	var payload []byte

	// If data is provided, use it as form-data (text)
	if data != nil {
		payload = []byte(*data)
	} else if jsonData != nil {
		// If jsonData is provided, marshal it into JSON
		var err error
		payload, err = json.Marshal(jsonData)
		if err != nil {
			return &SuperResponse{Error: fmt.Errorf("error marshaling JSON data: %v", err)}
		}
		headers["content-type"] = "application/json"
	} else {
		return &SuperResponse{Error: fmt.Errorf("either data or jsonData must be provided")}
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return &SuperResponse{Error: fmt.Errorf("error creating POST request for URL %s: %v", url, err)}
	}
	if params != nil {
		query := req.URL.Query()
		for key, value := range params {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}
	resp, err := sc.MakeRequest(req, headers)
	if err != nil {
		return &SuperResponse{Error: fmt.Errorf("error excecuting GET request for URL %s: %v", url, err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SuperResponse{Error: fmt.Errorf("error reading response body: %v", err)}
	}
	response = &SuperResponse{
		URL:        resp.Request.URL,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
		Text:       string(body),
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
		Error:      nil,
	}
	if isJSONResponse(resp) {
		json.Unmarshal([]byte(response.Text), &response.Unmarshaled)
	}
	return response
}

func NewSuperClient(conf *ClientConfig) (*SuperClient, error) {
	superClient := &SuperClient{Timeout: 30,
		SecCH:      `"Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"`,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		AcceptLang: "en-GB,en;q=0.9",
		Platform:   `"Windows"`,
		Profile:    profiles.Chrome_133,
	} //setting default values
	err := superClient.Init(conf)
	return superClient, err
}
