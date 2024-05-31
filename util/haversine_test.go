package util

import (
	"math"
	"testing"
)

type haversineTestCase struct {
    lat1, lon1, lat2, lon2 float64
    expected              float64
}

func TestHaversine(t *testing.T) {
    testCases := []haversineTestCase{
        {52.2296756, 21.0122287, 41.8919300, 12.5113300, 1318.0},  // Warsaw to Rome
		{56.764768, -111.359646, 46.875213, -100.701767, 1318.01}, // Bismarck to Fort McMurray
		{61.259669, -149.939319, -33.833920, 151.062189, 11832.04}, // Acnhorage to Sidney
		{22.573438, 114.131640, 69.660905, 18.929955, 7859.32}, // Hong Kong to Tromso
		{38.693013, -90.257504, -0.641314, -90.398045, 4373.81}, // Saint Louise to Galapagos (is included by airport)
		{47.478928, -111.371140, 49.868640, -111.377673, 265.72}, // Great Falls to Bow Island (is included by airport)
    }

	threshold := 5.0

	for _, tc := range testCases {
        result := Haversine(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
        if math.Abs(result-tc.expected) > threshold {
            t.Errorf("Haversine Failed(%f, %f, %f, %f) = %f; want %f", tc.lat1, tc.lon1, tc.lat2, tc.lon2, result, tc.expected)
        } else {
			t.Logf("Haversine Succeded(%f, %f, %f, %f) = %f; want %f", tc.lat1, tc.lon1, tc.lat2, tc.lon2, result, tc.expected)
		}
    }
}