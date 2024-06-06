package routeplanner

import (
	//"log"
	//"strings"
	"github.com/notjishy/evrnflightfinder"
	//"go.mongodb.org/mongo-driver/bson"
	"fmt"
)

func FindFlights(start string, destination string, credentials string) {
	startAirports := flightfinder.GetAirportsViaCity(start, credentials)
	endAirports := flightfinder.GetAirportsViaCity(destination, credentials)

	directFlights := flightfinder.FindDirect(startAirports, endAirports, credentials)

	fmt.Println("Direct Flights: ")
	for _, flight := range directFlights {
		fmt.Println(flight)
	}
}
