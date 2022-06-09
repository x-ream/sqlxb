# sqlxb  
![workflow build](https://github.com/x-ream/sqlxb/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/x-ream/sqlxb)](https://goreportcard.com/report/github.com/x-ream/sqlxb)

a tool of sql golang builder, build sql for sql.DB, sqlx, 
or build condition sql for some orm framework

## API
    Builder: build sql like: SELECT * FROM t_foo WHERE name Like "%xx%" ORDER BY ID DESC
    BuilderX: build sql like: SELECT DISTINCT(f.id) FROM t_foo f INNER JOIN t_bar b ON ....
    
    builder.Gte("id", 10000)
    builder.And(SubCondition().Gte("price", catRo.Price).OR().Eq("is_sold", catRo.IsSold))

## Builder DEMO

```Go

import (
    . "github.com/x-ream/sqlxb"
)

type Cat struct {
	Id       uint64    `db:"id"`
	Name     string    `db:"name"`
	Age      uint      `db:"age"`
	Color    string    `db:"color"`
	Weight   float64   `db:"weight"`
	IsSold   *bool     `db:"is_sold"`
	Price    *float64  `db:"price"`
	CreateAt time.Time `db:"create_at"`
}

func (*Cat) TableName() string {
	return "t_cat"
}

// IsSold, Price, fields can be zero, must be pointer, like Java Boolean....
// sqlxb has func: Bool(true), Int(v) ....
// sqlxb no relect, not support omitempty, should rewrite ro, dto
type CatRo struct {
	Name   string   `json:"name, string"`
	IsSold *bool    `json:"isSold, *bool"`
	Price  *float64 `json:"price, *float64"`
	Age    uint     `json:"age", unit`
}

func main() {
	cat := Cat{
		Id:       100002,
		Name:     "Tuanzi",
		Age:      1,
		Color:    "B",
		Weight:   8.5,
		IsSold:   Bool(true),
		Price:    Float64(10000.00),
		CreateAt: time.Now(),
	}
    // INSERT .....

    // PREPARE TO QUERY
	catRo := CatRo{
		Name:	"Tu",
		IsSold: nil,
		Price:  Float64(5000.00),
		Age:    1,
	}

	preCondition := func() bool {
		if cat.Color == "W" {
			return true
		} else if cat.Weight <= 3 {
			return false
		} else {
			return true
		}
	}

	c := Cat{}
	var builder = NewBuilder(&c)
	builder.LikeRight("name",catRo.Name)
	builder.And(SubCondition().Gte("price", catRo.Price).OR().Gte("age", catRo.Age).OR().Eq("is_sold", catRo.IsSold))
	builder.Bool(preCondition, func(cb *ConditionBuilder) {
		cb.Or(SubCondition().Lt("price", 5000))
	})
	builder.Sort("id", ASC)
	builder.Paged().Rows(10).Last(100)
	vs, dataSql, countSql := builder.Build().Sql()
    // ....

    //dataSql: SELECT * FROM t_cat WHERE id > ? AND name LIKE ? AND (price >= ? OR age >= ?) OR (price < ?)
    //ORDER BY id ASC LIMIT 10

    //countSql: SELECT COUNT(*) FROM t_cat WHERE name LIKE ? AND (price >= ? OR age >= ?) OR (price < ?)
    
    //sqlx: 	rows, _ := Db.Query(*dataSql,(*vs)...)
	_, conditionSql := builder.Build().SqlOfCondition()
    
    //conditionSql: name LIKE ? AND (price >= ? OR age >= ?) OR (price < ?)

}
```