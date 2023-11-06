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

type FromX struct {
	po   Po
	alia string
	join *Join
	sub  *BuilderX
	s    string
}

type FromBuilder struct {
	x  *FromX
	xs *[]*FromX
}

func (sb *FromBuilder) Of(po Po) *FromBuilder {
	sb.x.po = po
	return sb
}

func (sb *FromBuilder) Alia(alia string) *FromBuilder {
	sb.x.alia = alia
	return sb
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

func (sb *FromBuilder) Cond(on func(on *ON)) *FromBuilder {
	if sb.x.join == nil || sb.x.join.on == nil {
		panic("call Cond(on *ON) after ON(onStr)")
	}
	on(sb.x.join.on)
	return sb
}

func (sb *FromBuilder) On(onStr string) *FromBuilder {
	sb.x.join.on = &ON{}
	sb.x.join.on.X(onStr)
	return sb
}

func (sb *FromBuilder) Using(key string) *FromBuilder {
	if key == "" {
		panic("USING.key can not blank")
	}
	sb.x.join.on = &ON{}
	sb.x.join.on.orUsingKey = key
	return sb
}

func (sb *FromBuilder) JOIN(join JOIN) *FromBuilder {
	if join == nil {
		panic("join, on can not nil")
	}

	x := FromX{}
	*sb.xs = append(*sb.xs, &x)
	sb.x = &x

	sb.x.join = &Join{
		join: join(),
	}
	return sb
}

func (sb *FromBuilder) Sub(sub func(sub *BuilderX)) *FromBuilder {
	x := new(BuilderX)
	sb.x.sub = x
	sub(x)
	return sb
}
