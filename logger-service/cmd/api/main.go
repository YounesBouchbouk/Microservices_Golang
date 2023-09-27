package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/cmd/api/data"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "3001"
	rpcPort  = "5001"
	mongoURL = "mongodb://localhost:27017"
	gRPCPort = "5001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	//connect tp mongoDb

	mongoClient, err := connectToMongo()

	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	//create a context in order to disconnect

	vtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	//close connectio
	defer func() {
		if err := client.Disconnect(vtx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// start web server
	// go app.serve()
	log.Println("Starting service on port", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()

	if err != nil {
		fmt.Println(err)
		log.Panic()
	}

}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}
}

func connectToMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)

	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect

	c, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Println("error connection to mongodb", err)

		return nil, err
	}

	return c, err
}
