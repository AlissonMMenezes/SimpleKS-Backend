package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func pages(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client := MongoConnector()
	ctx := context.TODO()

	matchStage := bson.D{{Key: "$match", Value: bson.M{"publish": true, "post_type": "page"}}}
	projectStage := bson.D{{Key: "$project", Value: bson.M{"title": 1, "post_name": 1}}}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(projectStage)
	cur, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage})
	if err != nil {
		log.Fatal(err)
	}
	pages := make(map[string][]Post)
	pages["pages"] = []Post{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		pages["pages"] = append(pages["pages"], Post{Post_Name: fmt.Sprintf("%v", result["post_name"]), Title: fmt.Sprintf("%v", result["title"]),
			Content: fmt.Sprintf("%v", result["content"]), Author: fmt.Sprintf("%v", result["author"])})
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Endpoint Hit: ReturnAllPages")
	c.JSON(200, pages)
}

func allPages(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client := MongoConnector()
	ctx := context.TODO()

	matchStage := bson.D{{Key: "$match", Value: bson.M{"post_type": "page"}}}
	projectStage := bson.D{{Key: "$project", Value: bson.M{"title": 1, "post_name": 1}}}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(projectStage)
	cur, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage})
	if err != nil {
		log.Fatal(err)
	}
	pages := make(map[string][]Post)
	pages["pages"] = []Post{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		pages["pages"] = append(pages["pages"], Post{Post_Name: fmt.Sprintf("%v", result["post_name"]), Title: fmt.Sprintf("%v", result["title"]),
			Content: fmt.Sprintf("%v", result["content"]), Author: fmt.Sprintf("%v", result["author"])})
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Endpoint Hit: ReturnAllPages")
	c.JSON(200, pages)
}
