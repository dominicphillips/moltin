package moltin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const host = "https://api.molt.in"

type Client struct {
	Id     string
	Secret string
	Token  *struct {
		Access_token string
		Token_type   string
		Expires      int64
		Expires_in   int
	}
}

type MoltinError map[string]interface{}

func (m MoltinError) Error() string {
	return fmt.Sprintf("moltin error: status %v - %v", m["status"], m["error"])
}

func NewError(res *http.Response) *MoltinError {
	var err MoltinError
	json.NewDecoder(res.Body).Decode(&err)
	return &err
}

func NewClient(id, secret string) (*Client, error) {
	c := Client{Id: id, Secret: secret}
	err := c.authenticate()
	return &c, err
}

func (c *Client) request(method, path string, body io.Reader, v interface{}) error {
	r, _ := http.NewRequest(method, host+path, body)
	if c.Token != nil {
		t := time.Unix(c.Token.Expires, 0)
		if t.Sub(time.Now()).Minutes() < 5 {
			// reauth
			c.authenticate()
		}
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token.Access_token))
	}

	client := http.Client{}
	res, err := client.Do(r)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return NewError(res)
	}

	return json.NewDecoder(res.Body).Decode(&v)

}

func (c *Client) authenticate() error {
	data := url.Values{}
	data.Set("client_id", c.Id)
	data.Add("client_secret", c.Secret)
	data.Add("grant_type", "client_credentials")
	return c.request("POST", "/oauth/access_token", bytes.NewBufferString(data.Encode()), &c.Token)

}

func (c *Client) GetProduct(id int) (interface{}, error) {
	var product interface{}
	err := c.request("GET", fmt.Sprintf("/v1/products/%d", id), nil, &product)
	return product, err
}
