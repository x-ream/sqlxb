package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/x-ream/sqlxb"
)

// CodeRepository 代码仓库
type CodeRepository struct {
	db *sqlx.DB
}

func NewCodeRepository(db *sqlx.DB) *CodeRepository {
	return &CodeRepository{db: db}
}

// Create 创建代码片段
func (r *CodeRepository) Create(code *CodeSnippet) error {
	sql, args := sqlxb.Of(&CodeSnippet{}).
		Insert(func(ib *sqlxb.InsertBuilder) {
			ib.Set("file_path", code.FilePath).
				Set("language", code.Language).
				Set("content", code.Content).
				Set("embedding", code.Embedding)
		}).
		Build().
		SqlOfInsert()

	result, err := r.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	code.ID = id
	return nil
}

// GetByID 根据 ID 获取
func (r *CodeRepository) GetByID(id int64) (*CodeSnippet, error) {
	sql, args, _ := sqlxb.Of(&CodeSnippet{}).
		Eq("id", id).
		Build().
		SqlOfSelect()

	var code CodeSnippet
	err := r.db.Get(&code, sql, args...)
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// VectorSearch 向量搜索
func (r *CodeRepository) VectorSearch(queryVector []float32, limit int) ([]*CodeSnippet, error) {
	if limit <= 0 {
		limit = 10
	}

	sql, args, _ := sqlxb.Of(&CodeSnippet{}).
		VectorSearch("embedding", queryVector, limit).
		Build().
		SqlOfVectorSearch()

	var codes []*CodeSnippet
	err := r.db.Select(&codes, sql, args...)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// HybridSearch 混合搜索（向量 + 标量过滤）
func (r *CodeRepository) HybridSearch(queryVector []float32, language string, limit int) ([]*CodeSnippet, error) {
	if limit <= 0 {
		limit = 10
	}

	sql, args, _ := sqlxb.Of(&CodeSnippet{}).
		VectorSearch("embedding", queryVector, limit).
		Eq("language", language). // 自动过滤空字符串
		Build().
		SqlOfVectorSearch()

	var codes []*CodeSnippet
	err := r.db.Select(&codes, sql, args...)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// KeywordSearch 关键词搜索
func (r *CodeRepository) KeywordSearch(keyword, language string, page, rows int) ([]*CodeSnippet, int64, error) {
	builder := sqlxb.Of(&CodeSnippet{}).
		Like("content", keyword).
		Eq("language", language).
		Paged(func(pb *sqlxb.PageBuilder) {
			pb.Page(int64(page)).Rows(int64(rows))
		})

	countSql, dataSql, args, _ := builder.Build().SqlOfPage()

	// 获取总数
	var total int64
	if countSql != "" {
		r.db.Get(&total, countSql)
	}

	// 获取数据
	var codes []*CodeSnippet
	err := r.db.Select(&codes, dataSql, args...)
	if err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}
