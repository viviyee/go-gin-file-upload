package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	/* load env */
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	/* setup gin app */
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	/* setup s3 uploader */
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v \n", err)
		return
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		// single file
		file, err := c.FormFile("image")
		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to upload file",
			})
			return
		}
		log.Println(file.Filename)

		// Upload the file to specific dst.
		/*
			err = c.SaveUploadedFile(file, "assets/uploads/"+file.Filename)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"error": "Failed to save file",
				})
				return
			}
		*/
		// open file
		f, openErr := file.Open()
		if openErr != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to open file",
			})
			return
		}

		// upload to aws
		result, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("go-test-projects"),
			Key:    aws.String(file.Filename),
			Body:   f,
			ACL:    "public-read",
		})
		if uploadErr != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to upload file to aws",
			})
			return
		}

		fmt.Println("\n\n\nfile was uploaded to ", result.Location)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"success": "Image was saved succesfullly",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
