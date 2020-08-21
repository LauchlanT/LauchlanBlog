package main

import (
	"os"
  "encoding/json"
  "log"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Post struct {
	Url string `json:"url"`
	Title string `json:"title"`
	Desc string `json:"desc"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	
	var header = make(map[string]string)
	header["Access-Control-Allow-Origin"] = "*"
	header["Content-Type"] = "text/json"
	header["X-Content-Type-Options"] = "nosniff"
	
	dsn, envFound := os.LookupEnv("DSN")
	if !envFound {
		return events.APIGatewayProxyResponse{
				Headers: header,
				Body: "Error accessing database credentials",
				StatusCode: 500,
			}, nil
	}
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	var success bool = true
	var posts []Post
	rows, err := db.Query("SELECT posturl, posttitle, postdesc FROM posts WHERE posttype = 0 ORDER BY postdate DESC")
	if err != nil {
		success = false
	}
	defer rows.Close()
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
	
	if success {
		jsondata, err := json.Marshal(posts)
		if err == nil {
			return events.APIGatewayProxyResponse{
				Headers: header,
				Body: string(jsondata),
				StatusCode: 200,
			}, nil
		} else {
			return events.APIGatewayProxyResponse{
				Headers: header,
				Body: "Error Marshalling",
				StatusCode: 500,
			}, nil
		}
	} else {
		return events.APIGatewayProxyResponse{
			Headers: header,
			Body: "Server Error",
			StatusCode: 500,
		}, nil
	}
	
}

func main() {
	lambda.Start(Handler)
}
