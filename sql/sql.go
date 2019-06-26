/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package sql

import (
	"context"
	"database/sql"
	"reflect"
)

type SQL struct {
	Db *sql.DB
}

type Tx struct {
	executor Executor
	tx       Transaction
}

func NewTx(executor Executor, tx Transaction) *Tx {
	return &Tx{executor: executor, tx: tx}
}

func NewSQL(dataSource string, driver string) *SQL {
	context.Background()
	s := &SQL{}
	db, e := sql.Open(driver, dataSource)
	if e != nil {
		panic("Unable to connect to: " + dataSource)
	}
	s.Db = db
	return s
}

func (s *SQL) BeginTx() *Tx {
	if tx, e := s.Db.Begin(); e != nil {
		panic(e.Error())
	} else {
		return &Tx{executor: tx, tx: tx}
	}
}

func (tx *Tx) Query(dest interface{}, sql string, param ...interface{}) (int64, error) {
	if reflect.TypeOf(dest).Elem().Kind() == reflect.Slice || reflect.TypeOf(dest).Kind() == reflect.Array {
		return tx.queryAll(dest, sql, param...)
	} else {
		return tx.queryOne(dest, sql, param...)
	}
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

func (tx *Tx) queryOne(dest interface{}, sql string, param ...interface{}) (int64, error) {
	rows, e := tx.executor.Query(sql, param...)
	if e != nil {
		//TODO logger
		return 0, e
	}
	defer func() {
		_ = rows.Close()
	}()

	columns, e := rows.Columns()
	if e != nil {
		return 0, e
	}

	elem := reflect.TypeOf(dest).Elem()
	vals := make([]interface{}, len(columns))
	maps := make([]interface{}, len(columns))
	if len(columns) == 1 {
		if rows.Next() {
			e := rows.Scan(dest)
			if e != nil {
				return 0, e
			}
			return 1, nil
		}
		return 0, e
	}
	for i := 0; i < elem.NumField(); i++ {
		for j, col := range columns {
			if col == elem.Field(i).Tag.Get("col") {
				maps[j] = i //column[j] map to elem.field[i]
				reflect.ValueOf(vals).Index(j).Set(reflect.New(elem.Field(i).Type.Elem()))
			}
		}
	}

	if rows.Next() {
		e := rows.Scan(vals...)
		if e != nil {
			return 0, e
		}

		for i := 0; i < elem.NumField(); i++ {
			for j, col := range columns {
				if col == elem.Field(i).Tag.Get("col") {
					reflect.ValueOf(dest).Elem().Field(i).Set(reflect.ValueOf(vals[j]))
				}
			}
		}
	} else {
		return 0, e

	}

	return 1, nil
}

//dest must be the pointer to the slice, not the slice reference : &[]slice
func (tx *Tx) queryAll(dest interface{}, sql string, param ...interface{}) (int64, error) {
	var total int64

	rows, e := tx.executor.Query(sql, param...)
	if e != nil {
		////TODO logger
		return 0, e
	}

	defer func() {
		_ = rows.Close()
	}()

	columns, e := rows.Columns()
	if e != nil {
		return 0, e
	}

	vals := make([]interface{}, len(columns))
	maps := make([]interface{}, len(columns))
	elem := reflect.TypeOf(dest).Elem().Elem()

	if len(columns) == 1 {
		for rows.Next() {
			total++
			//dest must be a slice of pointer eg: []*string
			reflect.ValueOf(vals).Index(0).Set(reflect.New(elem.Elem()))
			e := rows.Scan(vals...)
			if e != nil {
				return 0, e
			}
			//*dest = append(*dest, vals[0])
			reflect.ValueOf(dest).Elem().Set(reflect.Append(reflect.ValueOf(dest).Elem(), reflect.ValueOf(vals[0])))
		}

		return total, nil
	}

	for i := 0; i < elem.NumField(); i++ {
		for j, col := range columns {
			if col == elem.Field(i).Tag.Get("col") {
				maps[j] = i //column[j] map to elem.field[i] //field must be a pointer
				reflect.ValueOf(vals).Index(j).Set(reflect.New(elem.Field(i).Type.Elem()))
			}
		}
	}

	for rows.Next() {
		total++
		newVal := reflect.New(elem)
		e := rows.Scan(vals...)
		if e != nil {
			return 0, e
		}

		for i := 0; i < len(maps); i++ {
			if maps[i] != nil {
				newVal.Elem().Field(maps[i].(int)).Set(reflect.ValueOf(vals[i]))
			}
		}
		//*dest = append(*dest, *newVal(*Type struct))
		reflect.ValueOf(dest).Elem().Set(reflect.Append(reflect.ValueOf(dest).Elem(), newVal.Elem()))
	}

	return total, nil
}

func (tx *Tx) Execute(sql string, param ...interface{}) (int64, error) {
	if result, e := tx.executor.Exec(sql, param...); e != nil {
		return 0, e
	} else {
		return result.RowsAffected()
	}
}

func Scan(dest interface{}) {
	reflect.Append(reflect.ValueOf(dest))
}

type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Transaction interface {
	Commit() error
	Rollback() error
}

type UnTx struct {
}

func (*UnTx) Commit() error {
	return nil
}

func (*UnTx) Rollback() error {
	return nil
}
