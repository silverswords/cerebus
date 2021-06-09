package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
	"github.com/silverswords/cerebus/pkg/task/model/postgres"
)

type TasksController struct {
	db *sql.DB
}

func New(db *sql.DB) *TasksController {
	return &TasksController{
		db: db,
	}
}

func (tc *TasksController) RegisterRouter(r gin.IRouter) {
	if err := postgres.CreateSchema(tc.db); err != nil {
		log.Fatal(err)
		return
	}

	if err := postgres.CreateTable(tc.db); err != nil {
		log.Fatal(err)
		return
	}

	r.POST("/register", tc.registerTask)
	r.POST("/run", tc.run)

	r.GET("/tasks", tc.getTasks)
}

func (tc *TasksController) registerTask(c *gin.Context) {
	var req struct {
		Script string `json:"script,omitempty" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
		return
	}

	if err := postgres.RegisterTask(tc.db, req.Script); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
}

func (tc *TasksController) run(c *gin.Context) {
	var req struct {
		ID     uint32                 `json:"id,omitempty" binding:"required"`
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

	script, err := postgres.SelectScriptByID(tc.db, req.ID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	realScript, err := url.QueryUnescape(script)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError})
		return
	}

	file, err := os.Create(fmt.Sprintf("%d.txt", req.ID))
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError})
		return
	}
	defer file.Close()

	originStdout := os.Stdout
	os.Stdout = file
	defer func() {
		os.Stdout = originStdout
	}()

	result, err := vm.Run(realScript)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "result": result})
}

func (tc *TasksController) getTasks(c *gin.Context) {
	tasks, err := postgres.SelectTasks(tc.db)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "tasks": tasks})
}
