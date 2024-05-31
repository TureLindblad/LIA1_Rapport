package build

import (
	"encoding/json"
	"os"
	"net/http"
	"log"

	"rapport/util"
)

// From https://geojson.xyz/
func BuildGEOJSONfromURL(url string, f *util.FeatureCollection) {
	response, err := http.Get(url)
	if err != nil {
		log.Print("Failed to fetch GeoJSON data")
		return
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&f)
	if err != nil {
		log.Print("Failed to decode GeoJSON data")
		return
	}
}

func BuildGEOJSONfromFile(f *util.FeatureCollection) {
	// From: https://data.opendatasoft.com/	explore/dataset/geonames-all-cities-with-a-population-1000%40public/export/?disjunctive.cou_name_en&location=7,51.6998,12.62878&basemap=jawg.streets
	geoJSONData, err := os.ReadFile("assets/cities-population-1000.geojson")
	if err != nil {
		log.Printf("Error reading GeoJSON file: %v", err)
		return
	}

	err = json.Unmarshal(geoJSONData, &f)
	if err != nil {
		log.Printf("Error unmarshalling GeoJSON: %v", err)
		return
	}
}