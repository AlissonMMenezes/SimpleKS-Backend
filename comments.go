package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Comment struct {
	Author   string             `json:"comment_author"`
	URL      string             `json:"comment_author_url"`
	Date     primitive.DateTime `json:"comment_date"`
	Content  string             `json:"comment_content"`
	Id       int                `json:"comment_id"`
	ParentId int                `json:"comment_parent"`
}

func commentsApproved(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client := MongoConnector()
	ctx := context.TODO()

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("comments")
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"comment_date", -1}})
	cur, err := collection.Find(ctx, bson.D{{Key: "comment_approved", Value: "1"}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	comments := []bson.M{}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		comments = append(comments, result)

	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	fmt.Println(comments)

	json_comments := bson.M{"comments": comments}

	c.JSON(200, json_comments)
}

func allComments(c *gin.Context) {
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
	pages := make(map[string][]Page)
	pages["pages"] = []Page{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		pages["pages"] = append(pages["pages"], Page{Page_Name: fmt.Sprintf("%v", result["post_name"]), Title: fmt.Sprintf("%v", result["title"]),
			Content: fmt.Sprintf("%v", result["content"]), Author: fmt.Sprintf("%v", result["author"])})
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Endpoint Hit: ReturnAllPages")
	c.JSON(200, pages)
}
