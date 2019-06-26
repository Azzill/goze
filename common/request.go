/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package common

import (
	sql2 "database/sql"
	"encoding/json"
	"github.com/azzill/goze/log"
	"github.com/azzill/goze/sql"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

type RequestCtx struct {
	QueryString    map[string][]string
	PathVariable   map[string]string
	Request        *http.Request
	Form           *multipart.Form
	Tx             *sql.Tx
	ResponseWriter http.ResponseWriter
	//=========
	sql     *sql.SQL
	txBegan bool
}

func NewRequestCtx(queryString map[string][]string, pathVariable map[string]string, request *http.Request, form *multipart.Form, responseWriter http.ResponseWriter, sqls *sql.SQL) *RequestCtx {
	var db *sql2.DB
	if sqls != nil {
		db = sqls.Db
	}
	return &RequestCtx{
		QueryString:    queryString,
		PathVariable:   pathVariable,
		Request:        request,
		Form:           form,
		Tx:             sql.NewTx(db, &sql.UnTx{}),
		ResponseWriter: responseWriter,
		sql:            sqls,
		txBegan:        false,
	}
}

var logger = log.NewLogger("RestServer")

func (c *RequestCtx) BeginTx() {
	//if tx began, commit it
	if c.txBegan {
		if e := c.Tx.Commit(); e != nil {
			panic(e.Error())
		}
		c.txBegan = false
	}
	c.Tx = c.sql.BeginTx()
	c.txBegan = true

}
func (c *RequestCtx) ParseBody(dst interface{}) error {
	ct := c.Request.Header.Get("Content-Type")
	if !strings.EqualFold(ct, "application/json") {
		logger.Info("content-type: '%s' is not supported yet\n", ct)
		return nil
	}

	//TODO support url-encoded

	bytes, e := ioutil.ReadAll(c.Request.Body) //Note that ReadAll is not safe for oom attack
	defer func() {
		_ = c.Request.Body.Close()
	}()

	if e != nil {
		return e
	}

	e = json.Unmarshal(bytes, dst)
	if e != nil {
		return e
	}
	return nil
}
