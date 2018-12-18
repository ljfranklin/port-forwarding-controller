package unifi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"

	"github.com/ljfranklin/port-forwarding-controller/pkg/forwarding"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	HTTPClient    *http.Client
	ControllerURL string
	Username      string
	Password      string
}

type baseResponse struct {
	Meta meta `json:"meta"`
}
type meta struct {
	Msg string `json:"msg"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type listResponse struct {
	Data []listItem `json:"data"`
}
type listItem struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
	Port string `json:"fwd_port"`
	IP   string `json:"fwd"`
	Src  string `json:"src"`
}

type createRequest struct {
	Name    string `json:"name"`
	Port    string `json:"fwd_port"`
	IP      string `json:"fwd"`
	DstPort string `json:"dst_port"`
	Enabled bool   `json:"enabled"`
	Proto   string `json:"proto"`
	Src     string `json:"src"`
}

func (c Client) ListAddresses(options map[string]string) ([]forwarding.Address, error) {
	listResp, err := c.list(options)
	if err != nil {
		return nil, err
	}

	addresses := []forwarding.Address{}
	for _, a := range listResp.Data {
		p, err := strconv.Atoi(a.Port)
		if err != nil {
			return nil, err
		}
		src := a.Src
		if src == "any" {
			src = ""
		}
		addresses = append(addresses, forwarding.Address{
			Name:        a.Name,
			Port:        p,
			IP:          a.IP,
			SourceRange: src,
			Options:     options,
		})
	}

	return addresses, nil
}

func (c Client) list(options map[string]string) (listResponse, error) {
	endpoint := fmt.Sprintf("%s/api/s/%s/rest/portforward", c.ControllerURL, c.siteName(options))
	respBody, err := c.makeAPICall("GET", endpoint, nil)
	if err != nil {
		return listResponse{}, err
	}

	var listResp listResponse
	err = json.Unmarshal(respBody, &listResp)
	if err != nil {
		return listResponse{}, err
	}
	return listResp, nil
}

func (c Client) siteName(options map[string]string) string {
	if name, ok := options["unifi-site"]; ok {
		return name
	}
	return "default"
}

func (c Client) CreateAddress(address forwarding.Address) error {
	src := "any"
	if address.SourceRange != "" {
		src = address.SourceRange
	}
	reqBody, err := json.Marshal(createRequest{
		Name:    address.Name,
		Port:    fmt.Sprintf("%d", address.Port),
		IP:      address.IP,
		DstPort: fmt.Sprintf("%d", address.Port),
		Enabled: true,
		Proto:   "tcp_udp",
		Src:     src,
	})
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/api/s/%s/rest/portforward", c.ControllerURL, c.siteName(address.Options))
	// TODO: do we need content-type?
	if _, err := c.makeAPICall("POST", endpoint, reqBody); err != nil {
		return err
	}

	return nil
}

func (c Client) DeleteAddress(address forwarding.Address) error {
	listResp, err := c.list(address.Options)
	if err != nil {
		return err
	}

	matchingID := ""
	for _, a := range listResp.Data {
		if a.Name == address.Name &&
			a.Port == fmt.Sprintf("%d", address.Port) &&
			a.IP == address.IP {
			matchingID = a.ID
			break
		}
	}

	if matchingID == "" {
		return nil
	}

	endpoint := fmt.Sprintf("%s/api/s/%s/rest/portforward/%s", c.ControllerURL, c.siteName(address.Options), matchingID)
	_, err = c.makeAPICall("DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) makeAPICall(method, endpoint string, body []byte) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if c.isLoginError(resp) {
		if err := c.login(); err != nil {
			return nil, err
		}
		return c.makeAPICall(method, endpoint, body)
	} else if resp.StatusCode >= 300 {
		return nil, c.buildRespErr(resp)
	}

	return ioutil.ReadAll(resp.Body)
}

func (c Client) isLoginError(resp *http.Response) bool {
	if resp.StatusCode == 401 {
		var decodedResp baseResponse
		err := json.NewDecoder(resp.Body).Decode(&decodedResp)
		if err != nil {
			return false
		}
		if decodedResp.Meta.Msg == "api.err.LoginRequired" {
			return true
		}
	}
	return false
}

func (c Client) login() error {
	if c.HTTPClient.Jar == nil {
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return err
		}
		c.HTTPClient.Jar = jar
	}

	reqBody, err := json.Marshal(loginRequest{
		Username: c.Username,
		Password: c.Password,
	})
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/api/login", c.ControllerURL)
	resp, err := c.HTTPClient.Post(endpoint, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return c.buildRespErr(resp)
	}

	return nil
}

func (c Client) buildRespErr(resp *http.Response) error {
	if resp.Body == nil {
		return fmt.Errorf("Invalid response code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Invalid response code %d", resp.StatusCode)
	}

	return fmt.Errorf("Invalid response code %d: %s", resp.StatusCode, string(body))
}
