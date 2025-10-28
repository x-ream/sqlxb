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
package xb

type FromX struct {
	tableName string
	alia      string
	join      *Join
	sub       *BuilderX
}

type FromBuilder struct {
	x  *FromX
	xs *[]*FromX
}

func (fb *FromBuilder) Of(tableName string) *FromBuilder {
	fb.x.tableName = tableName
	return fb
}

func (fb *FromBuilder) As(alia string) *FromBuilder {
	fb.x.alia = alia
	return fb
}

type Join struct {
	join string
	on   *ON
}

type ON struct {
	CondBuilder
	orUsingKey string
}

type USING struct {
	key string
}

func (fb *FromBuilder) Cond(on func(on *ON)) *FromBuilder {
	if fb.x.join == nil || fb.x.join.on == nil {
		panic("call Cond(on *ON) after ON(onStr)")
	}
	on(fb.x.join.on)
	return fb
}

func (fb *FromBuilder) On(onStr string) *FromBuilder {
	fb.x.join.on = &ON{}
	fb.x.join.on.X(onStr)
	return fb
}

func (fb *FromBuilder) Using(key string) *FromBuilder {
	if key == "" {
		panic("USING.key can not blank")
	}
	fb.x.join.on = &ON{}
	fb.x.join.on.orUsingKey = key
	return fb
}

func (fb *FromBuilder) JOIN(join JOIN) *FromBuilder {
	if join == nil {
		panic("join, on can not nil")
	}

	x := FromX{}
	*fb.xs = append(*fb.xs, &x)
	fb.x = &x

	fb.x.join = &Join{
		join: join(),
	}
	return fb
}

func (fb *FromBuilder) Sub(sub func(sb *BuilderX)) *FromBuilder {
	x := new(BuilderX)
	fb.x.sub = x
	sub(x)

	fb.x.tableName = ""

	return fb
}
