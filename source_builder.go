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

type SourceBuilder struct {
	po   Po
	alia string
	join *Join
	sub  *BuilderX
	s    string
}

func (sb *SourceBuilder) Source(po Po) *SourceBuilder {
	sb.po = po
	return sb
}

func (sb *SourceBuilder) Alia(alia string) *SourceBuilder {
	sb.alia = alia
	return sb
}

type Join struct {
	join   string
	target string
	alia   string
	on     *On
}
type On struct {
	ConditionBuilder
	orUsingKey string
}
type Using struct {
	key string
}

func (join *Join) ON(k string) *On {
	join.on = &On{}
	join.on.X(k)
	return join.on
}

func (join *Join) USING(key string) {
	if key == "" {
		panic("Using.key can not blank")
	}
	join.on = &On{}
	join.on.orUsingKey = key
}

func (join *Join) Alia(alia string) *Join {
	join.alia = alia
	return join
}

func (sb *SourceBuilder) Join(join JOIN, po Po) *Join {
	if join == nil {
		panic("join, on can not nil")
	}
	if sb.join != nil {
		panic("call Join repeated")
	}
	sb.join = &Join{
		join:   join(),
		target: po.TableName(),
	}
	return sb.join
}

func (sb *SourceBuilder) JoinScript(joinScript string) {
	sb.s = joinScript
}

func (sb *SourceBuilder) Sub(sub *BuilderX) *SourceBuilder {
	sb.sub = sub
	return sb
}
