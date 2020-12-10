package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Subscriber struct {
	Email string `json:"email"`
}

func subscribe(c *gin.Context) {
	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("subscribers")
	var s Subscriber
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := collection.InsertOne(
		ctx,
		s,
	)
	fmt.Println(err)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("inserted %v Documents!\n", result.InsertedID)
	m := Message{Content: "Subscribed!"}
	c.JSON(200, m)

	return
}
