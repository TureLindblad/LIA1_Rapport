package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"os"

	"github.com/gin-gonic/gin"
)

var cityFeatures FeatureCollection
var countryFeatures FeatureCollection
var airportFeatures FeatureCollection

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

// From https://geojson.xyz/
func buildGEOJSONfromURL(url string, f *FeatureCollection) {
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

func buildGEOJSONfromFile(f *FeatureCollection) {
	// From: https://data.opendatasoft.com/explore/dataset/geonames-all-cities-with-a-population-1000%40public/export/?disjunctive.cou_name_en&location=7,51.6998,12.62878&basemap=jawg.streets
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

type Coordinate struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func processAllDataGEOJSON() {
	var timeCounter int64

	for _, country := range countryFeatures.Features {
		timeCounter++

		if timeCounter % 10 == 0 {
			percentage := float64(timeCounter) / 250.0 * 100
			log.Printf("%v percentage", percentage)
		}

		coordinates, ok := country.Geometry.Coordinates.([]interface{})
		if !ok {
			log.Printf("Error: Invalid coordinates")
			continue
		}
		
		var points [][][]Point

		for _, geom := range coordinates {
			geometry := country.Geometry
			coordinates := geom.([]interface{})
	
			var tmpPoints [][]Point

	
			if geometry.Type == "Polygon" {
				// Handle single Polygon
				var tmpRing []Point
				for _, coord := range coordinates {
					parseCoordinates(coord, &tmpRing)
				}
				tmpPoints = append(tmpPoints, tmpRing)
				points = append(points, tmpPoints)
			} else if geometry.Type == "MultiPolygon" {
				// Handle MultiPolygon
				var tmpRing []Point
				for _, ring := range coordinates {
					for _, coord := range ring.([]interface{}) {
						parseCoordinates(coord, &tmpRing)
					}
					tmpPoints = append(tmpPoints, tmpRing)
				}
			}
	
			points = append(points, tmpPoints)
		}

        var totalConnectedPopulation int64
		var numberAirports int64

        for _, airport := range airportFeatures.Features {


			var numConnections int64
			var connectedPopulation int64

            airportCoordinate := []float64{
                airport.Geometry.Coordinates.([]interface{})[0].(float64),
                airport.Geometry.Coordinates.([]interface{})[1].(float64),
            }

            for _, polygon := range points {
                if pointInPolygon(Point{X: airportCoordinate[0], Y: airportCoordinate[1]}, polygon) {
					numberAirports++
                    for _, city := range cityFeatures.Features {
						// cityCoordinate := []float64{
						// 	city.Geometry.Coordinates.([]interface{})[0].(float64),
						// 	city.Geometry.Coordinates.([]interface{})[1].(float64),
						// }

						if city.Properties["marked"] != "marked" /*&& pointInPolygon(Point{X: cityCoordinate[0], Y: cityCoordinate[1]}, polygon)*/ {
							checkCity(&city, airportCoordinate, &numConnections, &connectedPopulation)
						}
                    }
                    break
                }	
            }

            airport.Properties["numConnections"] = numConnections
			airport.Properties["connectedPopulation"] = connectedPopulation

			totalConnectedPopulation += connectedPopulation
        }

		country.Properties["connectedPopulation"] = totalConnectedPopulation
		country.Properties["numberAirports"] = numberAirports
	}
}

func parseCoordinates(coord interface{}, points *[]Point) {
    switch c := coord.(type) {
    case []interface{}:
        if len(c) >= 2 {
            x, okX := c[0].(float64)
            y, okY := c[1].(float64)
            if okX && okY {
                *points = append(*points, Point{X: x, Y: y})
            }
        }
    }
}

// From https://www.geeksforgeeks.org/how-to-check-if-a-given-point-lies-inside-a-polygon/
func pointInPolygon(point Point, polygon [][]Point) bool {
	var checks []bool

	inside := false

	for _, ring := range polygon {
		if len(ring) > 0 {
			numVertices := len(ring)
			x, y := point.X, point.Y

			p1 := ring[0]
			var p2 Point

			for i := 1; i <= numVertices; i++ {
				p2 = ring[i%numVertices]

				if y > math.Min(p1.Y, p2.Y) {
					if y <= math.Max(p1.Y, p2.Y) {
						if x <= math.Max(p1.X, p2.X) {
							xIntersection := (y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y) + p1.X

							if p1.X == p2.X || x <= xIntersection {
								inside = !inside
							}
						}
					}
				}

				p1 = p2
			}
		}

		checks = append(checks, inside)
		inside = false
	}

	for _, b := range checks {
		if b {
			return true
		}
	}
	return false
}

func processPointGEOJSON(c *gin.Context) {
	var coordinate Coordinate
	if err := c.BindJSON(&coordinate); err != nil {
		log.Print("Error binding")
		return
	}

	lat, _ := strconv.ParseFloat(coordinate.Lat, 64)
	lon, _ := strconv.ParseFloat(coordinate.Lon, 64)
	startingCoordinate := []float64{lon, lat}
	

	lineFeatures = FeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{},
	}

	var numConnections int64
	var connectedPopulation int64

	for _, feature := range cityFeatures.Features {
		generateConnectingLines(&lineFeatures, &feature, startingCoordinate, &numConnections, &connectedPopulation)
	}

	pointFeature = FeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{
			{
				Type: "Feature",
				Properties: map[string]interface{}{
					"numConnections": numConnections,
					"connectedPopulation": connectedPopulation,
				},
				Geometry: Geometry{
					Type:        "Point",
					Coordinates: startingCoordinate,
				},
			},
		},
	}

	c.JSON(http.StatusOK, lineFeatures)
}

