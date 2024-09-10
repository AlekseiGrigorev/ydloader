package integrations

import (
	"strings"

	"github.com/AlekseiGrigorev/ydloader/internal/db"
)

type Token struct {
	Token string
}

func (model *Token) GetDefaultSql() string {
	var sql = []string{
		"SELECT",
		"token",
		"FROM smartis_stat.integrations",
		"WHERE id = ?",
		";",
	}
	return strings.Join(sql, " ")
}

func (model *Token) GetNewModel() db.RowModel {
	return new(Token)
}

func (model *Token) GetColumnPointers() []interface{} {
	columnPointers := make([]interface{}, 1)
	columnPointers[0] = &model.Token
	return columnPointers
}
