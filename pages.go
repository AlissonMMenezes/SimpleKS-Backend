package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Page struct {
	Title     string             `json:"title"`
	Page_Name string             `json:"page_name"`
	Page_Type string             `json:"page_type"`
	Content   string             `json:"content"`
	Author    string             `json:"author"`
	Page_Date primitive.DateTime `json:"page_date"`
	Publish   bool               `json:"boolean"`
	Layout    string             `json:"layout"`
}

func pages(c *gin.Context) {
	fmt.Println("Endpoint Hit: ReturnAllPages")

	client := MongoConnector()
	ctx := context.TODO()

	matchStage := bson.D{{Key: "$match", Value: bson.M{"publish": true, "post_type": "page"}}}
	projectStage := bson.D{{Key: "$project", Value: bson.M{"title": 1, "post_name": 1, "layout": 1}}}
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
		p := Page{Page_Name: fmt.Sprintf("%v", result["post_name"]),
			Title:   fmt.Sprintf("%v", result["title"]),
			Content: fmt.Sprintf("%v", result["content"]),
			Author:  fmt.Sprintf("%v", result["author"]),
			Layout:  fmt.Sprintf("%v", result["layout"])}
		pages["pages"] = append(pages["pages"], p)
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

func updatePage(c *gin.Context) {
	var p Page
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(p.Title, p.Page_Name)
	criteria := bson.M{"post_name": p.Page_Name}
	update := bson.D{{Key: "$set", Value: bson.M{"title": p.Title, "content": p.Content, "layout": p.Layout}}}
	fmt.Println(criteria)
	result, err := collection.UpdateOne(
		ctx,
		criteria,
		update,
	)
	fmt.Println(err)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	if result.ModifiedCount == 0 {
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
		m := Message{Content: "Page not Updated!"}
		c.JSON(200, m)
	} else {
		m := Message{Content: "Page Saved!"}
		c.JSON(200, m)
	}

	return

}

func newPage(c *gin.Context) {
	var p Page
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(p.Title, p.Page_Name)
	reg, err := regexp.Compile("[^a-zA-Z0-9/ ]+")
	if err != nil {
		log.Fatal(err)
	}
	p.Author = "Alisson Machado"
	p.Page_Date = primitive.NewDateTimeFromTime(time.Now())
	p.Page_Name = strings.ToLower(p.Title)
	p.Page_Name = reg.ReplaceAllString(p.Page_Name, "")
	p.Page_Name = strings.ReplaceAll(p.Page_Name, " ", "-")
	p.Page_Type = p.Page_Type
	p.Layout = p.Layout

	result, err := collection.InsertOne(
		ctx,
		p,
	)
	fmt.Println(err)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(500, Message{Content: "Error, verify if the post name is unique"})

	}
	fmt.Printf("inserted %v Documents!\n", result.InsertedID)
	defer client.Disconnect(ctx)
	m := Message{Content: "Page Saved!"}
	c.JSON(200, m)

	return

}

func getPage(c *gin.Context) {
	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	post_name := c.Param("post_name")
	fmt.Println(post_name)
	p := Page{Title: "not found", Page_Name: "not found", Content: "not found"}
	cur, err := collection.Find(ctx, bson.D{{Key: "post_name", Value: fmt.Sprintf("%v", post_name)}})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		p := Page{Title: fmt.Sprintf("%v", result["title"]),
			Page_Name: fmt.Sprintf("%v", result["post_name"]),
			Content:   fmt.Sprintf("%v", result["content"]),
			Page_Type: fmt.Sprintf("%v", result["post_type"]),
			Page_Date: result["post_date"].(primitive.DateTime),
			Layout:    fmt.Sprintf("%v", result["layout"]),
		}
		c.JSON(200, p)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	fmt.Println(p)
}
