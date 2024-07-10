package sqlxb

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type Cat struct {
	Id int64
	M  map[string]string
}

type Pet struct {
	One  Cat
	Id   int64
	Body []byte
}

func (*Pet) TableName() string {
	return "t_pet"
}

func TestInsert(t *testing.T) {

	t.Run("insert", func(t *testing.T) {

		str := "eHh4eHh4eHh4eHg="
		buffer, _ := json.Marshal(str)
		fmt.Println("-------------")
		fmt.Println(string(buffer))
		fmt.Println("+++++++++++++")

		mm := make(map[string]string)
		mm["xxxx"] = "zzzzz"
		body := []byte("xxxxxxxxxxx")
		fmt.Println(string(body))

		var po Pet
		sql, vs := Of(&po).
			Insert(func(b *InsertBuilder) {
				b.Set("id", 1).
					Set("one", Cat{
						Id: 2,
						M:  mm,
					}).
					Set("body", body)
			}).
			Build().
			SqlOfInsert()

		if !strings.Contains(sql, "one") {
			t.Error("sql erro")
		}

		fmt.Println(vs)
		fmt.Println(sql)
	})
}

func TestUpdate(t *testing.T) {

	t.Run("update", func(t *testing.T) {
		mm := make(map[string]string)
		mm["xxxx"] = "zzzzz"

		var po Pet
		sql, vs := Of(&po).
			Update(func(b *UpdateBuilder) {

				b.Set("one", Cat{
					Id: 2,
					M:  mm,
				})
			}).
			Eq("id", 2).
			Build().
			SqlOfUpdate()

		if !strings.Contains(sql, "one") {
			t.Error("sql erro")
		}

		fmt.Println(vs)
		fmt.Println(sql)

	})

}

func TestDelete(t *testing.T) {

	t.Run("delete", func(t *testing.T) {

		var po Pet
		sql, vs := Of(&po).Eq("id", 2).
			Any(func(x *BuilderX) {
				if po.Id != 4 {
					x.Gt("id", 1)
				}
			}).
			Build().
			SqlOfDelete()

		if !strings.Contains(sql, "t_pet") {
			t.Error("sql erro")
		}

		fmt.Println(vs)
		fmt.Println(sql)
	})

}
