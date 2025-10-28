package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fndome/xb"
)

// QdrantClient Qdrant 客户端
type QdrantClient struct {
	baseURL    string
	collection string
	httpClient *http.Client
}

func NewQdrantClient(baseURL, collection string) *QdrantClient {
	return &QdrantClient{
		baseURL:    baseURL,
		collection: collection,
		httpClient: &http.Client{},
	}
}

// Search 向量搜索
func (c *QdrantClient) Search(queryVector []float32, docType, language string, minScore float64, limit int) ([]*Document, error) {
	// 使用 xb 构建查询
	built := xb.Of(&Document{}).
		VectorSearch("embedding", queryVector, limit).
		Eq("doc_type", docType).
		Eq("language", language).
		Gt("score", minScore).
		QdrantX(func(qx *xb.QdrantBuilderX) {
			qx.ScoreThreshold(float32(minScore)).
				HnswEf(128)
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		return nil, err
	}

	// 发送请求到 Qdrant
	url := fmt.Sprintf("%s/collections/%s/points/search", c.baseURL, c.collection)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 解析结果（简化版）
	var result struct {
		Result []struct {
			ID      int64                  `json:"id"`
			Score   float64                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 转换为 Document
	docs := make([]*Document, 0, len(result.Result))
	for _, item := range result.Result {
		doc := &Document{
			ID:       item.ID,
			Title:    item.Payload["title"].(string),
			Content:  item.Payload["content"].(string),
			DocType:  item.Payload["doc_type"].(string),
			Language: item.Payload["language"].(string),
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// Recommend 推荐查询
func (c *QdrantClient) Recommend(positive, negative []int64, limit int) ([]*Document, error) {
	built := xb.Of(&Document{}).
		QdrantX(func(qx *xb.QdrantBuilderX) {
			qx.Recommend(func(rb *xb.RecommendBuilder) {
				rb.Positive(positive...).
					Negative(negative...).
					Limit(limit)
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantRecommendJSON()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/recommend", c.baseURL, c.collection)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader([]byte(jsonStr)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Result []struct {
			ID      int64                  `json:"id"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	docs := make([]*Document, 0, len(result.Result))
	for _, item := range result.Result {
		doc := &Document{
			ID:      item.ID,
			Title:   item.Payload["title"].(string),
			Content: item.Payload["content"].(string),
		}
		docs = append(docs, doc)
	}

	return docs, nil
}
