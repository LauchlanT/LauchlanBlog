/*
 * This program is used to access the information for all project posts from lauchlantoal.com
 * It is designed to work as an AWS Lambda function, proxied through AWS APIGateway
 */

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

// Frontend expects a JSON response with url, thumbnail, title, and desc (description)
type Post struct {
	Url       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	//Header must set appropriate content type and origin options to avoid CORS restrictions
	var header = make(map[string]string)
	header["Access-Control-Allow-Origin"] = "*"
	header["Content-Type"] = "text/json"
	header["X-Content-Type-Options"] = "nosniff"

	//Database connection string is stored as environment variable so code can be freely shared
	dsn, envFound := os.LookupEnv("DSN")
	if !envFound {
		return events.APIGatewayProxyResponse{
			Headers:    header,
			Body:       "Error accessing database credentials",
			StatusCode: 500,
		}, nil
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	//Ensure database connections are always closed when function exits with defer
	defer db.Close()

	//Success = true if no error
	var success bool = true

	rows, err := db.Query("SELECT posturl, thumbnail, posttitle, postdesc FROM posts WHERE posttype = 1 ORDER BY postdate DESC")
	if err != nil {
		success = false
	}
	defer rows.Close()

	//Load the database data into a slice of Post structs to return
	var posts []Post
	if success {
		for rows.Next() {
			var post Post
			if err := rows.Scan(&post.Url, &post.Title, &post.Desc); err != nil {
				success = false
				break
			} else {
				posts = append(posts, post)
			}
		}
	}

	//Return content or appropriate error
	if success {
		jsondata, err := json.Marshal(posts)
		if err == nil {
			return events.APIGatewayProxyResponse{
				Headers:    header,
				Body:       string(jsondata),
				StatusCode: 200,
			}, nil
		} else {
			return events.APIGatewayProxyResponse{
				Headers:    header,
				Body:       "Error Marshalling",
				StatusCode: 500,
			}, nil
		}
	} else {
		return events.APIGatewayProxyResponse{
			Headers:    header,
			Body:       "Server Error",
			StatusCode: 500,
		}, nil
	}

}

func main() {
	lambda.Start(Handler)
}
