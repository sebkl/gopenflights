package gopenflights_test

import(
	"fmt"
	. "gopenflights"
)

func ExampleDatabase() {
	// Initialize the database with online version of the "airport.dat" 
	// and "routes.dat" csv-files. (from sourceforge/openflights.org)
	db := NewDatabase()

	// Lookup JFK airport
	jfk := db.AirportsByIATA["JFK"]

	// Print the city of JFK airport
	fmt.Printf("JFK is in: %s\n",jfk.City)

	// Lookup Duesseldorf airport
	dus := db.AirportsByIATA["DUS"]

	// Print routes from JFK to DUS of all airlines
	routes := db.RoutesByAirport(dus.Id)
	for _,route := range routes {
		if route.DestAirportId == dus.Id && route.SourceAirportId == jfk.Id {
			fmt.Printf("%s -> %s, %d stops with %s\n",route.SourceAirportP.Name, route.DestAirportP.Name, route.Stops,route.AirlineP.Name)

		}
	}

	//Output:
	//JFK is in: New York
	//John F Kennedy Intl -> Dusseldorf, 0 stops with Air Berlin
	//John F Kennedy Intl -> Dusseldorf, 0 stops with American Airlines
}
