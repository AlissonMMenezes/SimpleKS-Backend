package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/api/option"
)

func images(c *gin.Context) {
	data, err := base64.StdEncoding.DecodeString(os.Getenv("GCP_SERVICEKEY"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	err = ioutil.WriteFile("/tmp/key.json", data, 0644)
	fmt.Println(err)
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("/tmp/key.json"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("========= IMAGE =====")
	f, img, _ := c.Request.FormFile("image")
	fmt.Println(img.Filename)
	currentTime := time.Now()
	filename := "content/" + currentTime.Format("2006/01") + "/" + img.Filename
	fmt.Println("======= Uploading Image =======")
	bucket := client.Bucket(os.Getenv("GCP_BUCKET"))
	obj := bucket.Object(filename)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, f); err != nil {
		fmt.Println("Error: ", err)
	}
	if err := w.Close(); err != nil {
		fmt.Println("Error: ", err)
	}
	url := "https://storage.googleapis.com/"
	url += os.Getenv("GCP_BUCKET")
	url += "/" + filename
	fmt.Println(url)
	c.JSON(200, bson.M{"url": url})

}
