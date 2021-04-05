/**
 * Copyright 2017 ~ 2025 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package xcloud_dopaas_coredns

import (
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin(PluginName)

// Debugf print debug format log
func Debugf(format string, v ...interface{}) {
	log.Debugf(concatLogPrefix(format), v...)
}

// Infof print info format log
func Infof(format string, v ...interface{}) {
	log.Infof(concatLogPrefix(format), v...)
}

// Warningf print warn format log
func Warningf(format string, v ...interface{}) {
	log.Warningf(concatLogPrefix(format), v...)
}

// Errorf print error format log
func Errorf(format string, v ...interface{}) {
	log.Errorf(concatLogPrefix(format), v...)
}

// Error print error log
func Error(v ...interface{}) {
	arr := make([]interface{}, 0, 10)
	arr = append(arr, getLogPrefix())
	v = append(arr, v)
	log.Error(v...)
}

func concatLogPrefix(str string) string {
	now := time.Now()
	return time2Str(now) + " " + str
}

func getLogPrefix() string {
	now := time.Now()
	return time2Str(now) + " "
}

func time2Str(t time.Time) string {
	const shortForm = "[2006-01-01 15:04:05]"
	temp := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	str := temp.Format(shortForm)
	return str
}
