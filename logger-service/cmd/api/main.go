package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/cmd/api/data"
	"net"
	"net/http"
	"net/rpc"
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

	//Register the RPC Server
	err = rpc.Register(new(RPCServer))

	go app.rpcListen()

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

func (app *Config) rpcListen() error {
	log.Println("starting rpc server on port ", rpcPort)

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))

	if err != nil {
		return err
	}

	defer listen.Close()

	for {
		rpcCon, err := listen.Accept()

		if err != nil {
			continue
		}
		// go
		go rpc.ServeConn(rpcCon)
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
