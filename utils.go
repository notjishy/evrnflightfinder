package flightfinder

import (
	"log"

	"github.com/jftuga/geodist"
	"go.mongodb.org/mongo-driver/mongo"
	"context"
)

func getFlightDistance(flight FlightInfo, client *mongo.Client, ctx context.Context) float64 {
	start, success := GetAirportViaCode(flight.Start, "icao", client, ctx)
	if !success {
		log.Fatalf("Error getting airport information: " + flight.Start)
	}
	end, success := GetAirportViaCode(flight.Destination, "icao", client, ctx)
	if !success {
		log.Fatalf("Error getting airport information: " + flight.Destination)
	}

	var startLoc = geodist.Coord{Lat: start.Latitude, Lon: start.Longitude}
	var endLoc = geodist.Coord{Lat: end.Latitude, Lon: end.Longitude}
	_, km, err := geodist.VincentyDistance(startLoc, endLoc)
	if err != nil {
		log.Fatalf("Error calculating base distance! %v", err)
	}

	return km
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
		if stop {
			break
		}
	}

	return isHub, foundAirport
}