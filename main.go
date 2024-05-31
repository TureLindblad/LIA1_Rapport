package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rapport/processing"
	"rapport/build"
)

func getAirportsGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, processing.AirportFeatures)
}

func getCountriesGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, processing.CountryFeatures)
}

func getLinesGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, processing.LineFeatures)
}

func getPointGEOJSON(c *gin.Context) {
	c.JSON(http.StatusOK, processing.PointFeature)
}

func main() {
	start := time.Now()

	build.BuildGEOJSONfromFile(&processing.CityFeatures)
	build.BuildGEOJSONfromURL("https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_50m_admin_0_countries.geojson", &processing.CountryFeatures)
	build.BuildGEOJSONfromURL("https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_10m_airports.geojson", &processing.AirportFeatures)
	processing.ProcessAllDataGEOJSON()
	
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

	r.POST("/process", processing.ProcessPointGEOJSON)
	
	duration := time.Since(start)
	log.Printf("Time to start program: %v", duration)

	r.Run("localhost:3000")
}