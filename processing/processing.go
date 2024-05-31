package processing

import (
	"log"
	"strconv"
	"math"
	"net/http"
	
	"github.com/gin-gonic/gin"

	"rapport/util"
)

var CityFeatures util.FeatureCollection
var CountryFeatures util.FeatureCollection
var AirportFeatures util.FeatureCollection

var LineFeatures util.FeatureCollection
var PointFeature util.FeatureCollection

func ProcessAllDataGEOJSON() {
	var timeCounter int64

	for _, country := range CountryFeatures.Features {
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
		
		var points [][][]util.Point

		for _, geom := range coordinates {
			geometry := country.Geometry
			coordinates := geom.([]interface{})
	
			var tmpPoints [][]util.Point

	
			if geometry.Type == "Polygon" {
				// Handle single Polygon
				var tmpRing []util.Point
				for _, coord := range coordinates {
					parseCoordinates(coord, &tmpRing)
				}
				tmpPoints = append(tmpPoints, tmpRing)
				points = append(points, tmpPoints)
			} else if geometry.Type == "MultiPolygon" {
				// Handle MultiPolygon
				var tmpRing []util.Point
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

        for _, airport := range AirportFeatures.Features {
			var numConnections int64
			var connectedPopulation int64			

            airportCoordinate := []float64{
                airport.Geometry.Coordinates.([]interface{})[0].(float64),
                airport.Geometry.Coordinates.([]interface{})[1].(float64),
            }

            for _, polygon := range points {
                if PointInPolygon(util.Point{X: airportCoordinate[0], Y: airportCoordinate[1]}, polygon) {
					numberAirports++
                    for _, city := range CityFeatures.Features {
						// cityCoordinate := []float64{
						// 	city.Geometry.Coordinates.([]interface{})[0].(float64),
						// 	city.Geometry.Coordinates.([]interface{})[1].(float64),
						// }

						if city.Properties["marked"] != "marked" /*&& pointInPolygon(Point{X: cityCoordinate[0], Y: cityCoordinate[1]}, polygon)*/ {
							CheckCity(&city, airportCoordinate, &numConnections, &connectedPopulation)
						}
                    }
                    break
                }	
            }

			if numConnections != 0 {
            	airport.Properties["numConnections"] = numConnections
			}
			if connectedPopulation != 0 {
				airport.Properties["connectedPopulation"] = connectedPopulation

				totalConnectedPopulation += connectedPopulation
			}
        }

		country.Properties["connectedPopulation"] = totalConnectedPopulation
		country.Properties["numberAirports"] = numberAirports
	}
}

func parseCoordinates(coord interface{}, points *[]util.Point) {
    switch c := coord.(type) {
    case []interface{}:
        if len(c) >= 2 {
            x, okX := c[0].(float64)
            y, okY := c[1].(float64)
            if okX && okY {
                *points = append(*points, util.Point{X: x, Y: y})
            }
        }
    }
}

// From https://www.geeksforgeeks.org/how-to-check-if-a-given-point-lies-inside-a-polygon/
func PointInPolygon(point util.Point, polygon [][]util.Point) bool {
	var checks []bool

	inside := false

	for _, ring := range polygon {
		if len(ring) > 0 {
			numVertices := len(ring)
			x, y := point.X, point.Y

			p1 := ring[0]
			var p2 util.Point

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

func ProcessPointGEOJSON(c *gin.Context) {
	var coordinate util.Coordinate
	if err := c.BindJSON(&coordinate); err != nil {
		log.Print("Error binding")
		return
	}

	lat, _ := strconv.ParseFloat(coordinate.Lat, 64)
	lon, _ := strconv.ParseFloat(coordinate.Lon, 64)
	startingCoordinate := []float64{lon, lat}
	

	LineFeatures = util.FeatureCollection{
		Type: "FeatureCollection",
		Features: []util.Feature{},
	}

	var numConnections int64
	var connectedPopulation int64

	for _, feature := range CityFeatures.Features {
		GenerateConnectingLines(&LineFeatures, &feature, startingCoordinate, &numConnections, &connectedPopulation)
	}

	PointFeature = util.FeatureCollection{
		Type: "FeatureCollection",
		Features: []util.Feature{
			{
				Type: "Feature",
				Properties: map[string]interface{}{
					"numConnections": numConnections,
					"connectedPopulation": connectedPopulation,
				},
				Geometry: util.Geometry{
					Type:        "Point",
					Coordinates: startingCoordinate,
				},
			},
		},
	}

	c.JSON(http.StatusOK, LineFeatures)
}

func GenerateConnectingLines(f *util.FeatureCollection, feature *util.Feature, startingCoordinate []float64, numConnections *int64, connectedPopulation *int64) {
	featureCoordinate := []float64{
		feature.Geometry.Coordinates.([]interface{})[0].(float64),
		feature.Geometry.Coordinates.([]interface{})[1].(float64),
	}

	distance := util.Haversine(startingCoordinate[0], startingCoordinate[1], featureCoordinate[0], featureCoordinate[1])

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

		newFeature :=[]util.Feature{
			{
				Type: "Feature",
				Geometry: util.Geometry{
					Type:        "LineString",
					Coordinates: lineString,
				},
			},
		}

		f.Features = append(f.Features, newFeature...)
	}
}

func CheckCity(feature *util.Feature, startingCoordinate []float64, numConnections *int64, connectedPopulation *int64) {
	featureCoordinate := []float64{
		feature.Geometry.Coordinates.([]interface{})[0].(float64),
		feature.Geometry.Coordinates.([]interface{})[1].(float64),
	}

	distance := util.Haversine(startingCoordinate[0], startingCoordinate[1], featureCoordinate[0], featureCoordinate[1])

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