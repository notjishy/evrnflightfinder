package flightfinder

import (
	"log"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Collection *mongo.Collection
var Ctx = context.TODO()

func Login(credentials string) (*mongo.Client) {
	// login to db
	clientOptions := options.Client().ApplyURI(credentials)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil { log.Fatal(err) }

	// verify db server was found and connected
	err = client.Ping(Ctx, nil)
	if err != nil { log.Fatal(err) }

	return client
}
