package routeplanner

import (
	//"log"
	//"strings"
	"github.com/notjishy/evrnflightfinder"
	//"go.mongodb.org/mongo-driver/bson"
	"fmt"
)

var credentials string = "urmom"

func FindFlights(start string, destination string) {
	startAirports := flightfinder.GetAirportsViaCity(start, credentials)
	endAirports := flightfinder.GetAirportsViaCity(destination, credentials)

	flight := flightfinder.FindDirect(startAirports, endAirports, credentials)

	fmt.Println(flight)
}
