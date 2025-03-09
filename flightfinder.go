package flightfinder

import (
	"context"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jftuga/geodist"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var DBs = [...]string{"evrnair", "flaxair"}
var orgSearchMethod string
var airlinedb string
var doc airlineInfo
var flight FlightInfo
var aircraft aircraftInfo

func ViaFlightNum(flightNum string, client *mongo.Client, ctx context.Context) (FlightInfo, airlineInfo, aircraftInfo, string) {
	num := regexp.MustCompile(`\d`).MatchString(flightNum) // confirm number is included
	if !num {
		log.Fatal("Invalid flight number!")
	} // stop if no number

	// must be seperate filters, one OR the other must be found
	filter := bson.D{{"ICAO", strings.ToUpper(string(flightNum[0:3]))}}
	filter2 := bson.D{{"IATA", strings.ToUpper(string(flightNum[0:2]))}}

	// iterate over databases defined in DBs array above
	var stop bool
	for i, dbname := range DBs {
		coll := client.Database(dbname).Collection("info") // enter airline info collection

		// find document with both filters
		// found documents are assined to the "doc" variable defined above
		var err error
		for j := 1; j <= 2; j++ {
			if j == 1 {
				err = coll.FindOne(ctx, filter).Decode(&doc)
				orgSearchMethod = "ICAO"
			} else {
				err = coll.FindOne(ctx, filter2).Decode(&doc)
				orgSearchMethod = "IATA"
			}

			if err != nil {
				if err == mongo.ErrNoDocuments {
					// do not stop looping until everything has been searched
					if i == len(DBs) {
						log.Fatal(err)
					}
				}
			} else {
				stop = true
				break // break out of j loop if document has been found
			}
		}
		if stop {
			airlinedb = dbname
			break // break out of i loop as well if document found
		}
	}

	// get flight number by itself
	switch orgSearchMethod {
	case "ICAO":
		flightNum = string(flightNum[3:len(flightNum)])
		break
	case "IATA":
		flightNum = string(flightNum[2:len(flightNum)])
		break
	}
	// ensure flight number is integer datatype
	v, err := strconv.Atoi(flightNum) // convert to integer
	if err != nil {
		log.Fatalf("there was an error converting to integer: %v", err)
	}

	collections, err := client.Database(airlinedb).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatalf("there was an error grabbing collection names: %v", err)
	}

	var coll string
	var i int = 0
	for i, coll = range collections {
		collection := client.Database(airlinedb).Collection(coll)

		filter := bson.D{{"flightNum", v}, {"isActive", true}}
		err = collection.FindOne(ctx, filter).Decode(&flight)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				if i+1 == len(collections) {
					log.Fatalf("No flights found. %v", err)
				}
			}
		} else {
			break
		}
	}
	flight.Airline = airlinedb

	if flight.Start == "" || flight.Destination == "" {
		if flight.IsReturn == false {
			flight.Start = strings.ToUpper(coll)
			flight.Destination = flight.Airport
		} else if flight.IsReturn == true {
			flight.Start = flight.Airport
			flight.Destination = strings.ToUpper(coll)
		}
	}

	// get flight distance
	flight.Distance = getFlightDistance(flight, client, ctx)

	var aircraftType string
	if doc.Aircraft == "" {
		if len(flight.AllowedAircraftTypes) >= 1 {
			rand.Seed(time.Now().UnixNano())
			randomIndex := rand.Intn(len(flight.AllowedAircraftTypes))
			aircraftType = flight.AllowedAircraftTypes[randomIndex]
		} else {
			aircraftType = flight.AllowedAircraftTypes[0]
		}
	} else {
		aircraftType = doc.Aircraft
	}

	collection := client.Database(airlinedb).Collection("fleet")

	filter = bson.D{{"type", aircraftType}}

	err = collection.FindOne(ctx, filter).Decode(&aircraft)
	if err != nil {
		log.Fatalf("Error acquiring aircraft information: %v", err)
	}

	return flight, doc, aircraft, flightNum
}

