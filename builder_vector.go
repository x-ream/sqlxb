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

// VectorSearch vector similarity search (BuilderX extension)
// Same functionality as CondBuilder.VectorSearch(), but returns *BuilderX for chaining
//
// Example:
//
//	xb.Of(&CodeVector{}).
//	    Eq("language", "golang").
//	    VectorSearch("embedding", queryVector, 10).
//	    Build().
//	    SqlOfVectorSearch()
func (x *BuilderX) VectorSearch(field string, queryVector Vector, topK int) *BuilderX {
	x.CondBuilder.VectorSearch(field, queryVector, topK)
	return x
}

// VectorDistance sets vector distance metric (BuilderX extension)
//
// Example:
//
//	builder.VectorSearch("embedding", vec, 10).
//	    VectorDistance(xb.L2Distance)
func (x *BuilderX) VectorDistance(metric VectorDistance) *BuilderX {
	x.CondBuilder.VectorDistance(metric)
	return x
}

// VectorDistanceFilter vector distance filtering (BuilderX extension)
//
// Example:
//
//	builder.VectorDistanceFilter("embedding", queryVector, "<", 0.3)
func (x *BuilderX) VectorDistanceFilter(
	field string,
	queryVector Vector,
	op string,
	threshold float32,
) *BuilderX {
	x.CondBuilder.VectorDistanceFilter(field, queryVector, op, threshold)
	return x
}

// WithDiversity sets diversity parameters (BuilderX extension)
// ⭐ Core: if database doesn't support, will be automatically ignored
//
// Example:
//
//	xb.Of(&CodeVector{}).
//	    VectorSearch("embedding", vec, 20).
//	    WithDiversity(xb.DiversityByHash, "semantic_hash").
//	    Build()
func (x *BuilderX) WithDiversity(strategy DiversityStrategy, params ...interface{}) *BuilderX {
	x.CondBuilder.WithDiversity(strategy, params...)
	return x
}

// WithMinDistance sets minimum distance diversity (BuilderX extension)
//
// Example:
//
//	xb.Of(&CodeVector{}).
//	    VectorSearch("embedding", vec, 20).
//	    WithMinDistance(0.3).
//	    Build()
func (x *BuilderX) WithMinDistance(minDistance float32) *BuilderX {
	x.CondBuilder.WithMinDistance(minDistance)
	return x
}

// WithHashDiversity sets hash deduplication (BuilderX extension)
//
// Example:
//
//	xb.Of(&CodeVector{}).
//	    VectorSearch("embedding", vec, 20).
//	    WithHashDiversity("semantic_hash").
//	    Build()
func (x *BuilderX) WithHashDiversity(hashField string) *BuilderX {
	x.CondBuilder.WithHashDiversity(hashField)
	return x
}

// WithMMR sets MMR algorithm (BuilderX extension)
//
// Example:
//
//	xb.Of(&CodeVector{}).
//	    VectorSearch("embedding", vec, 20).
//	    WithMMR(0.5).
//	    Build()
func (x *BuilderX) WithMMR(lambda float32) *BuilderX {
	x.CondBuilder.WithMMR(lambda)
	return x
}
