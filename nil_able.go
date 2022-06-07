// Copyright 2020 io.xream.sqlxb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package sqlxb

import "strconv"

func Bool(b bool) *bool {
	return &b
}
func Int(v int) *int {
	return &v
}
func Int64(v int64) *int64 {
	return &v
}
func Int32(v int32) *int32 {
	return &v
}
func Int16(v int16) *int16 {
	return &v
}
func Int8(v int8) *int8 {
	return &v
}
func Byte(b byte) *byte {
	return &b
}
func Float64(f float64) *float64 {
	return &f
}
func Float32(f float32) *float32 {
	return &f
}
func Uint64(v uint64) *uint64 {
	return &v
}
func Uint(v uint) *uint {
	return &v
}

func Np2s(p interface{}) (string, bool) {
	switch p.(type) {
	case *uint64:
		if np := p.(*uint64); np != nil {
			return strconv.FormatUint(*np, 10), true
		}
	case *uint:
		if np := p.(*uint); np != nil {
			return strconv.FormatUint(uint64(*np), 10), true
		}
	case *int64:
		if np := p.(*int64); np != nil {
			return strconv.FormatInt(*np, 10), true
		}
	case *int:
		if np := p.(*int); np != nil {
			return strconv.Itoa(*np), true
		}
	case *int32:
		if np := p.(*int32); np != nil {
			return strconv.Itoa(int(*np)), true
		}
	case *int16:
		if np := p.(*int16); np != nil {
			return strconv.Itoa(int(*np)), true
		}
	case *int8:
		if np := p.(*int8); np != nil {
			return strconv.Itoa(int(*np)), true
		}
	case *byte:
		if np := p.(*byte); np != nil {
			return strconv.Itoa(int(*np)), true
		}
	case *float64:
		if np := p.(*float64); np != nil {
			return strconv.FormatFloat(*np, 'f', -1, 64), true
		}
	case *float32:
		if np := p.(*float32); np != nil {
			return strconv.FormatFloat(float64(*np), 'f', -1, 32), true
		}
	}

	return "", false
}

func N2s(p interface{}) string {
	switch p.(type) {
	case uint64:
		return strconv.FormatUint(p.(uint64), 10)
	case uint:
		return strconv.FormatUint(uint64(p.(uint)), 10)
	case int64:
		return strconv.FormatInt(p.(int64), 10)
	case int:
		return strconv.Itoa(p.(int))
	case int32:
		return strconv.Itoa(int(p.(int32)))
	case int16:
		return strconv.Itoa(int(p.(int16)))
	case int8:
		return strconv.Itoa(int(p.(int8)))
	case byte:
		return strconv.Itoa(int(p.(byte)))
	case float64:
		return strconv.FormatFloat(p.(float64), 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(p.(float32)), 'f', -1, 32)
	}

	return ""
}

func IsNil(p interface{}) bool {
	switch p.(type) {
	case *uint64:
		return p.(*uint64) == nil
	case *uint:
		return p.(*uint) == nil
	case *int64:
		return p.(*int64) == nil
	case *int:
		return p.(*int) == nil
	case *int32:
		return p.(*int32) == nil
	case *int16:
		return p.(*int16) == nil
	case *int8:
		return p.(*int8) == nil
	case *byte:
		return p.(*byte) == nil
	case *float64:
		return p.(*float64) == nil
	case *float32:
		return p.(*float32) == nil
	case *bool:
		return p.(*bool) == nil
	}
	return false
}
