package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Task struct {
	ID           uint32    `json:"id,omitempty"`
	Name         string    `json:"name,omitempty"`
	ScriptID     uint32    `json:"script_id,omitempty"`
	ScriptName   string    `json:"script_name,omitempty"`
	ScriptType   string    `json:"script_type,omitempty"`
	State        string    `json:"state,omitempty"`
	Error        string    `json:"error,omitempty"`
	StartTime    time.Time `json:"start_time,omitempty"`
	FinishedTime time.Time `json:"finished_time,omitempty"`
	CreateTime   time.Time `json:"create_time,omitempty"`
}

const (
	SchemaName = "project"
	TableName  = "tasks"
)

const (
	postgresTaskCreateDatabase = iota
	postgresTaskCreateTable
	postgresTaskInsertTask
	postgresTaskSelectID
	postgresTaskSelectAll
	postgresTaskRun
	postgresTaskFinish
	postgresTaskError
)

var TaskSQLString = map[int]string{
	postgresTaskCreateDatabase: fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s;`, SchemaName),
	postgresTaskCreateTable: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.%s (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) UNIQUE NOT NULL ,
		script_id INT NOT NULL,
		state VARCHAR(20) NOT NULL,
		error TEXT NOT NULL DEFAULT '',
		start_time TIMESTAMP NOT NULL DEFAULT timestamp '2000-01-01 00:00:00',
		finished_time TIMESTAMP NOT NULL  DEFAULT timestamp '2000-01-01 00:00:00',
		create_time TIMESTAMP NOT NULL DEFAULT timestamp '2000-01-01 00:00:00'
	);`, SchemaName, TableName),
	postgresTaskInsertTask: fmt.Sprintf(`INSERT INTO %s.%s (name, script_id, state, create_time) VALUES ($1, $2, 'Pending', current_timestamp);`, SchemaName, TableName),
	postgresTaskSelectAll:  fmt.Sprintf(`SELECT tasks.id, tasks.name, tasks.script_id, scripts.name as script_name, scripts.type, tasks.state, tasks.error, tasks.start_time, tasks.finished_time, tasks.create_time FROM %s.%s LEFT JOIN project.scripts ON scripts.id = tasks.script_id;`, SchemaName, TableName),
	postgresTaskSelectID:   fmt.Sprintf(`SELECT id FROM %s.%s WHERE name = $1`, SchemaName, TableName),
	postgresTaskRun:        fmt.Sprintf(`UPDATE %s.%s SET state = 'Running', start_time = current_timestamp WHERE id = $1`, SchemaName, TableName),
	postgresTaskFinish:     fmt.Sprintf(`UPDATE %s.%s SET state = 'Finished', finished_time = current_timestamp WHERE id = $1`, SchemaName, TableName),
	postgresTaskError:      fmt.Sprintf(`UPDATE %s.%s SET state = 'Error', error = $1, finished_time = current_timestamp WHERE id = $2`, SchemaName, TableName),
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

func InsertTask(db *sql.DB, name string, scriptID uint32) error {
	result, err := db.Exec(TaskSQLString[postgresTaskInsertTask], name, scriptID)
	if err != nil {
		return err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if num == 0 {
		return errors.New("invalid insert")
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

		State        string
		Error        string
		StartTime    time.Time
		FinishedTime time.Time
		CreateTime   time.Time
	)

	rows, err := db.Query(TaskSQLString[postgresTaskSelectAll])
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&ID, &Name, &ScriptID, &ScriptName, &ScriptType, &State, &Error, &StartTime, &FinishedTime, &CreateTime); err != nil {
			return nil, err
		}

		Task := &Task{
			ID:           ID,
			Name:         Name,
			ScriptID:     ScriptID,
			ScriptName:   ScriptName,
			ScriptType:   ScriptType,
			State:        State,
			Error:        Error,
			StartTime:    StartTime,
			FinishedTime: FinishedTime,
			CreateTime:   CreateTime,
		}

		Tasks = append(Tasks, Task)
	}

	return Tasks, nil
}

func SelectIDByName(db *sql.DB, name string) (uint32, error) {
	row := db.QueryRow(TaskSQLString[postgresTaskSelectID], name)
	var result uint32
	if err := row.Scan(&result); err != nil {
		return 0, err
	}

	return result, nil
}

func TaskRun(db *sql.DB, id uint32) error {
	result, err := db.Exec(TaskSQLString[postgresTaskRun], id)
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

func TaskFinish(db *sql.DB, id uint32) error {
	result, err := db.Exec(TaskSQLString[postgresTaskFinish], id)
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

func TaskError(db *sql.DB, id uint32, err error) error {
	result, err := db.Exec(TaskSQLString[postgresTaskFinish], id, err.Error())
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
