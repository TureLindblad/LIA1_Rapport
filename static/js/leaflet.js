var map = L.map('map').setView([0, 0], 3);

const citiesURL = "/geojson/cities"
const geographyURL = "/geojson/geography"
const linesURL = "/geojson/lines"
const pointURL = "/geojson/point"

L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
}).addTo(map);

function getGEOJSON(url) {
    fetch(url)
    .then(response => response.json())
    .then((data) => {
        data.features.forEach(feature => {
            const mapFeature = L.geoJson(feature).addTo(map);

            if (url === citiesURL) {
                mapFeature.bindPopup(`${feature.properties.NAME}, Population: ${feature.properties.POP_MAX}`)
            }
            
            if (url === geographyURL) {
                mapFeature.setStyle({fillColor: getColor(feature.properties.featureclass), color: getColor(feature.properties.featureclass)});
            }

            if (url === linesURL) {
                mapFeature.setStyle({fillColor: "red", color: "red"});
            }

            if (url === pointURL) {
                // input = ""

                // feature.properties.connectingCities.forEach(city => {
                //     input += city.toString()
                // })
                mapFeature.bindPopup(`Connections: ${feature.properties.numConnections}, AreaValue: ${feature.properties.areaValue}`)
            }
        });
    })
}

function getColor(type) {
    switch (type) {
        case "Island group":
            return 'green'
        case "Continent":
            return 'yellow'
        case "Range/mtn":
            return 'brown'
        case "Basin":
            return 'blue'
        default:
            return 'gray'
    }
}

function onMapClick(e) {
    const lat = e.latlng.lat.toString();
    const lon = e.latlng.lng.toString();

    const coordinates = {
        "lat": lat,
        "lon": lon
    };

    fetch("/process", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(coordinates)
    })
    .then(() => {
        getGEOJSON(pointURL)
        getGEOJSON(linesURL)
    })
}

map.on('click', onMapClick);

getGEOJSON(citiesURL)
getGEOJSON(geographyURL)