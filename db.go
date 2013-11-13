// Package gopenflights offers a library for easy access to openflights data.
// All data is loaded and cached during initialization from either explicitly specified 
// CSV-files or directly from the openflights webpage (sourceforge)
package gopenflights

import(
	"encoding/csv"
	"fmt"
	"strconv"
	"net/http"
	"io"
	"strings"
	"os"
	"log"
)

const (
	DefaultAirportDatUrl = "http://sourceforge.net/p/openflights/code/HEAD/tree/openflights/data/airports.dat?format=raw"
	DefaultRoutesDatUrl = "http://sourceforge.net/p/openflights/code/HEAD/tree/openflights/data/routes.dat?format=raw"
)

// Database is an openflights database container.
type Database struct {
	Routes []RouteRecord
	Airports []AirportRecord
	AirportsByIdIndex map[int]*AirportRecord
	AirportsByIATA map[string]*AirportRecord
	AirportsByICAO map[string]*AirportRecord
}

type Record interface {
	Convert([]string) error
}


// AirportRecord represents an airport object. 
type AirportRecord struct {
	Id int
	Name,City,Country,IATA,ICAO string
	Lat, Long,Alt float64
	Timezone float64
	DST byte

	// references
	DestRoutes map[*RouteRecord]bool
	SourceRoutes map[*RouteRecord]bool
}

// RouteRecord represents a route object.
type RouteRecord struct {
	Airline string
	AirlineId int
	SourceAirport string
	SourceAirportId int
	DestAirport string
	DestAirportId int
	Codeshare bool
	Stops int
	Equipment string

	//references
	DestAirportP *AirportRecord
	SourceAirportP *AirportRecord
}

// NewDatabase initializes a new openflights database.
// If no parameter are given, the source files are loaded via http from sourceforge. 
// Otherwise first parameter is the "airport.dat" and second the "routes.dat" file.
func NewDatabase(s...string) (db *Database) {
	db = new(Database)
	sl := len(s)

	if sl == 0 {
		db.LoadAirportData(DefaultAirportDatUrl)
		db.LoadRouteData(DefaultRoutesDatUrl)
	} else if sl == 2 {
		db.LoadAirportData(s[0])
		db.LoadRouteData(s[1])
	} else {
		panic("Invalid initialization parameter. Either none or all source files must be specified.")
	}
	return
}

// Convert converts a string array read from the corresponding "routes.dat" csv file into the given RouteRecord object.
func (r *RouteRecord) Convert(s []string) error{
	l := len(s)
	if l < 9 {
		return fmt.Errorf("Invalid field count for Route record: %d/%d",l,9)
	}
	var ret error
	r.Airline = s[0]
	r.AirlineId,ret = strconv.Atoi(s[1])
	r.SourceAirport = s[2]
	r.SourceAirportId,ret = strconv.Atoi(s[3])
	r.DestAirport = s[4]
	r.DestAirportId,ret = strconv.Atoi(s[5])
	csb := []byte(s[6])
	if len(csb) > 0 {
		r.Codeshare = (csb[0] == 'Y')
	} else {
		r.Codeshare = false
	}
	r.Stops,ret = strconv.Atoi(s[7])
	r.Equipment = s[8]
	return ret
}

// Convert converts a string array read from the corresponding "airport.da"t csv file into the given AiportRecord object.
func (r *AirportRecord) Convert(s []string) error{
	l := len(s)
	if l < 11 {
		return fmt.Errorf("Invalid field count for Airport record: %d/%d",l,11)
	}
	var ret error
	r.Id,ret = strconv.Atoi(s[0])
	r.Name = s[1]
	r.City = s[2]
	r.Country = s[3]
	r.IATA = s[4]
	r.ICAO = s[5]
	r.Lat,ret = strconv.ParseFloat(s[6],32)
	r.Long,ret = strconv.ParseFloat(s[7],32)
	r.Alt,ret = strconv.ParseFloat(s[8],32)
	r.Timezone,ret = strconv.ParseFloat(s[9],32)
	r.DST = []byte(s[10])[0]

	r.DestRoutes = make(map[*RouteRecord]bool)
	r.SourceRoutes = make(map[*RouteRecord]bool)
	return ret
}

