const airportsURL = "/geojson/airports"
const countriesURL = "/geojson/countries"
const linesURL = "/geojson/lines"
const pointURL = "/geojson/point"

const airports = L.layerGroup([]);
const countries = L.layerGroup([]);
const lines = L.layerGroup([]);
const points = L.layerGroup([]);

const osm = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
});

mapLink =
    '<a href="http://www.esri.com/">Esri</a>';
wholink =
    'i-cubed, USDA, USGS, AEX, GeoEye, Getmapping, Aerogrid, IGN, IGP, UPR-EGP, and the GIS User Community';
const esri = L.tileLayer(
    'http://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
    attribution: '&copy; ' + mapLink + ', ' + wholink,
    maxZoom: 19,
});

var map = L.map('map', {
    center: [0, 0],
    zoom: 3,
    layers: [osm, airports, countries, lines, points]
});

function getGEOJSON(url) {
    fetch(url)
        .then(response => response.json())
        .then((data) => {
            data.features.forEach(feature => {
                const mapFeature = L.geoJson(feature).addTo(map);

                if (url === airportsURL) {
                    mapFeature.bindPopup(`Airport: ${feature.properties.name}, Connections: ${feature.properties.numConnections}, Connected population: ${feature.properties.connectedPopulation}`)
                    airports.addLayer(mapFeature)
                }

                if (url === countriesURL) {
                    const coveragePrecentage = Math.round(parseInt(feature.properties.connectedPopulation) / parseInt(feature.properties.pop_est) * 100)
                    mapFeature.bindPopup(`<p>Country: ${feature.properties.brk_name}</p> 
                    <p>Connected population: ${feature.properties.connectedPopulation}</p>
                    <p>Total population: ${feature.properties.pop_est}</p>
                    <p>Coverage precentage: ${coveragePrecentage}%</p>
                    <p>Number of airports:  ${feature.properties.numberAirports}</p>`)

                    mapFeature.setStyle({ fillColor: getColor(coveragePrecentage), color: getColor(coveragePrecentage) });
                    countries.addLayer(mapFeature)
                }

                if (url === linesURL) {
                    mapFeature.setStyle({ fillColor: "red", color: "red" });
                    lines.addLayer(mapFeature)
                }

                if (url === pointURL) {
                    mapFeature.bindPopup(`Connections: ${feature.properties.numConnections}, Connected population: ${feature.properties.connectedPopulation}`)
                    points.addLayer(mapFeature)
                }
            });
        })
}

function getColor(cov) {
    if (cov > 80) {
        return '#e93e3a';
    } else if (cov > 60) {
        return '#ed683c';
    } else if (cov > 40) {
        return '#f3903f';
    } else if (cov > 20) {
        return '#fdc70c';
    } else if (cov > 0) {
        return '#fff33b';
    } else {
        return 'gray';
    }
}

let placingActive = false

document.querySelector("#toggle").addEventListener('click', (element) => {
    if (placingActive) {
        placingActive = false
    } else {
        placingActive = true
    }

    document.querySelector("#toggle").innerHTML = placingActive ? "Placing: Active" : "Placing: Inactive";
})

function onMapClick(e) {
    if (placingActive) {
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
                points.clearLayers();
            })
            .then(() => {
                getGEOJSON(pointURL)
            })
    }
}

map.on('click', onMapClick);

function getLines(lat, lon) {
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
            lines.clearLayers();
        })
        .then(() => {
            getGEOJSON(linesURL)
        })
}

getGEOJSON(airportsURL)
getGEOJSON(countriesURL)

var baseMaps = {
    "OpenStreetMap": osm,
    "EsriWorldImagery": esri
};

var overlayMaps = {
    "Airports": airports,
    "Countries": countries
};

var layerControl = L.control.layers(baseMaps, overlayMaps).addTo(map);

map.on('popupopen', function (e) {
    if (placingActive) {
        var popup = e.popup;

        getLines(popup.getLatLng().lat.toString(), popup.getLatLng().lng.toString());
    }
});

map.on('popupclose', function () {
    points.clearLayers()
    lines.clearLayers();
});