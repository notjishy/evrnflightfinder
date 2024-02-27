package routeplanner

import (
	//"log"
	//"strings"
	"github.com/notjishy/evrnflightfinder"
	//"go.mongodb.org/mongo-driver/bson"
	"fmt"
)

func FindFlights(start string, destination string) {
	startAirports := flightfinder.GetAirportsViaCity(start)
	endAirports := flightfinder.GetAirportsViaCity(destination)

	flight := flightfinder.FindDirect(startAirports, endAirports)

	fmt.Println(flight)
}
