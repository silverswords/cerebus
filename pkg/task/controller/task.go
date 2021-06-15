package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/silverswords/cerebus/pkg/scheduler"
	script "github.com/silverswords/cerebus/pkg/script/model"
	"github.com/silverswords/cerebus/pkg/task/model"
)

var (
	bucketName = "task"
	location   = "us-east-1"
)

type TaskController struct {
	db          *sql.DB
	sche        *scheduler.Scheduler
	minioClient *minio.Client
}

func New(db *sql.DB, sche *scheduler.Scheduler, minioClient *minio.Client) *TaskController {
	ctx := context.Background()
	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	return &TaskController{
		db:          db,
		sche:        sche,
		minioClient: minioClient,
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

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "tasks": tasks})

}

func (tc *TaskController) run(c *gin.Context) {
	var req struct {
		ID     uint32                 `json:"id,omitempty" binding:"required"`
		Name   string                 `json:"name,omitempty" binding:"required"`
		Params map[string]interface{} `json:"params,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest})
		return
	}

	script, err := script.SelectScriptByID(tc.db, req.ID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	if err := model.InsertTask(tc.db, req.Name, req.ID); err != nil {
		c.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"status": http.StatusBadGateway})
		return
	}

	taskID, err := model.SelectIDByName(tc.db, req.Name)
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

	resultPath := fmt.Sprintf("%d.txt", taskID)
	if err := tc.sche.Schedule(scheduler.TaskFunc(func(context context.Context) error {
		args := []string{"-e", realScript}
		for key, value := range req.Params {
			args = append(args, fmt.Sprintf("%s=%s", key, value))
		}

		resultFile, err := os.Create(resultPath)
		if err != nil {
			return err
		}
		defer resultFile.Close()
		process := exec.Command("node", args...)
		process.Stdout = resultFile

		if err := process.Run(); err != nil {
			return err
		}

		return nil
	}).AddStartCallback(func(context.Context) error {
		err := model.TaskRun(tc.db, taskID)
		if err != nil {
			return err
		}
		return nil
	}).(scheduler.CallbackTask).AddFinishedCallback(func(context.Context) error {
		info, err := tc.minioClient.FPutObject(context.Background(), bucketName, resultPath, resultPath, minio.PutObjectOptions{})

		if err != nil {
			log.Fatalln(err)
		}

		log.Print(info)

		if err := os.Remove(resultPath); err != nil {
			return err
		}

		if err := model.TaskFinish(tc.db, taskID); err != nil {
			return err
		}

		return nil
	}).(scheduler.RetryTask).WithCatch(func(err error) {
		model.TaskError(tc.db, taskID, err)
	})); err != nil {
		log.Fatal(4)
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK})
}
