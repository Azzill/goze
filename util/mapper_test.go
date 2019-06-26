/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package util

import (
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	a := struct {
		Str    string
		Int    int
		Bool   bool
		time   time.Time
		StrArr []string
	}{Bool: true, Int: 1, Str: "Text", StrArr: []string{"T1", "T2"}, time: time.Now()}

	b := struct {
		Str    string
		Int    int
		Bool   bool
		time   time.Time
		StrArr []string
		Extra  int
	}{}

	e := Map(&a, &b)

	if e != nil {
		t.Error(e)
	}

	if a.time == b.time || len(a.StrArr) != len(b.StrArr) || a.Int != b.Int {
		t.Error("Mapped not correspond the origin")
	}

}

func TestMapAll(t *testing.T) {
	type b struct {
		Str string
	}

	a := make([]struct {
		Str string
		Int int
	}, 5)

	for i := range a {
		a[i].Str = "Str"
		a[i].Int = 1
	}

	mapped := make([]b, len(a))
	e := MapAll(a, mapped)

	if e != nil {
		t.Error(e)
	}

	for _, m := range mapped {
		if m.Str != "Str" {
			t.Error("Mapped not correspond the origin")
		}
	}

}
