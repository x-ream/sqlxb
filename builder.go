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
//
package sqlxb

//
// TO build sql, like: SELECT * FROM ....
// Can add L2Cache
//
// @author Sim
//
type Builder struct {
	ConditionBuilder
	pageBuilder *PageBuilder

	sorts    []*Sort
	//havings  []*Bb
	//groupBys []string

	po Po
}

func NewBuilder(poOrNil Po) *Builder {
	var instance = newBuilder()
	instance.po = poOrNil
	return instance
}

func newBuilder() *Builder {
	b := new(Builder)
	b.bbs = &[]*Bb{}
	return b
}

func (builder *Builder) Sort(orderBy string, direction Direction) *Builder {
	if orderBy == "" || direction == nil {
		return builder
	}
	sort := Sort{orderBy: orderBy, direction: direction()}
	builder.sorts = append(builder.sorts, &sort)
	return builder
}

func (builder *Builder) Paged() *PageBuilder {
	pageBuilder := new(PageBuilder)
	builder.pageBuilder = pageBuilder
	return pageBuilder
}

func (builder *Builder) Build() *Built {
	if builder == nil {
		panic("sqlxb.Builder is nil")
	}
	built := Built{
		ResultKeys: nil,
		ConditionX: builder.bbs,
		Sorts:      &builder.sorts,

		Po: builder.po,
	}
	if builder.pageBuilder != nil {
		built.PageCondition = &builder.pageBuilder.condition
	}

	return &built
}
