package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func categories(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client := MongoConnector()
	ctx := context.TODO()

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	matchStage := bson.D{{"$match", bson.M{"category": bson.M{"$exists": true}}}}
	projectStage := bson.D{{"$project", bson.M{"category": 1, "_id": 0}}}
	groupStage := bson.D{{"$group", bson.M{"_id": "$category"}}}
	fmt.Println(matchStage)
	fmt.Println(projectStage)
	fmt.Println(groupStage)
	fmt.Println(mongo.Pipeline{matchStage, projectStage})
	cur, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage, groupStage})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	categories := []string{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		categories = append(categories, fmt.Sprintf("%v", result["_id"]))
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Endpoint Hit: ReturnAllPages")
	c.JSON(200, bson.M{"categories": categories})
}

func getCategory(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)

	matchStage := bson.D{{Key: "$match", Value: bson.M{"publish": true, "post_type": "page"}}}
	projectStage := bson.D{{Key: "$project", Value: bson.M{"title": 1, "post_name": 1}}}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(projectStage)
	cur, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, projectStage})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

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

	fmt.Println("Endpoint Hit: ReturnAllPages")
	c.JSON(200, pages)
}
