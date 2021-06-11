package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Script struct {
	ID         uint32    `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Script     string    `json:"script,omitempty"`
	Type       string    `json:"type,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

const (
	SchemaName = "project"
	TableName  = "scripts"
)

const (
	postgresScriptCreateDatabase = iota
	postgresScriptCreateTable
	postgresScriptRegisterScript
	postgresScriptSelectAll
	postgresScriptSelectScriptByID
	postgresScriptUpdateScriptByID
	postgresScriptDeleteScriptByID
)

var scriptSQLString = map[int]string{
	postgresScriptCreateDatabase: fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, SchemaName),
	postgresScriptCreateTable: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		id serial PRIMARY KEY,
		name CHAR(50) NOT NULL,
		script text NOT NULL,
		type CHAR(10) NOT NULL,
		create_time timestamp DEFAULT timestamp '2000-01-01 00:00:00'
	);`, SchemaName, TableName),
	postgresScriptRegisterScript:   fmt.Sprintf(`INSERT INTO %s.%s (name, script, type, create_time) VALUES ($1, $2, $3, current_timestamp);`, SchemaName, TableName),
	postgresScriptSelectAll:        fmt.Sprintf(`SELECT id, name, script, type, create_time FROM %s.%s;`, SchemaName, TableName),
	postgresScriptSelectScriptByID: fmt.Sprintf(`SELECT id, name, script, type, create_time FROM %s.%s WHERE id = $1;`, SchemaName, TableName),
	postgresScriptUpdateScriptByID: fmt.Sprintf("UPDATE script = $1 FROM %s.%s WHERE id = $2", SchemaName, TableName),
	postgresScriptDeleteScriptByID: fmt.Sprintf("DELETE FROM %s.%s WHERE id = $1", SchemaName, TableName),
}

func CreateSchema(db *sql.DB) error {
	_, err := db.Exec(scriptSQLString[postgresScriptCreateDatabase])
	if err != nil {
		return err
	}

	return nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(scriptSQLString[postgresScriptCreateTable])
	if err != nil {
		return err
	}

	return nil
}

func InsertScript(db *sql.DB, name string, script string, scriptType string) error {
	_, err := db.Exec(scriptSQLString[postgresScriptRegisterScript], name, script, scriptType)
	if err != nil {
		return err
	}

	return nil
}

func SelectScripts(db *sql.DB) ([]*Script, error) {
	var (
		Scripts []*Script

		ID         uint32
		Name       string
		Scirpt     string
		Type       string
		CreateTime time.Time
	)

	rows, err := db.Query(scriptSQLString[postgresScriptSelectAll])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&ID, &Name, &Scirpt, &Type, &CreateTime); err != nil {
			return nil, err
		}

		Script := &Script{
			ID:         ID,
			Name:       Name,
			Script:     Scirpt,
			Type:       Type,
			CreateTime: CreateTime,
		}

		Scripts = append(Scripts, Script)
	}

	return Scripts, nil
}

func SelectScriptByID(db *sql.DB, id uint32) (*Script, error) {
	row := db.QueryRow(scriptSQLString[postgresScriptSelectScriptByID], id)
	var script *Script
	if err := row.Scan(&script.ID, &script.Name, &script.Script, &script.Type, &script.CreateTime); err != nil {
		return nil, err
	}

	return script, nil
}

func UpdateScriptByID(db *sql.DB, id uint32, script string) error {
	result, err := db.Exec(scriptSQLString[postgresScriptUpdateScriptByID], script, id)
	if err != nil {
		return err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return errors.New("invalid update")
	}

	return nil
}

func DeleteScriptByID(db *sql.DB, id uint32) error {
	result, err := db.Exec(scriptSQLString[postgresScriptDeleteScriptByID], id)
	if err != nil {
		return err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return errors.New("invalid delete")
	}

	return nil
}
