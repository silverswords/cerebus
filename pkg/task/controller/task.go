package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
	"github.com/silverswords/cerebus/pkg/scheduler"
	script "github.com/silverswords/cerebus/pkg/script/model"
	"github.com/silverswords/cerebus/pkg/task/model"
)

type TaskController struct {
	db   *sql.DB
	sche *scheduler.Scheduler
}

func New(db *sql.DB, sche *scheduler.Scheduler) *TaskController {
	return &TaskController{
		db:   db,
		sche: sche,
	}
}

func (tc *TaskController) RegisterRouter(r gin.IRouter) {
	if err := model.CreateSchema(tc.db); err != nil {
		log.Fatal(err)
		return
	}

	if err := model.CreateTable(tc.db); err != nil {
		log.Fatal(err)
		return
	}

	r.GET("/tasks", tc.getTasks)
	r.POST("/run", tc.run)
}

func (tc *TaskController) getTasks(c *gin.Context) {
	tasks, err := model.SelectTasks(tc.db)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "script": tasks})

}

func (tc *TaskController) run(c *gin.Context) {
	var req struct {
		ID     uint32                 `json:"id,omitempty" binding:"required"`
		Name   string                 `json:"name,omitempty"`
		Params map[string]interface{} `json:"params,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
		return
	}

	vm := otto.New()
	for key, value := range req.Params {
		if err := vm.Set(key, value); err != nil {
			c.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
			return
		}
	}

	script, err := script.SelectScriptByID(tc.db, req.ID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	taskID, err := model.InsertTask(tc.db, req.Name, req.ID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	realScript, err := url.QueryUnescape(script.Script)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError})
		return
	}

	tc.sche.Schedule(scheduler.TaskFunc(func(context context.Context) error {
		file, err := os.Create(fmt.Sprintf("%d.txt", taskID))
		if err != nil {

			return err
		}
		defer file.Close()

		originStdout := os.Stdout
		os.Stdout = file
		defer func() {
			os.Stdout = originStdout
		}()

		_, err = vm.Run(realScript)
		if err != nil {
			return err
		}

		return nil
	}).AddStartCallback(func(context.Context) error {
		err := model.ChangeTaskState(tc.db, taskID, "Running")
		if err != nil {
			return err
		}
		return nil
	}).(scheduler.CallbackTask).AddFinishedCallback(func(context.Context) error {
		err := model.ChangeTaskState(tc.db, taskID, "Finished")
		if err != nil {
			return err
		}
		return nil
	}))

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
}