// get direct flights via start and end cities
func FindDirect(startAirports []airportInfo, endAirports []airportInfo, client *mongo.Client, ctx context.Context) []FlightInfo {
	var flights []FlightInfo

	var startIsHub, endIsHub bool = false, false
	var startAirport, endAirport airportInfo
	for _, dbname := range DBs {
		collections, err := client.Database(dbname).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			log.Fatalf("there was an error grabbing collection names: %v", err)
		}

		startIsHub, startAirport = IsHub(startAirports, dbname, collections)
		endIsHub, endAirport = IsHub(endAirports, dbname, collections)

		// handles is one or both cities are a hub
		if startIsHub {
			flights = getHubFlightViaAirports(flights, client, ctx, dbname, startAirport, endAirports, false)
		}
		if endIsHub {
			flights = getHubFlightViaAirports(flights, client, ctx, dbname, endAirport, startAirports, true)
		}

		// find non-hub direct flights
		flights = getNonHubFlightViaAirports(flights, client, ctx, dbname, startAirports, endAirports)
	}

	return flights
}

// get non-direct flights via start and end cities, will include a connection in returned flights
func FindConnections(startAirports []airportInfo, endAirports []airportInfo, client *mongo.Client, ctx context.Context) []FlightInfo {
	var flights []FlightInfo
	var startFlights []FlightInfo
	var endFlights []FlightInfo

	// used for sperating groups
	var breakFlight FlightInfo
	breakFlight.Notes = "BREAK"

	// get base distance
	var startLoc = geodist.Coord{Lat: startAirports[0].Latitude, Lon: startAirports[0].Longitude}
	var endLoc = geodist.Coord{Lat: endAirports[0].Latitude, Lon: endAirports[0].Longitude}
	_, baseKm, err := geodist.VincentyDistance(startLoc, endLoc)
	if err != nil {
		log.Fatalf("Error calculating base distance! %v", err)
	}

	for _, dbname := range DBs {
		collections, err := client.Database(dbname).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			log.Fatalf("there was an error grabbing collection names: %v", err)
		}

		for _, collection := range collections {
			coll := client.Database(dbname).Collection(collection)

			for _, startAirport := range startAirports {
				filter1 := bson.D{{"start", startAirport.ICAO}, {"isActive", true}}
				filter2 := bson.D{}
				if collection == strings.ToLower(startAirport.ICAO) {
					filter2 = bson.D{{"isActive", true}, {"isReturn", false}}
				} else {
					filter2 = bson.D{{"airport", startAirport.ICAO}, {"isActive", true}, {"isReturn", true}}
				}
				var flightFilters = []bson.D{filter1, filter2}

				for _, filter := range flightFilters {
					cursor, err := coll.Find(ctx, filter)
					if err != nil {
						log.Fatalf("Error finding flights in filters: %v", err)
					}
					defer cursor.Close(ctx)

					for cursor.Next(ctx) {
						var flight FlightInfo
						if err := cursor.Decode(&flight); err != nil {
							log.Fatalf("Error decoding documents: %v", err)
						}
						flight.Airline = dbname
						if flight.Start == "" && flight.Destination == "" {
							if flight.IsReturn == true {
								flight.Destination = strings.ToUpper(collection)
								flight.Start = flight.Airport
							} else {
								flight.Start = strings.ToUpper(collection)
								flight.Destination = flight.Airport
							}
						}

						flight.Distance = getFlightDistance(flight, client, ctx)

						startFlights = append(startFlights, flight)
					}
				}
			}

			for _, endAirport := range endAirports {
				filter1 := bson.D{{"destination", endAirport.ICAO}, {"isActive", true}}
				filter2 := bson.D{}
				if collection == strings.ToLower(endAirport.ICAO) {
					filter2 = bson.D{{"isActive", true}, {"isReturn", true}}
				} else {
					filter2 = bson.D{{"airport", endAirport.ICAO}, {"isActive", true}, {"isReturn", false}}
				}
				var flightFilters = []bson.D{filter1, filter2}

				for _, filter := range flightFilters {
					cursor, err := coll.Find(ctx, filter)
					if err != nil {
						log.Fatalf("Error finding flights in filters: %v", err)
					}
					defer cursor.Close(ctx)

					for cursor.Next(ctx) {
						var flight FlightInfo
						if err := cursor.Decode(&flight); err != nil {
							log.Fatalf("Error decoding documents: %v", err)
						}
						flight.Airline = dbname
						if flight.Start == "" && flight.Destination == "" {
							if flight.IsReturn == true {
								flight.Destination = strings.ToUpper(collection)
								flight.Start = flight.Airport
							} else {
								flight.Start = strings.ToUpper(collection)
								flight.Destination = flight.Airport
							}
						}

						flight.Distance = getFlightDistance(flight, client, ctx)

						endFlights = append(endFlights, flight)
					}
				}
			}
		}
	}

	for _, startFlight := range startFlights {
		startFlightAppended := false
		for _, endFlight := range endFlights {
			if endFlight.Start == startFlight.Destination {
				connectionAirport, success := GetAirportViaCode(startFlight.Destination, "icao", client, ctx)
				if !success {
					log.Fatal("Error getting airport from code!")
				}

				startAirport, success := GetAirportViaCode(startFlight.Start, "icao", client, ctx)
				if !success {
					log.Fatal("Error getting airport from code!")
				}
				endAirport, success := GetAirportViaCode(endFlight.Destination, "icao", client, ctx)
				if !success {
					log.Fatal("Error getting airport from code!")
				}

				// prevent international connections for a domestic route
				var validateFlight bool = true
				if startAirport.Country == endAirport.Country {
					if startAirport.Country != connectionAirport.Country {
						validateFlight = false
					}
				}

				if validateFlight {
					if endFlight.Distance <= (baseKm * 1.1) {
						if !startFlightAppended {
							flights = append(flights, startFlight)
							startFlightAppended = true
						}
						flights = append(flights, endFlight)
					}
				}

				// for organization when looping through results
				breakFlight.Distance = startFlight.Distance + endFlight.Distance
			}
		}
		flights = append(flights, breakFlight)
	}

	return flights
}


