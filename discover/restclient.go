/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package discover

import (
	"net/http"
	"time"
)

// RestClient is the recommended approach to do rest request in server
type RestClient struct {
	client  http.Client
	baseUrl string
	Timeout time.Duration
}

func (s *RestClient) Do(r *http.Request) (*http.Response, error) {
	return s.client.Do(r)
}

type RestRequester interface {
	Do(r *http.Request) (*http.Response, error)
}

func NewRestClient(timeout time.Duration, service *MicroService) *RestClient {
	return &RestClient{
		client:  http.Client{Timeout: timeout},
		Timeout: timeout,
		baseUrl: service.Address + ":" + string(service.Port),
	}
}
