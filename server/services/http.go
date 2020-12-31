package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/vickey290/go-modules/server/models"
)

type HTTPClient struct {
	notifierURI string
	client      *http.Client
}

func NewHTTPClient(uri string) HTTPClient {
	return HTTPClient{
		notifierURI: uri,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

type NotificationResponse struct {
	completed bool
	duration  time.Duration
}

func (c HTTPClient) Notify(reminder models.Reminder) (NotificationResponse, error) {
	var notifierResponse struct {
		ActivationType  string `json:"activationType"`
		ActivationValue string `json:"activationValue"`
	}
	bs, err := json.Marshal(reminder)
	if err != nil {
		e := models.WrapError("could not marshal json", err)
		return NotificationResponse{}, e
	}

	res, err := c.client.Post(c.notifierURI+"/notify",
		"application/json", bytes.NewReader(bs))
	if err != nil {
		e := models.WrapError("could not marshal json", err)
		return NotificationResponse{}, e
	}
	err = json.NewDecoder(res.Body).Decode(&notifierResponse)
	if err != nil && err != io.EOF {
		e := models.WrapError("could not decode notifier response", err)
		return NotificationResponse{}, e
	}

	t := notifierResponse.ActivationType
	v := notifierResponse.ActivationValue

	if t == "closed" {
		return NotificationResponse{completed: true}, nil
	}

	d, err := time.ParseDuration(v)
	if err != nil && d != 0 {
		e := models.WrapError("could not parse notifier duration", err)
		return NotificationResponse{}, e
	}
	if d == 0 {
		return NotificationResponse{}, errors.New("notification duration must be > 0s")
	}

	return NotificationResponse{duration: d}, nil
}
