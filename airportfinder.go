package flightfinder

import (
	"log"
	"strings"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	//"fmt"
)

type airportInfo struct {
	ID      int32           `bson:"_id"`
	ICAO    string          `bson:"icao_code"`
	IATA    string          `bson:"iata_code"`
	Name    string          `bson:"name"`
	City    string          `bson:"city"`
	Country string          `bson:"country"`
}

var airport airportInfo

func getCollection(credentials string) (*mongo.Collection) {
	client := Login(credentials)
	coll := client.Database("airports").Collection("airports")
	
	return coll
}

func GetAirportViaCode(airportCode string, org string, credentials string) (airportInfo, bool) {
	if airportCode == "" { return airport, false }
	airportCode = strings.ToUpper(airportCode)

	coll := getCollection(credentials)

	filter := bson.D{{org + "_code", airportCode}}
	err := coll.FindOne(Ctx, filter).Decode(&airport)
	if err != nil { log.Fatalf("Error finding airport: %v", err) }

	return airport, true
}

func GetAirportsViaCity(city string, credentials string) []airportInfo {
	coll := getCollection(credentials)

	filter := bson.D{{"city", strings.ToUpper(city)}}

	cursor, err := coll.Find(Ctx, filter)
	if err != nil { log.Fatalf("Error finding airports via city: %v", err) }
	defer cursor.Close(Ctx)

	var airports []airportInfo

	for cursor.Next(Ctx) {
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
