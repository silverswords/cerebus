package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID         uint32    `json:"id,omitempty"`
	Script     string    `json:"script,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

const (
	SchemaName = "project"
	TableName  = "tasks"
)

const (
	postgresTaskCreateDatabase = iota
	postgresTaskCreateTable
	postgresTaskRegisterTask
	postgresTaskSelectAll
	postgresTaskSelectScriptByID
)

var taskSQLString = map[int]string{
	postgresTaskCreateDatabase: fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, SchemaName),
	postgresTaskCreateTable: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		id serial,
		script text,
		create_time timestamp
	);`, SchemaName, TableName),
	postgresTaskRegisterTask:     fmt.Sprintf(`INSERT INTO %s.%s (script, create_time) VALUES ($1, current_timestamp);`, SchemaName, TableName),
	postgresTaskSelectAll:        fmt.Sprintf(`SELECT id, script, current_timestamp FROM %s.%s;`, SchemaName, TableName),
	postgresTaskSelectScriptByID: fmt.Sprintf(`SELECT script FROM %s.%s WHERE id = $1;`, SchemaName, TableName),
}

func CreateSchema(db *sql.DB) error {
	_, err := db.Exec(taskSQLString[postgresTaskCreateDatabase])
	if err != nil {
		return err
	}

	return nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(taskSQLString[postgresTaskCreateTable])
	if err != nil {
		return err
	}

	return nil
}

func RegisterTask(db *sql.DB, script string) error {
	_, err := db.Exec(taskSQLString[postgresTaskRegisterTask], script)
	if err != nil {
		return err
	}

	return nil
}

func SelectTasks(db *sql.DB) ([]*Task, error) {
	var (
		tasks []*Task

		ID         uint32
		Scirpt     string
		CreateTime time.Time
	)

	rows, err := db.Query(taskSQLString[postgresTaskSelectAll])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&ID, &Scirpt, &CreateTime); err != nil {
			return nil, err
		}

		task := &Task{
			ID:         ID,
			Script:     Scirpt,
			CreateTime: CreateTime,
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func SelectScriptByID(db *sql.DB, id uint32) (string, error) {
	row := db.QueryRow(taskSQLString[postgresTaskSelectScriptByID], id)
	var script string
	if err := row.Scan(&script); err != nil {
		return "", err
	}

	return script, nil
}
