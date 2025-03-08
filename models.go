package flightfinder

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type airlineInfo struct {
	ID       primitive.ObjectID `bson:"_id"`
	ICAO     string             `bson:"ICAO"`
	IATA     string             `bson:"IATA"`
	Name     string             `bson:"name"`
	Aircraft string             `bson:"aircraft"`
}

type FlightInfo struct {
	ID                   primitive.ObjectID `bson:"_id"`
	Airline              string
	FlightNum            int32    `bson:"flightNum"`
	IsReturn             bool     `bson:"isReturn"`
	Start                string   `bson:"start"`
	Stopover             string   `bson:"stopover"`
	Destination          string   `bson:"destination"`
	Airport              string   `bson:"airport"`
	AllowedAircraftTypes []string `bson:"allowedAircraftTypes"`
	Check                bool     `bson:"check"`
	IsActive             bool     `bson:"isActive"`
	Notes                string   `bson:"notes"`
	Distance             float64
}

type aircraftInfo struct {
	ID           primitive.ObjectID `bson:"_id"`
	Type         string             `bson:"type"`
	Manufacturer string             `bson:"manufacturer"`
	Model        string             `bson:"model"`
	Liveries     []string           `bson:"liveries"`
}

type airportInfo struct {
	ID        int32   `bson:"_id"`
	ICAO      string  `bson:"icao_code"`
	IATA      string  `bson:"iata_code"`
	Name      string  `bson:"name"`
	City      string  `bson:"city"`
	Country   string  `bson:"country"`
	Latitude  float64 `bson:"lat_decimal"`
	Longitude float64 `bson:"lon_decimal"`
}