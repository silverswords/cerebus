package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Task struct {
	ID         uint32    `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	ScriptID   uint32    `json:"script_id,omitempty"`
	ScriptName string    `json:"script_name,omitempty"`
	ScriptType string    `json:"script_type,omitempty"`
	State      string    `json:"state,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
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
	postgresTaskRun
)

var TaskSQLString = map[int]string{
	postgresTaskCreateDatabase: fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, SchemaName),
	postgresTaskCreateTable: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) NOT NULL,
		script_id INT NOT NULL,
		state CHAR(10) NOT NULL,
		start_time TIMESTAMP NOT NULL DEFAULT timestamp '2000-01-01 00:00:00',
		create_time TIMESTAMP NOT NULL DEFAULT timestamp '2000-01-01 00:00:00'
	);`, SchemaName, TableName),
	postgresTaskInsertTask:  fmt.Sprintf(`INSERT INTO %s.%s (name, script_id, state, create_time) VALUES ($1, $2, $3, current_timestamp);`, SchemaName, TableName),
	postgresTaskChangeState: fmt.Sprintf(`UPDATE %s.%s SET state = $1 WHERE id = $2`, SchemaName, TableName),
	postgresTaskSelectAll:   fmt.Sprintf(`SELECT tasks.id, tasks.name, tasks.script_id, scripts.name as script_name, scripts.type, tasks.state, tasks.start_time, tasks.create_time FROM %s.%s LEFT JOIN project.scripts ON scripts.id = tasks.script_id;`, SchemaName, TableName),
	postgresTaskRun:         fmt.Sprintf(`UPDATE %s.%s SET state = $1, start_time = current_timestamp WHERE id = $2`, SchemaName, TableName),
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

func InsertTask(db *sql.DB, name string, scriptID uint32) (uint32, error) {
	result, err := db.Exec(TaskSQLString[postgresTaskInsertTask], name, scriptID, "Pending")
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint32(id), nil
}

func ChangeTaskState(db *sql.DB, id uint32, state string) error {
	result, err := db.Exec(TaskSQLString[postgresTaskChangeState], state, id)
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

func SelectTasks(db *sql.DB) ([]*Task, error) {
	var (
		Tasks []*Task

		ID         uint32
		Name       string
		ScriptID   uint32
		ScriptName string
		ScriptType string

		State      string
		StartTime  time.Time
		CreateTime time.Time
	)

	rows, err := db.Query(TaskSQLString[postgresTaskSelectAll])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&ID, &Name, &ScriptID, &ScriptName, &ScriptType, &State, &StartTime, &CreateTime); err != nil {
			return nil, err
		}

		Task := &Task{
			ID:         ID,
			Name:       Name,
			ScriptID:   ScriptID,
			ScriptName: ScriptName,
			ScriptType: ScriptType,
			State:      State,
			StartTime:  StartTime,
			CreateTime: CreateTime,
		}

		Tasks = append(Tasks, Task)
	}

	return Tasks, nil
}

func TaskRun(db *sql.DB, id uint32) error {
	result, err := db.Exec(TaskSQLString[postgresTaskRun], "Running", id)
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
