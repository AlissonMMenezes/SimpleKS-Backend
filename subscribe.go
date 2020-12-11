package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

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

	domain := strings.Split(s.Email, "@")[1]
	fmt.Printf(domain)
	mx, err := net.LookupMX(domain)
	fmt.Println(mx)
	if err != nil || len(mx) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, Message{Content: "Invalid Domain!"})
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
