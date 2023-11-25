// Copyright 2020 io.xream.sqlxb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package sqlxb

const (
	X        = ""
	AGG      = ""
	SUB      = "SUB"
	AND      = "AND"
	OR       = "OR"
	AND_SUB  = AND
	OR_SUB   = OR
	EQ       = "="
	NE       = "<>"
	GT       = ">"
	LT       = "<"
	GTE      = ">="
	LTE      = "<="
	LIKE     = "LIKE"
	NOT_LIKE = "NOT LIKE"
	IN       = "IN"
	NIN      = "NOT IN"
	IS_NULL  = "IS NULL"
	NON_NULL = "IS NOT NULL"
)

type Op func() string

func Eq() string {
	return EQ
}
func Ne() string {
	return NE
}
func Gt() string {
	return GT
}
func Gte() string {
	return GTE
}
func Lt() string {
	return LT
}
func Lte() string {
	return LTE
}
func Like() string {
	return LIKE
}
func LikeLeft() string {
	return LIKE
}
func NotLike() string {
	return NOT_LIKE
}
func IsNull() string {
	return IS_NULL
}
func NonNull() string {
	return NON_NULL
}
