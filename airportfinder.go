package flightfinder

import (
	"log"
	"strings"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	//"fmt"
)

type airportInfo struct {
	ID        int32           `bson:"_id"`
	ICAO      string          `bson:"icao_code"`
	IATA      string          `bson:"iata_code"`
	Name      string          `bson:"name"`
	City      string          `bson:"city"`
	Country   string          `bson:"country"`
	Latitude  float64          `bson:"lat_decimal"`
	Longitude float64          `bson:"lon_decimal"`
}

var airport airportInfo

func GetAirportViaCode(airportCode string, org string, client *mongo.Client, ctx context.Context) (airportInfo, bool) {
	if airportCode == "" { return airport, false }
	airportCode = strings.ToUpper(airportCode)

	coll := client.Database("airports").Collection("airports")

	filter := bson.D{{org + "_code", airportCode}}
	err := coll.FindOne(ctx, filter).Decode(&airport)
	if err != nil { log.Fatalf("Error finding airport " + airportCode + ": %v", err) }

	return airport, true
}

func GetAirportsViaCity(city string, client *mongo.Client, ctx context.Context) []airportInfo {
	coll := client.Database("airports").Collection("airports")

	filter := bson.D{{"city", strings.ToUpper(city)}}

	cursor, err := coll.Find(ctx, filter)
	if err != nil { log.Fatalf("Error finding airports via city: %v", err) }
	defer cursor.Close(ctx)

	var airports []airportInfo

	for cursor.Next(ctx) {
		var airport airportInfo
		if err := cursor.Decode(&airport); err != nil {
			log.Fatalf("Error decoding document: %v", err)
		}
		airports = append(airports,  airport)
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Error iterating cursor: %v", err)
	}

	return airports
}

func IsHub(airports []airportInfo, dbname string, collections []string) (bool, airportInfo) {
	var isHub, stop bool = false, false
	var foundAirport airportInfo
	for _, coll := range collections {
		for _, airport := range airports {
			if strings.ToUpper(coll) == airport.ICAO {
				foundAirport = airport
				stop = true
				isHub = true
				break
			}
		}
		if stop { break }
	}

	return isHub, foundAirport
}
