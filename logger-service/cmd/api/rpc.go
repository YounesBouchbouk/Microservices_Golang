package main

import (
	"context"
	"fmt"
	"logger-service/cmd/api/data"
	"time"
)

type RPCServer struct{}

type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LoginInfo(payload RPCPayload, resp *string) error {

	collection := client.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})

	if err != nil {
		fmt.Println("error writing to mongo", err)
		return nil
	}

	*resp = "Processes payload via RPC:" + payload.Name

	return nil

}
