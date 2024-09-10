package ydirectlogins

import (
	"strings"

	"github.com/AlekseiGrigorev/ydloader/internal/db"
)

type IntegrationLogin struct {
	Id    int
	Login string
}

func (model *IntegrationLogin) GetDefaultSql() string {
	var sql = []string{
		"SELECT",
		"ydl.id, ydl.Login",
		"FROM smartis_stat.YDirect_Logins ydl",
		"JOIN smartis_stat.YDirect_integrations_logins ydil ON ydl.id = ydil.login_id",
		"WHERE ydil.integration_id = ?",
		";",
	}
	return strings.Join(sql, " ")
}

func (model *IntegrationLogin) GetNewModel() db.RowModel {
	return new(IntegrationLogin)
}

func (model *IntegrationLogin) GetColumnPointers() []interface{} {
	columnPointers := make([]interface{}, 2)
	columnPointers[0] = &model.Id
	columnPointers[1] = &model.Login
	return columnPointers
}

func (model *IntegrationLogin) ToType(items []db.RowModel) []*IntegrationLogin {
	tItems := []*IntegrationLogin{}
	for _, item := range items {
		tItems = append(tItems, item.(*IntegrationLogin))
	}
	return tItems
}
