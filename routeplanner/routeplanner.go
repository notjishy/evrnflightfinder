package routeplanner

import (
	//"log"
	//"strings"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/notjishy/evrnflightfinder"
	//"go.mongodb.org/mongo-driver/bson"
)

func FindFlights(start string, destination string, client *mongo.Client, ctx context.Context) ([]flightfinder.FlightInfo, []flightfinder.FlightInfo) {
	startAirports := flightfinder.GetAirportsViaCity(start, client, ctx)
	endAirports := flightfinder.GetAirportsViaCity(destination, client, ctx)

	directFlights := flightfinder.FindDirect(startAirports, endAirports, client, ctx)

	nonDirectFlights := flightfinder.FindConnections(startAirports, endAirports, client, ctx)

	return directFlights, nonDirectFlights
}
