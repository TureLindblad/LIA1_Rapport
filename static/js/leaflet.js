const citiesURL = "/geojson/cities"
const geographyURL = "/geojson/geography"
const linesURL = "/geojson/lines"
const pointURL = "/geojson/point"

const cities = L.layerGroup([]);
const geography = L.layerGroup([]);
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
    attribution: '&copy; '+mapLink+', '+wholink,
    maxZoom: 19,
});

var map = L.map('map', {
    center: [0, 0],
    zoom: 3,
    layers: [osm, cities, geography, lines, points]
});

function getGEOJSON(url) {
    fetch(url)
    .then(response => response.json())
    .then((data) => {
        data.features.forEach(feature => {
            const mapFeature = L.geoJson(feature).addTo(map);

            if (url === citiesURL) {
                mapFeature.bindPopup(`${feature.properties.NAME}, Population: ${feature.properties.POP_MAX}`)
                cities.addLayer(mapFeature)
            }
            
            if (url === geographyURL) {
                mapFeature.setStyle({fillColor: getColor(feature.properties.featureclass), color: getColor(feature.properties.featureclass)});
                geography.addLayer(mapFeature)
            }

            if (url === linesURL) {
                mapFeature.setStyle({fillColor: "red", color: "red"});
                lines.addLayer(mapFeature)
            }

            if (url === pointURL) {
                mapFeature.bindPopup(`Connections: ${feature.properties.numConnections}, AreaValue: ${feature.properties.areaValue}`)
                points.addLayer(mapFeature)
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
        points.clearLayers();
        lines.clearLayers();
    })
    .then(() => {
        getGEOJSON(pointURL)
        getGEOJSON(linesURL)
    })
}

map.on('click', onMapClick);

getGEOJSON(citiesURL)
getGEOJSON(geographyURL)

var baseMaps = {
    "OpenStreetMap": osm,
    "EsriWorldImagery": esri
};

var overlayMaps = {
    "Cities": cities,
    "Geography": geography
};

var layerControl = L.control.layers(baseMaps, overlayMaps).addTo(map);