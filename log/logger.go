/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Level int

const (
	_ Level = iota
	Info
	Warn
	Error
)

type Logger struct {
	name string
}

type loggerConfig struct {
	EnableColor bool
	Writer      io.Writer
}

var LoggerConfig = loggerConfig{EnableColor: true, Writer: os.Stdout}

func NewLogger(name string) *Logger {
	return &Logger{name: name}
}

func (l *Logger) Info(message ...interface{}) {
	l.outputMessage(Info, message)
}

func (l *Logger) Error(message ...interface{}) {
	l.outputMessage(Error, message)
}

func (l *Logger) Warn(message ...interface{}) {
	l.outputMessage(Warn, message)
}

func (l *Logger) outputMessage(level Level, message ...interface{}) {
	var levelColorId int
	var levelPrefix string
	switch level {
	case Info:
		levelColorId = 36
		levelPrefix = "INFO"
	case Warn:
		levelColorId = 33
		levelPrefix = "WARN"
	case Error:
		levelColorId = 31
		levelPrefix = "ERROR"
	}
	colorChange := fmt.Sprintf("\x1b[0;%dm", levelColorId)
	colorReset := "\x1b[0m"

	var s string
	if LoggerConfig.EnableColor {
		s = fmt.Sprintln(time.Now().Format(time.RFC3339), colorChange, levelPrefix, colorReset, l.name, message)
	} else {
		s = fmt.Sprintln(time.Now().Format(time.RFC3339), levelPrefix, l.name, message)

	}
	_, _ = LoggerConfig.Writer.Write([]byte(s))
}
