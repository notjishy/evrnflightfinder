package flightfinder

import (
	"log"
	"regexp"
	"strings"
	"strconv"
	"math/rand"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"fmt"
)

type airlineInfo struct {
	ID   primitive.ObjectID `bson:"_id"`
	ICAO string             `bson:"ICAO"`
	IATA string             `bson:"IATA"`
	Name string             `bson:"name"`
	Aircraft string         `bson:"aircraft`
}

type flightInfo struct {
	ID                   primitive.ObjectID `bson:"_id"`
	FlightNum            int32              `bson:"flightNum"`
	IsReturn             bool               `bson:"isReturn"`
	Start                string             `bson:"start"`
	Stopover             string             `bson:"stopover"`
	Destination          string             `bson:"destination"`
	Airport              string             `bson:"airport"`
	AllowedAircraftTypes []string           `bson:"allowedAircraftTypes"`
	Check                bool               `bson:"check"`
	IsActive             bool               `bson:"isActive"`
	Notes                string             `bson:"notes"`
}

type aircraftInfo struct {
	ID           primitive.ObjectID `bson:"_id"`
	Type         string             `bson:"type"`
	Manufacturer string             `bson:"manufacturer"`
	Model        string             `bson:"model"`
	Liveries     []string           `bson:"liveries"`
}

var DBs = [...]string{"evrnair", "flaxair"}
var orgSearchMethod string
var airlinedb string
var doc airlineInfo
var flight flightInfo
var aircraft aircraftInfo

func ViaFlightNum(flightNum string) (flightInfo, airlineInfo, aircraftInfo, string) {
	num := regexp.MustCompile(`\d`).MatchString(flightNum) // confirm number is included
	if !num { log.Fatal("Invalid flight number!") } // stop if no number
	client := Login() // log into database

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
				err = coll.FindOne(Ctx, filter).Decode(&doc)
				orgSearchMethod = "ICAO"
			} else {
				err = coll.FindOne(Ctx, filter2).Decode(&doc)
				orgSearchMethod = "IATA"
			}

			if err != nil {
				if err == mongo.ErrNoDocuments {
					// do not stop looping until everything has been searched
					if i == len(DBs) { log.Fatal(err) }
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
	if err != nil { log.Fatalf("there was an error converting to integer: %v", err) }

	collections, err := client.Database(airlinedb).ListCollectionNames(Ctx, bson.M{})
	if err != nil { log.Fatalf("there was an error grabbing collection names: %v", err) }

	var coll string
	var i int = 0
	for i, coll = range collections {
		Collection = client.Database(airlinedb).Collection(coll)

		filter := bson.D{{"flightNum", v},{"isActive", true}}

		err = Collection.FindOne(Ctx, filter).Decode(&flight)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				if i+1 == len(collections) { log.Fatalf("No flights found. %v", err) }
			}
		} else { break }
	}

	if flight.Start == "" || flight.Destination == "" {
		if flight.IsReturn == false {
			flight.Start = strings.ToUpper(coll)
			flight.Destination = flight.Airport
		} else if flight.IsReturn == true {
			flight.Start = flight.Airport
			flight.Destination = strings.ToUpper(coll)
		}
	}

	var aircraftType string
	if doc.Aircraft == "" {
		if len(flight.AllowedAircraftTypes) >= 1 {
			rand.Seed(time.Now().UnixNano())
			randomIndex := rand.Intn(len(flight.AllowedAircraftTypes))
			aircraftType = flight.AllowedAircraftTypes[randomIndex]
		} else { aircraftType = flight.AllowedAircraftTypes[0] }
	} else { aircraftType = doc.Aircraft }

	collection := client.Database(airlinedb).Collection("fleet")

	filter = bson.D{{"type", aircraftType}}

	err = collection.FindOne(Ctx, filter).Decode(&aircraft)
	if err != nil { log.Fatalf("Error acquiring aircraft information: %v", err) }

	return flight, doc, aircraft, flightNum
}

// get nonstop flights via start and end cities
func FindDirect(startAirports []airportInfo, endAirports []airportInfo) flightInfo {
	client := Login()
	var flights []flightInfo

	var startIsHub, endIsHub bool = false, false
	var startAirport, endAirport airportInfo
	for _, dbname := range DBs {
		collections, err := client.Database(dbname).ListCollectionNames(Ctx, bson.M{})
		if err != nil { log.Fatalf("there was an error grabbing collection names: %v", err) }

		startIsHub, startAirport = IsHub(startAirports, dbname, collections)
		endIsHub, endAirport = IsHub(endAirports, dbname, collections)

		if startIsHub {
			flights = getHubFlightViaAirports(flights, client, dbname, startAirport, endAirport, false)
		}
		if endIsHub {
			flights = getHubFlightViaAirports(flights, client, dbname, endAirport, startAirport, true)
		}
	}

	var flight flightInfo
	if len(flights) >= 1 {
		rand.Seed(time.Now().UnixNano())
		randomIndex := rand.Intn(len(flights))
		flight = flights[randomIndex] 
	} else { flight = flights[0] }

	return flight
}

func getHubFlightViaAirports(flights []flightInfo, client *mongo.Client, dbname string, startAirport airportInfo, endAirport airportInfo, isReturn bool) ([]flightInfo) {
	coll := client.Database(dbname).Collection(strings.ToLower(startAirport.ICAO))

	filter := bson.D{{"airport", endAirport.ICAO},{"isActive", true},{"isReturn", isReturn}}

	cursor, err := coll.Find(Ctx, filter)
	if err != nil { log.Fatalf("Error finding flights: %v", err) }
	defer cursor.Close(Ctx)

	for cursor.Next(Ctx) {
		var flight flightInfo
		if err := cursor.Decode(&flight); err != nil {
			log.Fatalf("Error decoding document: %v", err)
		}
		flights = append(flights, flight)
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Error iterating cursor: %v", err)
	}

	return flights
}
