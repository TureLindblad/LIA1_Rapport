package util

import (
	"math"
)

var BuildFromFileMode bool
var RunHeavy bool

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

type Coordinate struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
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