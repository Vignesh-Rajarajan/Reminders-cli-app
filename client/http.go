package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	client     *http.Client
	BackendURI string
}

type reminderBody struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
}

func NewHTTPClient(uri string) HTTPClient {
	return HTTPClient{
		BackendURI: uri,
		client:     &http.Client{},
	}
}

func (c HTTPClient) Create(title, message string, duration time.Duration) ([]byte, error) {
	requestBody := reminderBody{
		Title:    title,
		Message:  message,
		Duration: duration,
	}
	return c.apiCall(
		http.MethodPost,
		"/reminders",
		&requestBody,
		http.StatusCreated,
	)
}

func (c HTTPClient) Edit(id, title, message string, duration time.Duration) ([]byte, error) {
	requestBody := reminderBody{
		ID:       id,
		Title:    title,
		Message:  message,
		Duration: duration,
	}
	return c.apiCall(
		http.MethodPatch,
		"/reminders",
		&requestBody,
		http.StatusOK,
	)
}
func (c HTTPClient) Fetch(ids []string) ([]byte, error) {

	id := strings.Join(ids, ",")

	return c.apiCall(
		http.MethodGet,
		"/reminders"+id,
		nil,
		http.StatusOK,
	)
}

func (c HTTPClient) Delete(ids []string) error {

	id := strings.Join(ids, ",")
	_, err := c.apiCall(
		http.MethodDelete,
		"/reminders"+id,
		nil,
		http.StatusNoContent,
	)
	return err
}

func (c HTTPClient) Healthy(host string) bool {
	res, err := http.Get(host + "/health")
	if err != nil || res.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func (c HTTPClient) apiCall(method, path string, body interface{}, respCode int) ([]byte, error) {
	req, err := json.Marshal(body)

	if err != nil {
		e := wrapError("could not marshal request body", err)
		return nil, e
	}
	reqBody, err := http.NewRequest(method, c.BackendURI+path, bytes.NewReader(req))

	if err != nil {
		e := wrapError("could not create the request", err)
		return []byte{}, e
	}

	resp, err := c.client.Do(reqBody)

	if err != nil {
		e := wrapError("could not make the http call", err)
		return []byte{}, e
	}

	response, err := c.readResBody(resp.Body)

	if err != nil {
		return []byte{}, err
	}
	if resp.StatusCode != respCode {
		if len(response) > 0 {
			fmt.Printf("got this response body:\n%s\n", response)
		}
		return []byte{}, fmt.Errorf(
			"expected response code: %d, got: %d",
			respCode,
			resp.StatusCode,
		)
	}

	return []byte(response), err

}

func (c HTTPClient) readResBody(b io.Reader) (string, error) {
	result, err := ioutil.ReadAll(b)

	if err != nil {
		return "", wrapError("could not read response body", err)
	}
	if len(result) == 0 {
		return "", nil
	}

	var buff bytes.Buffer
	if err := json.Indent(&buff, result, "", "\t"); err != nil {
		return "", wrapError("could not indent json", err)
	}
	return buff.String(), nil
}
