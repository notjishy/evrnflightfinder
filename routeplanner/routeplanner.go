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

	if directFlights != nil {
		fmt.Println("Direct Flights: ")
		for _, flight := range directFlights {
			fmt.Println(flight)
		}
	} else {
		nonDirectFlights := flightfinder.FindConnections(startAirports, endAirports, credentials)

		fmt.Println("Non-Direct Flights: ")
		var allowBreak bool = false
		for _, flight := range nonDirectFlights {
			if flight.Notes == "BREAK" {
				if allowBreak {
					fmt.Println("")
					allowBreak = false
				}
			} else {
				fmt.Println(flight)
				allowBreak = true
			}
		}
	}
}
