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

type Post struct {
	Title     string             `json:"title"`
	Post_Name string             `json:"post_name"`
	Post_Type string             `json:"post_type"`
	Content   string             `json:"content"`
	Author    string             `json:"author"`
	Post_Date primitive.DateTime `json:"post_date"`
	Publish   bool               `json:"publish"`
	Layout    string             `json:"layout"`
	Thumbnail string             `json:"thumbnail"`
	Category  string             `json:"category"`
}

func publishedPosts(c *gin.Context) {

	client := MongoConnector()
	ctx := context.TODO()

	fmt.Println(c.Query("category"))

	matchStage := bson.D{{}}

	if c.Query("category") != "" {
		fmt.Println("Filtering by category", c.Query("category"))
		matchStage = bson.D{{Key: "$match", Value: bson.M{"publish": true, "post_type": "post", "category": c.Query("category")}}}
		fmt.Println(matchStage)
	} else if (c.Query("term")) != "" {
		fmt.Println("Filtering by term", c.Query("term"))
		matchStage = bson.D{{Key: "$match", Value: bson.M{"publish": true, "post_type": "post", "content": bson.M{"$regex": "/*." + c.Query("term") + ".*/", "$options": "i"}}}}
		fmt.Println(matchStage)

	} else {
		matchStage = bson.D{{Key: "$match", Value: bson.M{"publish": true, "post_type": "post"}}}
		fmt.Println(matchStage)
	}

	//r := bson.M{"input": "$content", "chars": bson.M{"$regexFindAll": bson.M{"input": "$content", "regex": "\\<.*?\\>"}}}
	projection := bson.M{"title": 1, "thumbnail": 1, "post_type": 1, "post_name": 1, "author": 1,
		"post_date": 1, "category": 1, "content": bson.M{"$substrCP": bson.A{"$content", 0, 400}}}
	lookup := bson.D{{Key: "$lookup", Value: bson.M{"from": "categories", "localField": "category", "foreignField": "name", "as": "category"}}}
	unwind := bson.D{{Key: "$unwind", Value: bson.M{"path": "$category"}}}
	projectStage := bson.D{{Key: "$project", Value: projection}}
	orderStage := bson.D{{Key: "$sort", Value: bson.M{"post_date": -1}}}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")
	fmt.Println(projectStage)
	cur, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, orderStage, lookup, unwind, projectStage})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	posts := []bson.M{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Endpoint Hit: returnAllArticles")
	c.JSON(200, bson.M{"posts": posts})
}

func allPosts(c *gin.Context) {

	client := MongoConnector()
	ctx := context.TODO()
	matchStage := bson.D{{}}

	matchStage = bson.D{{Key: "$match", Value: bson.M{"post_type": "post"}}}
	fmt.Println(matchStage)

	projectStage := bson.D{{Key: "$project", Value: bson.M{"title": 1, "post_type": 1, "post_name": 1, "author": 1, "post_date": 1, "content": bson.M{"$substrCP": bson.A{"$content", 0, 400}}}}}
	orderStage := bson.D{{Key: "$sort", Value: bson.M{"post_date": -1}}}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")
	fmt.Println(projectStage)
	cur, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, orderStage, projectStage})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	posts := []bson.M{}

	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Endpoint Hit: returnAllArticles")
	c.JSON(200, bson.M{"posts": posts})
}

func updatePost(c *gin.Context) {
	var p Post
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(p.Title, p.Post_Name)
	criteria := bson.M{"post_name": p.Post_Name}
	update := bson.D{{Key: "$set", Value: bson.M{"title": p.Title, "content": p.Content, "publish": p.Publish, "thumbnail": p.Thumbnail, "category": p.Category}}}
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
		m := Message{Content: "Post not Updated!"}
		c.JSON(200, m)
	} else {
		m := Message{Content: "Post Saved!"}
		c.JSON(200, m)
	}

	return

}

func newPost(c *gin.Context) {
	var p Post
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	fmt.Println(p.Title, p.Post_Name)
	reg, err := regexp.Compile("[^a-zA-Z0-9/ ]+")
	if err != nil {
		log.Fatal(err)
	}
	p.Author = "Alisson Machado"
	p.Post_Date = primitive.NewDateTimeFromTime(time.Now())
	p.Post_Name = strings.ToLower(p.Title)
	p.Post_Name = reg.ReplaceAllString(p.Post_Name, "")
	p.Post_Name = strings.ReplaceAll(p.Post_Name, " ", "-")
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
	m := Message{Content: "Post Saved!"}
	c.JSON(200, m)

	return

}

func getPost(c *gin.Context) {
	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")

	post_name := c.Param("post_name")
	fmt.Println(post_name)
	p := Post{Title: "not found", Post_Name: "not found", Content: "not found"}
	cur, err := collection.Find(ctx, bson.D{{Key: "post_name", Value: post_name}})
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
		if result["layout"] != nil {
			result["layout"] = result["layout"].(string)
		} else {
			result["layout"] = ""
		}

		c.JSON(200, result)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	fmt.Println(p)
}