func generateConnectingLines(f *FeatureCollection, feature *Feature, startingCoordinate []float64, numConnections *int64, connectedPopulation *int64) {
	featureCoordinate := []float64{
		feature.Geometry.Coordinates.([]interface{})[0].(float64),
		feature.Geometry.Coordinates.([]interface{})[1].(float64),
	}

	distance := haversine(startingCoordinate[0], startingCoordinate[1], featureCoordinate[0], featureCoordinate[1])

	maxDistanceInKm := 100.0 //make adjustable?
	if distance < maxDistanceInKm {
		*numConnections++
		populationString, ok := feature.Properties["population"].(float64)
		if !ok {
			log.Printf("Error: population property is not a float64 for feature %v", feature.Properties["population"])
		}

		population := int64(populationString)

		// Mark city so its not counted by other airports
		feature.Properties["marked"] = "marked"
	
		*connectedPopulation += population

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

		f.Features = append(f.Features, newFeature...)
	}
}

func checkCity(feature *Feature, startingCoordinate []float64, numConnections *int64, connectedPopulation *int64) {
	featureCoordinate := []float64{
		feature.Geometry.Coordinates.([]interface{})[0].(float64),
		feature.Geometry.Coordinates.([]interface{})[1].(float64),
	}

	distance := haversine(startingCoordinate[0], startingCoordinate[1], featureCoordinate[0], featureCoordinate[1])

	maxDistanceInKm := 100.0 //make adjustable?
	if distance < maxDistanceInKm {
		*numConnections++
		populationString, ok := feature.Properties["population"].(float64)
		if !ok {
			log.Printf("Error: population property is not a float64 for feature %v", feature.Properties["population"])
		}

		population := int64(populationString)

		// Mark city so its not counted by other airports
		feature.Properties["marked"] = "marked"
	
		*connectedPopulation += population
	}
}

func getAirportsGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, airportFeatures)
}

func getCountriesGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, countryFeatures)
}

func getLinesGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, lineFeatures)
}

func getPointGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, pointFeature)
}

func main() {
	buildGEOJSONfromFile(&cityFeatures)
	buildGEOJSONfromURL("https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_50m_admin_0_countries.geojson", &countryFeatures)
	buildGEOJSONfromURL("https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_10m_airports.geojson", &airportFeatures)
	processAllDataGEOJSON()

	r := gin.Default()

    r.LoadHTMLGlob("static/html/*.html")

    r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })
	
	r.GET("/geojson/airports", getAirportsGEOJSON)
	r.GET("/geojson/countries", getCountriesGEOJSON)
	r.GET("/geojson/lines", getLinesGEOJSON)
	r.GET("/geojson/point", getPointGEOJSON)

	r.POST("/process", processPointGEOJSON)
	
	r.Run("localhost:3000")
}