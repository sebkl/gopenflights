package gopenflights

import(
	"testing"
)

var db *Database
var jfk int

func TestInitialize(t *testing.T) {
	db = NewDatabase()
	t.Logf("Record count: %d",(len(db.Routes) + len(db.Airports) + len(db.Airlines)))
}

func TestAirportsByIATA( t *testing.T) {
	jfk = db.AirportsByIATA["JFK"].Id
	if db.Airport(jfk).City != "New York" {
		t.Errorf("JFK is in New York, nothing else :-/")
	}
}
func TestRoutesByAirport( t *testing.T) {
	all := db.RoutesByAirport(jfk)

	if len(all) < 1 {
		t.Errorf("No routes found for JFK.")
	}

	to := db.RoutesToAirport(jfk)
	from := db.RoutesFromAirport(jfk)


	lall := len(all)
	lfrom := len(from)
	lto := len(to)

	if lall != (lfrom + lto) {
		t.Errorf("Count of incoming and outgoing routes does not match all routes: %d + %d ~ %d",lfrom,lto,lall)
	}
	t.Logf("AirportId[%d] Incoming: %d, Outgoing: %d, Total: %d",jfk,lto,lfrom,lall)
}

