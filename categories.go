package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Category struct {
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}

func updateCategory(c *gin.Context) {
	var cat Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("categories")
	query := bson.M{"name": cat.Name}
	update := bson.M{"$set": bson.M{"thumbnail": cat.Thumbnail}}
	result, err := collection.UpdateOne(ctx, query, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Println("Error! ", err)
		c.AbortWithStatusJSON(500, bson.M{"Content": "Failed to save thumbnail"})
	}
	fmt.Print("Result: ", result)
	c.JSON(200, bson.M{"message": "Category updated!"})
}

func categories(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client := MongoConnector()
	ctx := context.TODO()

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("categories")

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	categories := []Category{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		cat := Category{Name: result["name"].(string), Thumbnail: result["thumbnail"].(string)}
		categories = append(categories, cat)
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
	defer client.Disconnect(ctx)
	fmt.Println("Endpoint Hit: ReturnAllPages")
	c.JSON(200, pages)
}
