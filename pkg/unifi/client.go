package unifi

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type listResponse struct {
	Data []listItem `json:"data"`
}
type listItem struct {
	Name string `json:"name"`
	Port string `json:"fwd_port"`
	IP   string `json:"fwd"`
}

func (c Client) ListAddresses() ([]forwarding.Address, error) {
	if err := c.login(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/api/s/default/rest/portforward", c.ControllerURL)
	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, c.buildRespErr(resp)
	}

	var listResp listResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	if err != nil {
		return nil, err
	}

	addresses := []forwarding.Address{}
	for _, a := range listResp.Data {
		p, err := strconv.Atoi(a.Port)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, forwarding.Address{
			Name: a.Name,
			Port: p,
			IP:   a.IP,
		})
	}

	return addresses, nil
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