// loadCsv loads the contents of the given file or http-URL.
func loadCsv(source string) (all [][]string){
	var rc io.ReadCloser
	if strings.HasPrefix(source,"http") {
		resp, err := http.Get(source)
		if err != nil {
			panic(err)
		}
		rc = resp.Body
	} else {
		file, err := os.Open(source)
		if err != nil {
			panic(err)
		}
		rc = file
	}

	reader := csv.NewReader(rc)
	reader.TrailingComma = true
	all,err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Could not read source: %s",err.Error())
	}
	return
}

// LoadAirportData reads the airport data from the given source.
// The source could be either a localfile or http based URL.
func (d *Database) LoadAirportData(source string){
	log.Printf("Loading Airport data.")
	data := loadCsv(source)
	d.Airports =  make([]AirportRecord,len(data))
	d.AirportsByIdIndex = make(map[int]*AirportRecord)
	d.AirportsByIATA = make(map[string]*AirportRecord)
	d.AirportsByICAO = make(map[string]*AirportRecord)
	for i,v := range data {
		err := d.Airports[i].Convert(v)
		if (err != nil) {
			log.Printf("Cannot convert AirportRecord: %s",err.Error())
		} else {
			d.AirportsByIdIndex[d.Airports[i].Id] = &d.Airports[i]
			d.AirportsByIATA[d.Airports[i].IATA] = &d.Airports[i]
			d.AirportsByICAO[d.Airports[i].ICAO] = &d.Airports[i]
			d.Airports[i].DestRoutes = make(map[*RouteRecord]bool)
			d.Airports[i].SourceRoutes = make(map[*RouteRecord]bool)
		}
	}
}

// LoadRouteData reads the route data from the given source.
// The source could be either a localfile or http based URL.
func (d *Database) LoadRouteData(source string) {
	log.Printf("Loading Route data.")
	data := loadCsv(source)
	d.Routes =  make([]RouteRecord,len(data))
	idx := 0
	for i,v := range data {
		err := d.Routes[idx].Convert(v)
		route := &(d.Routes[idx])
		if (err != nil) {
			log.Printf("Cannot convert RouteRecord: %s",err.Error())
		} else if route.DestAirportId == 0 {
			log.Printf("Destination aiportId of \"%s\" @line %d is not specified. Ignoring route.",route.DestAirport,i+1)
		} else if route.SourceAirportId == 0 {
			log.Printf("Source aiportId of \"%s\" @line %d is not specified. Ignoring route.",route.SourceAirport,i+1)
		} else {
			idx++
			route.DestAirportP = d.AirportsByIdIndex[route.DestAirportId]
			route.SourceAirportP = d.AirportsByIdIndex[route.SourceAirportId]
			if route.DestAirportP != nil {
				route.DestAirportP.DestRoutes[route] = true
			} else {
				log.Printf("Could not find destination airportId: %d/%s",route.DestAirportId,route.DestAirport)
			}

			if route.SourceAirportP != nil {
				route.SourceAirportP.SourceRoutes[route] = true
			} else {
				log.Printf("Could not find source airportId: %d/%s",route.SourceAirportId,route.SourceAirport)
			}
		}
	}
}

// keys returns a slice of RouteRecord pointers of the given map.
func keys(m map[*RouteRecord]bool) (ret []*RouteRecord) {
	ret = make([]*RouteRecord,len(m))
	i:= 0
	for rp,_ := range m {
		ret[i] = rp
		i++
	}
	return
}

// Airport returns the AirportRecord of the given airport id.
func (d *Database) Airport(aid int) (*AirportRecord) {
	return d.AirportsByIdIndex[aid]
}

// RoutesToAirport returns all routes to the given airport id.
func (d *Database) RoutesToAirport(aid int) ([]*RouteRecord) {
	return keys(d.AirportsByIdIndex[aid].DestRoutes)
}

// RoutesFromAirport returns all routes from the given airport id.
func (d *Database) RoutesFromAirport(aid int) ([]*RouteRecord) {
	return keys(d.AirportsByIdIndex[aid].SourceRoutes)
}

// RoutesByAirport returns all routes from or to the given airport id.
func (d *Database) RoutesByAirport(aid int) ([]*RouteRecord) {
	result := make(map[*RouteRecord]bool)
	ap := d.AirportsByIdIndex[aid]
	for rp,_ := range ap.DestRoutes {
		result[rp] = true
	}

	for rp,_ := range ap.SourceRoutes {
		result[rp] = true
	}

	return keys(result)
}

