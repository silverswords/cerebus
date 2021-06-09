package controller

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/silverswords/cerebus/pkg/script/model"
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

func (sc *Scripscontroller) getScript(c *gin.Context) {
	scripts, err := model.SelectScripts(sc.db)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "script": scripts})
}
