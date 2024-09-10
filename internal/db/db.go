// Copyright 2024 Aleksei Grigorev
// https://aleksvgrig.com, https://github.com/AlekseiGrigorev, aleksvgrig@gmail.com.
// Package define interfaces, structures and functions for working with mysql database
package db

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/AlekseiGrigorev/ydloader/internal/trace"
	"github.com/go-sql-driver/mysql"
)

// RowModel define interface for model, created from on row of selected data
type RowModel interface {
	GetNewModel() RowModel
	GetColumnPointers() []interface{}
}

// Db define database config and connection
type Db struct {
	config mysql.Config
	db     *sql.DB
}

// Init database connection
func (dbIn *Db) Init(user string, passwd string, addr string, port int, dbName string) {
	if port != 0 {
		addr += ":" + strconv.Itoa(port)
	}

	dbIn.config = mysql.Config{
		User:                 user,
		Passwd:               passwd,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               dbName,
		AllowNativePasswords: true,
	}
}

// Open database connection
func (dbIn *Db) open() error {
	var dsn = dbIn.config.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		db = nil
		fmt.Println(err, trace.GetTrace())
		return err
	}
	dbIn.db = db
	return nil
}

// Connect to database. Open connection if needed
func (dbIn *Db) connect() error {
	if dbIn.db == nil {
		err := dbIn.open()
		if err != nil {
			fmt.Println(err, trace.GetTrace())
			return err
		}
	}
	err := dbIn.db.Ping()
	if err != nil {
		fmt.Println(err, trace.GetTrace())
		return err
	}
	return nil
}

// QueryRow query one row data from database
// Returns RowModel
func (dbIn *Db) QueryRow(sql string, params []any, model RowModel) (RowModel, error) {
	err := dbIn.connect()
	if err != nil {
		fmt.Println(err, trace.GetTrace())
		return nil, err
	}
	newModel := model.GetNewModel()
	err = dbIn.db.QueryRow(sql, params...).Scan(newModel.GetColumnPointers()...)
	if err != nil {
		fmt.Println(err, trace.GetTrace())
		return nil, err
	}

	return newModel, nil
}

// Query query rows data from database
// Returns slice of RowModels
func (dbIn *Db) Query(sql string, params []any, model RowModel) ([]RowModel, error) {
	err := dbIn.connect()
	if err != nil {
		fmt.Println(err, trace.GetTrace())
		return nil, err
	}
	rows, err := dbIn.db.Query(sql, params...)
	if err != nil {
		fmt.Println(err, trace.GetTrace())
		return nil, err
	}
	defer rows.Close()
	retRows := make([]RowModel, 0)
	for rows.Next() {
		newModel := model.GetNewModel()
		if err := rows.Scan(newModel.GetColumnPointers()...); err != nil {
			fmt.Println(err, trace.GetTrace())
			return nil, err
		}
		retRows = append(retRows, newModel)
	}

	return retRows, nil
}
