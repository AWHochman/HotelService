package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"io/ioutil"
	// "encoding/json"
	"github.com/tidwall/gjson"
	"strconv"
	"sort"
	"os"
	"log"
	"math"
)

const LOCAL bool = false
const DUMMYDATA bool = false

type Hotel struct {
	Name, Locality, Country string
	Price int
	StarRating int64
	Long, Lat float64
	DistanceFromAirport float64
}

func main() {
	router := gin.Default()
	router.GET("/api/query-hotels", queryHotels)
	port := getPort()
	log.Printf("About to listen on %s. Go to https://127.0.0.1%s/", port, port)
	router.Run(port)
}

func getPort() string {
	if LOCAL {
		return ":8081"
	}
    port := ":8080"
    if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
        port = ":" + val
    }
    return port
}

// /queryhotels?location=SAF&budget=500&start=2022-03-26&end=2022-03-27&latitude=51.509865&longitude=-0.118092&people=3
func queryHotels(c *gin.Context) {
	if DUMMYDATA {
		hotels := []Hotel{Hotel{"St. Pancras Renaissance Hotel", "London Euston Road", "United Kingdom", 500, 5, 0.0, 0.0, 0.0},
		Hotel{"St Martins Lane", "London 45 St Martin's Lane", "United Kingdom", 485, 5,  0.0, 0.0, 0.0},
		Hotel{Name:"ME London",Locality:"336-337 The Strand",Country:"United Kingdom",Price:440,StarRating:5},
		{Name:"South Place Hotel",Locality:"3 South Place",Country:"United Kingdom",Price:357,StarRating:5},
		{Name:"Andaz London Liverpool Street - a concept by Hyatt",Locality:"40 Liverpool Street",Country:"United Kingdom",Price:350,StarRating:5},
		}

		sortedHotels := sortResults(hotels)

		c.PureJSON(http.StatusOK, sortedHotels[:min(len(hotels), 10)])
	} else {
		// location := c.DefaultQuery("location", "N/A")
		budget := c.Query("budget") // shortcut for c.Request.URL.Query().Get("budget")
		start := c.Query("start")
		end := c.Query("end")
		longitude := c.Query("longitude")
		latitude := c.Query("latitude")
		people := c.Query("people")

		url := fmt.Sprintf("https://hotels-com-provider.p.rapidapi.com/v1/hotels/nearby?latitude=%v&currency=USD&longitude=%v&checkout_date=%v&sort_order=STAR_RATING_HIGHEST_FIRST&checkin_date=%v&adults_number=%v&locale=en_US&page_number=1&price_min=10&price_max=%v", latitude, longitude, end, start, people, budget)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("X-RapidAPI-Host", "hotels-com-provider.p.rapidapi.com")
		req.Header.Add("X-RapidAPI-Key", "8c1dea45d1mshf68481c4f40bdc2p19580bjsnfeafc8fd25ba")

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatalln(err)
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Fatalln(err)
		}

		hotelRes := gjson.Get(string(body), "searchResults.results").Array()
		hotels := parseQueryResult(hotelRes, longitude, latitude)

		// calculateDistances(hotels, longitude, latitude)

		sortedHotels := sortResults(hotels)
		// log.Printf(fmt.Sprintf("%v", hotels))
		
		// returns the 3 cheapest hotels
		c.PureJSON(http.StatusOK, sortedHotels[0])
		// c.PureJSON(http.StatusOK, sortedHotels[0])
	}
}

// func calculateDistances(hotels []Hotel, airportLong, airportLat string) {
// 	for _, h := range hotels {
// 		lat1, err := strconv.ParseFloat(airportLat, 64)
// 		if err != nil {
// 			panic(err)
// 		}
// 		long1, err := strconv.ParseFloat(airportLong, 64)
// 		if err != nil {
// 			panic(err)
// 		}
// 		h.distanceFromAirport = calcDistance(lat1, long1, h.Lat, h.Long)
// 	}
// }

func calcDistance(lat1, long1, lat2, long2 float64) float64 {
	R := 6371000.0
	phi1 := lat1 * math.Pi/180
	phi2 := lat2 * math.Pi/180
	deltPhi := (lat2 - lat1) * math.Pi/180
	deltLamb := (long2 - long1) * math.Pi/180
	a := math.Sin(deltPhi/2) * math.Sin(deltPhi/2) + math.Cos(phi1) * math.Cos(phi2) * math.Sin(deltLamb/2) * math.Sin(deltLamb/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return (R * c)/1609 //meters to miles 
}

func parseQueryResult(result []gjson.Result, airportLong, airportLat string) []Hotel{
	hotels := make([]Hotel, 0)
	lat1, err := strconv.ParseFloat(airportLat, 64)
	if err != nil {
		panic(err)
	}
	long1, err := strconv.ParseFloat(airportLong, 64)
	if err != nil {
		panic(err)
	}
	for _, v := range result {
		// log.Printf(v.String())
		h := Hotel{}
		h.Name = v.Get("name").String()
		h.StarRating = v.Get("starRating").Int()
		address := v.Get("address")
		h.Locality = address.Get("streetAddress").String()
		h.Country = address.Get("countryName").String()
		price := v.Get("ratePlan.price.current").String()
		h.Price, _ = strconv.Atoi(price[1:])
		h.Long = v.Get("coordinate.lon").Float()
		h.Lat = v.Get("coordinate.lat").Float()
		h.DistanceFromAirport = calcDistance(lat1, long1, h.Lat, h.Long)
		hotels = append(hotels, h)
		// log.Printf(fmt.Sprintf("%v", h))
	}
	return hotels
}

func sortResults(hotels []Hotel) []Hotel{
	sort.SliceStable(hotels, func(i, j int) bool {
		return hotels[i].Price < hotels[j].Price
	})
	return hotels
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
