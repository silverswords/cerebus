package model

import (
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID         uint32    `json:"id,omitempty"`
	ScriptID   uint32    `json:"script_id,omitempty"`
	State      string    `json:"state,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

const (
	SchemaName = "project"
	TableName  = "tasks"
)

const (
	postgresTaskCreateDatabase = iota
	postgresTaskCreateTable
	postgresTaskInsertTask
	postgresTaskChangeState
	postgresTaskSelectAll
)

var TaskSQLString = map[int]string{
	postgresTaskCreateDatabase: fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, SchemaName),
	postgresTaskCreateTable: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		id SERIAL PRIMARY KEY,
		script_id INT NOT NULL,
		state CHAR(10) NOT NULL,
		create_time TIMESTAMP
	);`, SchemaName, TableName),
	postgresTaskInsertTask:  fmt.Sprintf(`INSERT INTO %s.%s (script_id, state, create_time) VALUES ($1, $2, current_timestamp);`, SchemaName, TableName),
	postgresTaskChangeState: fmt.Sprintf(`UPDATE %s.%s SET state = $1 WHERE id = $2`, SchemaName, TableName),
	postgresTaskSelectAll:   fmt.Sprintf(`SELECT id, script_id, state, current_timestamp FROM %s.%s;`, SchemaName, TableName),
}

func CreateSchema(db *sql.DB) error {
	_, err := db.Exec(TaskSQLString[postgresTaskCreateDatabase])
	if err != nil {
		return err
	}

	return nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(TaskSQLString[postgresTaskCreateTable])
	if err != nil {
		return err
	}

	return nil
}

func InsertTask(db *sql.DB, scriptID uint32) error {
	_, err := db.Exec(TaskSQLString[postgresTaskInsertTask], scriptID, "Pending")
	if err != nil {
		return err
	}

	return nil
}

func ChangeTaskState(db *sql.DB, id uint32, state string) error {
	_, err := db.Exec(TaskSQLString[postgresTaskChangeState], state, id)
	if err != nil {
		return err
	}

	return nil
}

func SelectTasks(db *sql.DB) ([]*Task, error) {
	var (
		Tasks []*Task

		ID         uint32
		ScriptID   uint32
		State      string
		CreateTime time.Time
	)

	rows, err := db.Query(TaskSQLString[postgresTaskSelectAll])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&ID, &ScriptID, &State, &CreateTime); err != nil {
			return nil, err
		}

		Task := &Task{
			ID:         ID,
			ScriptID:   ScriptID,
			State:      State,
			CreateTime: CreateTime,
		}

		Tasks = append(Tasks, Task)
	}

	return Tasks, nil
}
