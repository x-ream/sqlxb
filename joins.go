// Copyright 2025 me.fndo.xb
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
package xb

const (
	inner_join      = "INNER JOIN"
	left_join       = "LEFT JOIN"
	right_join      = "RIGHT JOIN"
	cross_join      = "CROSS JOIN"
	asof_join       = "ASOF JOIN"
	global_join     = "GLOBAL JOIN"
	full_outer_join = "FULL OUTER JOIN"
)

/**
 * Config your own JOIN string as string func
 */
type JOIN func() string

func NON_JOIN() string {
	return ", "
}

func INNER() string {
	return inner_join
}

func LEFT() string {
	return left_join
}

func RIGHT() string {
	return right_join
}

func CROSS() string {
	return cross_join
}

func ASOF() string {
	return asof_join
}

func GLOBAL() string {
	return global_join
}

func FULL_OUTER() string {
	return full_outer_join
}
