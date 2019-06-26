/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/azzill/goze/server"
	"io/ioutil"
	"net/http"
	"strings"
)

func RestRequest(method server.RequestMethod, url string, reqBody interface{}, respBody interface{}) (bool, error) {
	buf, err := json.Marshal(reqBody)

	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(string(method), url, bytes.NewReader(buf))
	if err != nil {
		return false, err
	}

	//json request
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	//Only accept OK status
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(fmt.Sprintf("returned status code is %d", resp.StatusCode))
	}

	//Do not need to parse response body
	if respBody == nil {
		return false, nil
	}

	//No body
	if resp.ContentLength == 0 {
		return false, nil
	}

	switch resp.Header.Get("Content-Type") {
	case "application/json":
		b, err := ioutil.ReadAll(resp.Body)
		defer func() {
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return true, json.Unmarshal(b, respBody)
	case "text/plain":
		all, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		*respBody.(*string) = string(all)
		return true, nil
	default:
		return false, nil
	}
}

func ParseIPAddr(withPort string) string {
	return withPort[:strings.Index(withPort, ":")]
}
