package controller

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/silverswords/cerebus/pkg/task/model"
)

type TaskController struct {
	db *sql.DB
}

func New(db *sql.DB) *TaskController {
	return &TaskController{
		db: db,
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
