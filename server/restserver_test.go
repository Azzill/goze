/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package server

import (
	"encoding/json"
	"fmt"
	"github.com/azzill/goze/common"
	"net/http"
	"testing"
)

type CustomResponse struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

type CustomResponseWrapper struct{}

func (CustomResponseWrapper) Wrap(v interface{}, wr http.ResponseWriter) bool {
	////Only wrap struct as json
	//if reflect.TypeOf(v).Kind() != reflect.Struct {
	//	return false
	//}
	wr.Header().Set("Content-Type", "application/json")
	if bytes, err := json.Marshal(CustomResponse{Code: 0, Msg: v}); err != nil {
		http.Error(wr, err.Error(), 500)
	} else {
		_, err = wr.Write(bytes)
	}
	return true
}
func TestServer(t *testing.T) {
	fmt.Println("You can test by opening urls below in 30s")
	fmt.Println("http://127.0.0.1:8080/")
	fmt.Println("http://127.0.0.1:8080/world")
	fmt.Println("http://127.0.0.1:8080/yet/another/url")

	restServer := NewRestServer(":8080", &HttpConfig{})

	restServer.Mapping(Get, "/", func(ctx *common.RequestCtx) (i interface{}) {
		return "Welcome home"
	})

	restServer.Mapping(Get, "/yet/another/url", func(ctx *common.RequestCtx) (i interface{}) {
		return "Yet another rest-server"
	})

	restServer.Mapping(Get, "/:name", func(ctx *common.RequestCtx) (i interface{}) {
		return fmt.Sprintf("Hello %s!", ctx.PathVariable["name"])
	})

	wrapper := CustomResponseWrapper{}
	restServer.AddResponseWrapper(wrapper)
	restServer.StartServer()

}
