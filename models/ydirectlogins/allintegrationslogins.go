package ydirectlogins

import (
	"strings"

	"github.com/AlekseiGrigorev/ydloader/internal/db"
)

type AllIntegrationsLogin struct {
	Id            int
	Login         string
	IntegrationId int
	Token         string
}

func (model *AllIntegrationsLogin) GetDefaultSql() string {
	var sql = []string{
		"SELECT",
		"ydl.id AS id, ydl.Login AS login, i.id AS integration_id, i.token AS token",
		"FROM smartis_stat.YDirect_Logins ydl",
		"JOIN smartis_stat.YDirect_integrations_logins ydil ON ydl.id = ydil.login_id",
		"JOIN smartis_stat.integrations i ON ydil.integration_id = i.id",
		"WHERE i.isActive = 1 AND i.isDeleted = 0",
		";",
	}
	return strings.Join(sql, " ")
}

func (model *AllIntegrationsLogin) GetNewModel() db.RowModel {
	return new(AllIntegrationsLogin)
}

func (model *AllIntegrationsLogin) GetColumnPointers() []interface{} {
	columnPointers := make([]interface{}, 4)
	columnPointers[0] = &model.Id
	columnPointers[1] = &model.Login
	columnPointers[2] = &model.IntegrationId
	columnPointers[3] = &model.Token
	return columnPointers
}

func (model *AllIntegrationsLogin) ToType(items []db.RowModel) []*AllIntegrationsLogin {
	tItems := []*AllIntegrationsLogin{}
	for _, item := range items {
		tItems = append(tItems, item.(*AllIntegrationsLogin))
	}
	return tItems
}
