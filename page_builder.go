/*
 * Copyright 2020 io.xream.sqlxb
 *
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package sqlxb

type PageBuilder struct {
	page uint
	rows uint
	last uint64
	isTotalRowsIgnored bool
}

func (pb * PageBuilder) Paged() *PageBuilder{
	return new(PageBuilder)
}

func (pb * PageBuilder) Page(page uint) *PageBuilder{
	pb.page = page
	return pb
}

func (pb * PageBuilder) Rows(rows uint) *PageBuilder{
	pb.rows = rows
	return pb
}

/**
 * ASC: orderBy > last | DESC: orderBy < last
 * LIMIT rows
 */
func (pb * PageBuilder) Last(last uint64) *PageBuilder{
	pb.last = last
	return pb
}

func (pb * PageBuilder) IgnoreTotalRows(ignored bool) *PageBuilder{
	pb.isTotalRowsIgnored = ignored
	return pb
}
