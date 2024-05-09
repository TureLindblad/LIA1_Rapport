package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var cityFeatures FeatureCollection
var geographyFeatures FeatureCollection
var lineFeatures FeatureCollection
var pointFeature FeatureCollection

type FeatureCollection struct {
    Type     string     `json:"type"`
    Features []Feature  `json:"features"`
}

type Feature struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
    Geometry   Geometry               `json:"geometry"`
}

type Geometry struct {
    Type        string      `json:"type"`
    Coordinates interface{} `json:"coordinates"`
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Haversine formula
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad
	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(deltaLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	R := 6371.0 // Earth radius in kilometers
	distance := R * c

	return distance
}

// From https://geojson.xyz/ populated places
func buildGEOJSON(url string, f *FeatureCollection) {
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

type Coordinate struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

func processGEOJSON(c *gin.Context) {
	var coordinate Coordinate
	if err := c.BindJSON(&coordinate); err != nil {
		log.Print("Error binding")
		return
	}

	lat, _ := strconv.ParseFloat(coordinate.Lat, 64)
	lon, _ := strconv.ParseFloat(coordinate.Lon, 64)
	// startingCoordinates := generateStartingCoordinate()
	startingCoordinates := [][]float64{{lon, lat}}
	

	lineFeatures = FeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{},
	}

	var numConnections int16
	var areaValue uint64

	for _, startingCoordinate := range startingCoordinates {
		for _, feature := range cityFeatures.Features {
			featureCoordinate := []float64{
				feature.Geometry.Coordinates.([]interface{})[0].(float64),
				feature.Geometry.Coordinates.([]interface{})[1].(float64),
			}

			distance := haversine(startingCoordinate[0], startingCoordinate[1], featureCoordinate[0], featureCoordinate[1])

			maxDistanceInKm := 400.0
			if distance < maxDistanceInKm {
				numConnections++

				tmp := uint64(0)
				switch v := feature.Properties["POP_MAX"].(type) {
				case uint64:
					tmp = v
				case float64:
					tmp = uint64(v)
				default:
					log.Println("Unexpected type for POP_MAX property")
				}
	
				areaValue += uint64((float64(tmp) / distance))
	
				log.Printf("val: %d", areaValue)

				lineString := [][]float64{startingCoordinate, featureCoordinate}

				newFeature :=[]Feature{
					{
						Type: "Feature",
						Geometry: Geometry{
							Type:        "LineString",
							Coordinates: lineString,
						},
					},
				}

				lineFeatures.Features = append(lineFeatures.Features, newFeature...)
			}
		}

		pointFeature = FeatureCollection{
			Type: "FeatureCollection",
			Features: []Feature{
				{
					Type: "Feature",
					Properties: map[string]interface{}{
						"numConnections": numConnections,
						"areaValue": areaValue,
					},
					Geometry: Geometry{
						Type:        "Point",
						Coordinates: startingCoordinate,
					},
				},
			},
		}
	}

	c.JSON(http.StatusOK, lineFeatures)
}

// func generateStartingCoordinate() [][]float64{
// 	var startingCoordinates [][]float64
// 	numPoints := 10

// 	latStep := 180.0 / math.Sqrt(float64(numPoints))
// 	lonStep := 360.0 / math.Sqrt(float64(numPoints))

// 	for lat := -90.0; lat <= 90.0; lat += latStep {
// 		for lon := -180.0; lon <= 180.0; lon += lonStep {
// 			newCoordinate := []float64{lon, lat}
// 			startingCoordinates = append(startingCoordinates, newCoordinate)
// 		}
// 	}

// 	return startingCoordinates
// }

func getCitiesGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, cityFeatures)
}

func getGeographyGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, geographyFeatures)
}

func getLinesGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, lineFeatures)
}

func getPointGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, pointFeature)
}

func main() {
	buildGEOJSON("https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_50m_populated_places.geojson", &cityFeatures)
	buildGEOJSON("https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_50m_geography_regions_polys.geojson", &geographyFeatures)

	r := gin.Default()

    r.LoadHTMLGlob("static/html/*.html")

    r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })
	
	r.GET("/geojson/cities", getCitiesGEOJSON)
	r.GET("/geojson/geography", getGeographyGEOJSON)
	r.GET("/geojson/lines", getLinesGEOJSON)
	r.GET("/geojson/point", getPointGEOJSON)

	r.POST("/process", processGEOJSON)
	
	r.Run("localhost:3000")
}