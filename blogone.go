/*
 * This program is used to access the information for a single blog post from lauchlantoal.com
 * It is designed to work as an AWS Lambda function, proxied through AWS APIGateway
 */

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

//Frontend expects a JSON response with title and content
type Post struct {
	Title string `json:"title"`
	Content string `json:"content"`
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
				Headers: header,
				Body: "Error accessing database credentials",
				StatusCode: 500,
			}, nil
	}
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	//Ensure database connections are always closed when function exits with defer
	defer db.Close()
	
	var url = request.PathParameters["post"]
	
	//Success = true if no error, found = true if desired post exists
	var success bool = true
	var found bool = false
	
	rows, err := db.Query("SELECT posttitle, posttext FROM posts WHERE posturl = ?", url)
	if err != nil {
		success = false
	}
	defer rows.Close()
	
	//Load the database data into a Post struct to return
	var post Post
	if success {
		for rows.Next() {
			found = true
			if err := rows.Scan(&post.Title, &post.Content); err != nil {
				success = false
				break
			}
		}
	}
	
	//Return content or appropriate error
	if success && found {
		jsondata, err := json.Marshal(post)
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
	} else if success && !found {
		return events.APIGatewayProxyResponse{
			Headers: header,
			Body: "Record not found",
			StatusCode: 404,
		}, nil
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
