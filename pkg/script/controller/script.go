package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"
	net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
	"github.com/silverswords/cerebus/pkg/script/model"
	task "github.com/silverswords/cerebus/pkg/task/model"
)

type Scripscontroller struct {
	db *sql.DB
}

func New(db *sql.DB) *Scripscontroller {
	return &Scripscontroller{
		db: db,
	}
}

func (sc *Scripscontroller) RegisterRouter(r gin.IRouter) {
	if err := model.CreateSchema(sc.db); err != nil {
		log.Fatal(err)
		return
	}

	if err := model.CreateTable(sc.db); err != nil {
		log.Fatal(err)
		return
	}

	r.POST("/register", sc.registerTask)
	r.POST("/run", sc.run)

	r.GET("/Script", sc.getScript)
}

func (sc *Scripscontroller) registerTask(c *gin.Context) {
	var req struct {
		Script string `json:"script,omitempty" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
		return
	}

	if err := model.InsertScript(sc.db, req.Script); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
}

func (sc *Scripscontroller) run(c *gin.Context) {
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

	if err := task.InsertTask(sc.db, req.ID); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	script, err := model.SelectScriptByID(sc.db, req.ID)
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

func (sc *Scripscontroller) getScript(c *gin.Context) {
	scripts, err := model.SelectScripts(sc.db)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "script": scripts})
}
