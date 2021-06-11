package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/silverswords/cerebus/pkg/scheduler"
	script "github.com/silverswords/cerebus/pkg/script/controller"
	task "github.com/silverswords/cerebus/pkg/task/controller"
)

func main() {
	router := gin.Default()
	router.Use(Cors())

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s database=%s sslmode=disable",
		"server", "5432", "postgres", "123456", "project")
	// fmt.Println(psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	endpoint := "server:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	sche := scheduler.New()
	go sche.Start(2)
	scriptController := script.New(db)
	taskController := task.New(db, sche, minioClient)

	scriptController.RegisterRouter(router)
	taskController.RegisterRouter(router)

	log.Fatal(router.Run("0.0.0.0:10001"))
	sche.Wait()
	sche.Stop()
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}
