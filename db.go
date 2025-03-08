package flightfinder

import (
	"context"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAirportViaCode(airportCode string, org string, client *mongo.Client, ctx context.Context) (airportInfo, bool) {
	var airport airportInfo
	if airportCode == "" {
		return airport, false
	}
	airportCode = strings.ToUpper(airportCode)

	coll := client.Database("airports").Collection("airports")

	filter := bson.D{{org + "_code", airportCode}}
	err := coll.FindOne(ctx, filter).Decode(&airport)
	if err != nil {
		log.Fatalf("Error finding airport "+airportCode+": %v", err)
	}

	return airport, true
}

func GetAirportsViaCity(city string, client *mongo.Client, ctx context.Context) []airportInfo {
	coll := client.Database("airports").Collection("airports")

	filter := bson.D{{"city", strings.ToUpper(city)}}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatalf("Error finding airports via city: %v", err)
	}
	defer cursor.Close(ctx)

	var airports []airportInfo

	for cursor.Next(ctx) {
		var airport airportInfo
		if err := cursor.Decode(&airport); err != nil {
			log.Fatalf("Error decoding document: %v", err)
		}
		airports = append(airports, airport)
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Error iterating cursor: %v", err)
	}

	return airports
}