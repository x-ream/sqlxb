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
package interceptor

// Interceptor 拦截器接口
// 用于基础设施观察（日志、监控）
// 不用于业务逻辑
type Interceptor interface {
	// Name 拦截器名称（用于注册/卸载）
	Name() string

	// BeforeBuild 在 Build() 之前执行
	// ⭐ 只传 Metadata，编译时强制只能设置元数据
	// ⭐ 无法修改查询逻辑（类型系统保证）
	// 返回 error 可以阻止 Build()
	BeforeBuild(meta *Metadata) error

	// AfterBuild 在 Build() 之后执行
	// 用于观察生成的 SQL（日志、监控、审计）
	// 返回 error 可以阻止后续执行
	AfterBuild(built interface{}) error
}
