/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package util

import (
	"errors"
	"reflect"
)

//This is a utility for mapping one struct to another

func Map(src interface{}, dst interface{}) error {
	return mapValue(reflect.ValueOf(src).Elem(), reflect.ValueOf(dst).Elem())
}

func MapAll(srcSlice interface{}, dstSlice interface{}) error {

	dstArrayValue := reflect.ValueOf(dstSlice)
	srcSliceValue := reflect.ValueOf(srcSlice)
	if dstArrayValue.Len() < srcSliceValue.Len() {
		return errors.New("len of dest slice must be the same as src slice")
	}
	for i := 0; i < srcSliceValue.Len(); i++ {
		src := srcSliceValue.Index(i)
		e := mapValue(src, dstArrayValue.Index(i))
		if e != nil {
			return e
		}
	}
	return nil
}

func mapValue(src reflect.Value, dst reflect.Value) error {
	typeSrc := src.Type()
	typeDst := dst.Type()

	for i := 0; i < typeSrc.NumField(); i++ {
		field, has := typeDst.FieldByName(typeSrc.Field(i).Name)
		if !has || field.Anonymous || !dst.Field(field.Index[0]).CanSet() {
			continue
		}
		dst.Field(field.Index[0]).Set(src.Field(i))
	}
	return nil
}
