package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Content string `json:"content"`
}

func CreateToken(username string) (string, error) {
	var err error
	//Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = username
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}

func login(c *gin.Context) {

	var u User

	if err := c.ShouldBindJSON(&u); err != nil {
		fmt.Println(u)
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	cur, err := collection.Find(ctx, bson.M{"username": u.Username})

	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	user := User{Username: "", Password: ""}
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		user = User{Username: fmt.Sprintf("%v", result["username"]), Password: fmt.Sprintf("%v", result["password"])}
	}
	defer client.Disconnect(ctx)
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"messsage": "Please provide valid login details"})
		return
	}

	token, err := CreateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	c.Writer.Header().Set("Authenticated", token)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func isAuthenticated(c *gin.Context) {
	fmt.Println(c.Request.Header)
	fmt.Println("test")
	if val, ok := c.Request.Header["Authorization"]; ok {
		fmt.Println(val, ok)
		fmt.Println(c.Request.Header)
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	val := strings.Split(c.Request.Header["Authorization"][0], " ")[1]
	fmt.Println("token to be validated", val)

	tkn, err := jwt.Parse(val, func(token *jwt.Token) (interface{}, error) {
		fmt.Println(val)
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	fmt.Println("Token validated ", tkn)
	if tkn.Valid {
		fmt.Println("Welcome =)")
		return
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			fmt.Println("1 - That's not even a token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			fmt.Println("2 - Timing is everything")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		} else {
			fmt.Println("3 - Couldn't handle this token:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		}
	} else {
		fmt.Println("4 - Couldn't handle this token:", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}

	//c.JSON(http.StatusUnauthorized, "Please authenticate")

	return
}

func changePassword(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		fmt.Println(u)
		c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
		return
	}
	fmt.Println("=======> DEBUG")
	val := strings.Split(c.Request.Header["Authorization"][0], " ")[1]
	tkn, err := jwt.Parse(val, func(token *jwt.Token) (interface{}, error) {
		fmt.Println(val)
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	u.Username = fmt.Sprintf("%v", tkn.Claims.(jwt.MapClaims)["user_id"])
	fmt.Println(u.Username)
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)

	client := MongoConnector()
	ctx := context.TODO()
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	criteria := bson.D{{"username", u.Username}}
	fmt.Println(criteria)
	update := bson.D{{"$set", bson.M{"password": string(hash)}}}
	fmt.Println(update)
	result, err := collection.UpdateOne(
		ctx,
		criteria,
		update,
	)
	defer client.Disconnect(ctx)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.ModifiedCount)
	if result.ModifiedCount == 0 {
		fmt.Println("== Entered into PUT Method")
		fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
		m := Message{Content: "Password not Updated!"}
		c.JSON(200, m)
	} else {
		m := Message{Content: "Password Updated!"}
		c.JSON(200, m)
	}
}

func main() {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowAllHeaders = true

	r.Use(cors.New(config))
	r.POST("/subscribe", subscribe)
	r.GET("/pages", pages)
	r.GET("/posts", posts)
	r.GET("/categories", categories)
	r.GET("/posts/:post_name", getPost)
	r.POST("/login", login)
	r.GET("/validateToken", isAuthenticated)

	authorized := r.Group("/", isAuthenticated)
	authorized.POST("/posts", newPost)
	authorized.PUT("/posts/:post_name", updatePost)
	authorized.POST("/images", images)
	authorized.PUT("/password", changePassword)

	// Listen and Server in 0.0.0.0:8000
	r.Run(":8000")
}
