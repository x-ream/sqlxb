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

type SourceBuilder struct {
	po   Po
	alia string
	join *Join
	sub  *BuilderX
	bbs  []Bb
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
	join string
	on   *On
}
type On struct {
	key        string
	targetAlia string
	targetKey  string
}
type Using struct {
	key string
}

func ON(key string, targetOrAlia string, targetKey string) *On {
	if key == "" || targetOrAlia == "" || targetKey == "" {
		panic("On.key, On.targetOrAlia, On.targetKey can not blank")
	}
	return &On{
		key,
		targetOrAlia,
		targetKey,
	}
}

func USING(key string) *Using {
	if key == "" {
		panic("Using.key can not blank")
	}
	return &Using{
		key: key,
	}
}

func (sb *SourceBuilder) JoinOn(join JOIN, on *On) *SourceBuilder {
	if join == nil || on == nil {
		panic("join, on can not nil")
	}
	if sb.join != nil {
		panic("call Join repeated")
	}
	sb.join = &Join{
		join: join(),
		on:   on,
	}
	return sb
}

func (sb *SourceBuilder) JoinUsing(join JOIN, using *Using) *SourceBuilder {
	if join == nil || using == nil {
		panic("join, using can not nil")
	}
	if sb.join != nil {
		panic("call Join repeated")
	}
	sb.join = &Join{
		join: join(),
		on: &On{
			key: using.key,
		},
	}
	return sb
}

func (sb *SourceBuilder) More(cb *ConditionBuilder) {
	sb.bbs = cb.bbs
}

func (sb *SourceBuilder) Sub(sub *BuilderX) *SourceBuilder {
	sb.sub = sub
	return sb
}
