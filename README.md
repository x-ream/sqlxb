# sqlxb  
[![OSCS Status](https://www.oscs1024.com/platform/badge/x-ream/sqlxb.svg?size=small)](https://www.oscs1024.com/project/x-ream/sqlxb?ref=badge_small)
![workflow build](https://github.com/x-ream/sqlxb/actions/workflows/go.yml/badge.svg)
[![GitHub tag](https://img.shields.io/github/tag/x-ream/sqlxb.svg?style=flat)](https://github.com/x-ream/sqlxb/tags)
[![Go Report Card](https://goreportcard.com/badge/github.com/x-ream/sqlxb)](https://goreportcard.com/report/github.com/x-ream/sqlxb)

a tool of sql query builder, build sql for sql.DB, [sqlx](https://github.com/jmoiron/sqlx/), [gorp](https://github.com/go-gorp/gorp),
or build condition sql for some orm framework, like [gorm](https://github.com/go-gorm/gorm)....

## Example

    SELECT * FROM t_cat WHERE id > ? AND (price >= ? OR is_sold = ?)

    var Db *sqlx.DB
    ....

    c := Cat{}
	builder := sqlxb.Of(&c).Gt("id", 10000).And(func(sub *CondBuilder) {
		sub.Gte("price", catRo.Price).OR().Eq("is_sold", catRo.IsSold))
    })

    vs, dataSql, countSql, _ := builder.Build().Sql()
    catList := []Cat{}
	err = Db.Select(&catList, dataSql, vs...)


## Contributing

Contributors are welcomed to join the sqlxb project. <br>
Please check [CONTRIBUTING](./CONTRIBUTING.md)

## Quickstart

* [Single Example](#single-example)
* [Join Example](#join-example)


### Single Example

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
	var builder = Of(&c)
	builder.LikeLeft("name",catRo.Name)
	builder.X("weight <> ?", 0) //X(k, v...), hardcode func, value 0 and nil will NOT ignore
    //Eq,Ne,Gt.... value 0 and nil will ignore, like as follow: OR().Eq("is_sold", catRo.IsSold)
	builder.And(func(sub *CondBuilder) {
            sub.Gte("price", catRo.Price).OR().Gte("age", catRo.Age).OR().Eq("is_sold", catRo.IsSold))
	    })
    //func Bool NOT designed for value nil or 0; designed to convert complex logic to bool
    //Decorator pattern suggest to use func Bool preCondition, like:
    //myBoolDecorator := NewMyBoolDecorator(para)
    //builder.Bool(myBoolDecorator.fooCondition, func(cb *CondBuilder) {
	builder.Bool(preCondition, func(cb *CondBuilder) {
            cb.Or(func(sub *CondBuilder) {
                sub.Lt("price", 5000)
            })
	})
	builder.Sort("id", ASC)
        builder.Paged(func(pb *PageBuilder) {
                pb.Page(1).Rows(10).IgnoreTotalRows()
            })
	vs, dataSql, countSql, _ := builder.Build().Sql()
    // ....

    //dataSql: SELECT * FROM t_cat WHERE id > ? AND name LIKE ? AND weight <> 0 AND (price >= ? OR age >= ?) OR (price < ?)
    //ORDER BY id ASC LIMIT 10

	//.IgnoreTotalRows(), will not output countSql
    //countSql: SELECT COUNT(*) FROM t_cat WHERE name LIKE ? AND weight <> 0 AND (price >= ? OR age >= ?) OR (price < ?)
    
    //sqlx: 	err = Db.Select(&catList, dataSql,vs...)
	joinSql, condSql, cvs := builder.Build().SqlOfCond()
    
    //conditionSql: id > ? AND name LIKE ? AND weight <> 0 AND (price >= ? OR age >= ?) OR (price < ?)

}
```


### Join Example

```Go
import (
        . "github.com/x-ream/sqlxb"
    )
    
func main() {
	
	sub := func(sub *BuilderX) {
                sub.Select("id","type").From("t_pet").Gt("id", 10000) //....
            }
	
        builder := Of(nil).
		Select("p.id").
		FromX(func(sb *FromBuilder) {
                    sb.
                        Sub(sub).Alia("p").
                        JOIN(INNER).Of("t_dog").Alia("d").On("d.pet_id = p.id").
                        JOIN(LEFT).Of("t_cat").Alia("c").On("c.pet_id = p.id").
                            Cond(func(on *ON) {
                                on.Gt("c.id", ro.MinCatId)
                            })
		    }).
	        Ne("p.type","PIG")
    
}


```