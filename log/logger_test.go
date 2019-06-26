/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package log

import (
	"testing"
)

func TestLogger_Error(t *testing.T) {
	logger := NewLogger("Test")
	logger.Error("Error message")
}

func TestLogger_Warn(t *testing.T) {

	logger := NewLogger("Test")
	logger.Warn("Warn message")
}

func TestLogger_Info(t *testing.T) {

	logger := NewLogger("Test")
	logger.Info("Info message")
}
