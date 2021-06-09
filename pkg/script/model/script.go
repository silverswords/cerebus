package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Script struct {
	ID         uint32    `json:"id,omitempty"`
	Script     string    `json:"script,omitempty"`
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
)

var ScriptSQLString = map[int]string{
	postgresScriptCreateDatabase: fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, SchemaName),
	postgresScriptCreateTable: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		id serial PRIMARY KEY,
		script text NOT NULL,
		create_time timestamp
	);`, SchemaName, TableName),
	postgresScriptRegisterScript:   fmt.Sprintf(`INSERT INTO %s.%s (script, create_time) VALUES ($1, current_timestamp);`, SchemaName, TableName),
	postgresScriptSelectAll:        fmt.Sprintf(`SELECT id, script, current_timestamp FROM %s.%s;`, SchemaName, TableName),
	postgresScriptSelectScriptByID: fmt.Sprintf(`SELECT script FROM %s.%s WHERE id = $1;`, SchemaName, TableName),
}

func CreateSchema(db *sql.DB) error {
	_, err := db.Exec(ScriptSQLString[postgresScriptCreateDatabase])
	if err != nil {
		return err
	}

	return nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(ScriptSQLString[postgresScriptCreateTable])
	if err != nil {
		return err
	}

	return nil
}

func InsertScript(db *sql.DB, script string) error {
	_, err := db.Exec(ScriptSQLString[postgresScriptRegisterScript], script)
	if err != nil {
		return err
	}

	return nil
}

func SelectScripts(db *sql.DB) ([]*Script, error) {
	var (
		Scripts []*Script

		ID         uint32
		Scirpt     string
		CreateTime time.Time
	)

	rows, err := db.Query(ScriptSQLString[postgresScriptSelectAll])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&ID, &Scirpt, &CreateTime); err != nil {
			return nil, err
		}

		Script := &Script{
			ID:         ID,
			Script:     Scirpt,
			CreateTime: CreateTime,
		}

		Scripts = append(Scripts, Script)
	}

	return Scripts, nil
}

func SelectScriptByID(db *sql.DB, id uint32) (string, error) {
	row := db.QueryRow(ScriptSQLString[postgresScriptSelectScriptByID], id)
	var script string
	if err := row.Scan(&script); err != nil {
		return "", err
	}

	return script, nil
}
