package routeplanner

import (
	//"log"
	//"strings"
	"github.com/notjishy/evrnflightfinder"
	//"go.mongodb.org/mongo-driver/bson"
)

func FindFlights(start string, destination string, credentials string) ([]flightfinder.FlightInfo, []flightfinder.FlightInfo) {
	startAirports := flightfinder.GetAirportsViaCity(start, credentials)
	endAirports := flightfinder.GetAirportsViaCity(destination, credentials)

	directFlights := flightfinder.FindDirect(startAirports, endAirports, credentials)

	nonDirectFlights := flightfinder.FindConnections(startAirports, endAirports, credentials)

	return directFlights, nonDirectFlights
}
