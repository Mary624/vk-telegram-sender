package vk

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"vk-telegram/lib/e"
)

const (
	groups_GetLongPollServer   = "groups.getLongPollServer"
	messages_GetLongPollServer = "messages.getLongPollServer"
	messages_send              = "messages.send"
)
const pathMethod = "method"
const waitSec = 25

type Client struct {
	server string
	key    string
	token  string
	v      string
	host   string
	client http.Client
}

func NewClient(token, host, v string) *Client {
	return &Client{
		token:  token,
		v:      v,
		host:   host,
		client: http.Client{},
	}
}

func (c *Client) Updates(ts int) (*ResultTsUpdates, error) {
	data, err := c.doRequestUpdates(ts)
	if err != nil {
		return nil, err
	}

	var res ResultTsUpdates
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if len(res.Updates) == 0 {
		log.Print("0 updates")
	}

	return &res, nil
}

func (c *Client) SendMessage(text string, user_id, ts int) error {
	err := c.doRequestSendMessage(text, user_id, ts)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Connect(group_id int) (int, error) {
	data, err := c.doRequestConnectionParams(group_id)
	if err != nil {
		return 0, err
	}

	var res Connection
	if err := json.Unmarshal(data, &res); err != nil {
		return 0, err
	}

	c.key = res.Response.Key
	c.server = res.Response.Server

	ts, err := strconv.Atoi(res.Response.Ts)
	if err != nil {
		return 0, err
	}
	return ts, nil
}

func (c *Client) doRequestConnectionParams(group_id int) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can't connect", err) }()

	q := url.Values{}
	q.Add("access_token", c.token)
	q.Add("v", c.v)
	q.Add("group_id", strconv.Itoa(group_id))
	q.Add("scrope", "manage")

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(pathMethod, groups_GetLongPollServer),
	}

	res, err := doRequest(u.String(), q, c.client)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) doRequestSendMessage(text string, user_id, ts int) (err error) {
	defer func() { err = e.WrapIfErr("can't send message", err) }()

	q := url.Values{}
	q.Add("access_token", c.token)
	q.Add("v", c.v)
	q.Add("random_id", strconv.Itoa(0))
	q.Add("user_id", strconv.Itoa(user_id))
	q.Add("message", text)

	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(pathMethod, messages_send),
	}

	_, err = doRequest(u.String(), q, c.client)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doRequestUpdates(ts int) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can't do request", err) }()

	q := url.Values{}
	q.Add("act", "a_check")
	q.Add("key", c.key)
	q.Add("ts", strconv.Itoa(ts))
	q.Add("wait", strconv.Itoa(waitSec))

	res, err := doRequest(c.server, q, c.client)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func doRequest(u string, query url.Values, client http.Client) (data []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
