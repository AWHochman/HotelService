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
)

var Local bool = true

type Hotel struct {
	Name, Locality, Country string
	StayPrice int
	StarRating int64
}

func main() {
	router := gin.Default()
	router.GET("/queryhotels", queryHotels)
	router.Run(":8080")
}

// /queryhotels?location=SAF&budget=500&start=2022-03-26&end=2022-03-27&latitude=51.509865&longitude=-0.118092
func queryHotels(c *gin.Context) {
	if Local {
		hotels := []Hotel{Hotel{"St. Pancras Renaissance Hotel", "London Euston Road", "United Kingdom", 500, 5},
		Hotel{"St Martins Lane", "London 45 St Martin's Lane", "United Kingdom", 485, 5},
		Hotel{Name:"ME London",Locality:"336-337 The Strand",Country:"United Kingdom",StayPrice:440,StarRating:5},
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

		url := fmt.Sprintf("https://hotels-com-provider.p.rapidapi.com/v1/hotels/nearby?latitude=%v&currency=USD&longitude=%v&checkout_date=%v&sort_order=STAR_RATING_HIGHEST_FIRST&checkin_date=%v&adults_number=1&locale=en_US&page_number=1&price_min=10&price_max=%v", latitude, longitude, end, start, budget)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("X-RapidAPI-Host", "hotels-com-provider.p.rapidapi.com")
		req.Header.Add("X-RapidAPI-Key", "8c1dea45d1mshf68481c4f40bdc2p19580bjsnfeafc8fd25ba")

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		hotelRes := gjson.Get(string(body), "searchResults.results").Array()
		hotels := parseQueryResult(hotelRes)

		sortedHotels := sortResults(hotels)
		
		c.PureJSON(http.StatusOK, sortedHotels[:min(len(hotels), 10)])
	}
}

func parseQueryResult(result []gjson.Result) []Hotel{
	hotels := make([]Hotel, 0)
	for _, v := range result {
		h := Hotel{}
		h.Name = v.Get("name").String()
		h.StarRating = v.Get("starRating").Int()
		address := v.Get("address")
		h.Locality = address.Get("streetAddress").String()
		h.Country = address.Get("countryName").String()
		price := v.Get("ratePlan.price.current").String()
		h.StayPrice, _ = strconv.Atoi(price[1:])
		hotels = append(hotels, h)
	}
	return hotels
}

func sortResults(hotels []Hotel) []Hotel{
	sort.SliceStable(hotels, func(i, j int) bool {
		return hotels[i].StayPrice < hotels[j].StayPrice
	})
	return hotels
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