func getHubFlightViaAirports(flights []FlightInfo, client *mongo.Client, ctx context.Context, dbname string, startAirport airportInfo, endAirports []airportInfo, isReturn bool) []FlightInfo {


	coll := client.Database(dbname).Collection(strings.ToLower(startAirport.ICAO))





	for _, endAirport := range endAirports {
		filter := bson.D{{"airport", endAirport.ICAO}, {"isActive", true}, {"isReturn", isReturn}}
		cursor, err := coll.Find(ctx, filter)
		if err != nil {
			log.Fatalf("Error finding flights: %v", err)
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var flight FlightInfo
			if err := cursor.Decode(&flight); err != nil {
				log.Fatalf("Error decoding document: %v", err)
			}
			flight.Airline = dbname
			if flight.Start == "" && flight.Destination == "" {
				if flight.IsReturn == true {
					flight.Destination = strings.ToUpper(strings.ToLower(startAirport.ICAO))

					flight.Start = flight.Airport
				} else {
					flight.Start = strings.ToUpper(strings.ToLower(startAirport.ICAO))

					flight.Destination = flight.Airport
				}
			}
			flight.Distance = getFlightDistance(flight, client, ctx)
			flights = append(flights, flight)
		}
		if err := cursor.Err(); err != nil {
			log.Fatalf("Error iterating cursor: %v", err)
		}
	}
	return flights
}

func getNonHubFlightViaAirports(flights []FlightInfo, client *mongo.Client, ctx context.Context, dbname string, startAirports []airportInfo, endAirports []airportInfo) []FlightInfo {
	collectionNames, err := client.Database(dbname).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	// iterate over collections
	for _, collectionName := range collectionNames {
		coll := client.Database(dbname).Collection(collectionName)

		for _, startAirport := range startAirports {
			for _, endAirport := range endAirports {
				filter := bson.D{{"start", startAirport.ICAO}, {"destination", endAirport.ICAO}, {"isActive", true}}

				cursor, err := coll.Find(ctx, filter)
				if err != nil {
					log.Fatalf("Error finding non-hub direct flights: %v", err)
				}
				defer cursor.Close(ctx)

				for cursor.Next(ctx) {
					var flight FlightInfo
				    if err := cursor.Decode(&flight); err != nil {
				        log.Fatalf("Error decoding document: %v", err)
				    }
				    flight.Airline = dbname

				    if flight.Start == "" && flight.Destination == "" {
				        if flight.IsReturn == true {
				            flight.Destination = strings.ToUpper(collectionName)
				            flight.Start = flight.Airport
				        } else {
				            flight.Start = strings.ToUpper(collectionName)
				            flight.Destination = flight.Airport
				        }
				    }

				    flight.Distance = getFlightDistance(flight, client, ctx)

				    flights = append(flights, flight)
				}

				if err := cursor.Err(); err != nil {
				    log.Fatalf("Error iterating cursor: %v", err)
				}
			}
		}
	}

	return flights
}
