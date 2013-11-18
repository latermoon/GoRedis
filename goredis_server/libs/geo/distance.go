/**
 * copy from https://github.com/kellydunn/golang-geo/blob/master/point.go
 */
package geo

import (
	"math"
)

const (
	EARTH_RADIUS = 6356.7523 // Earth's radius ~= 6,356.7523km
)

// Calculates the Haversine distance between two points.
// Original Implementation from: http://www.movable-type.co.uk/scripts/latlong.html
func GreatCircleDistance(frlat, frlng, tolat, tolng float64) float64 {
	dLat := (tolat - frlat) * (math.Pi / 180.0)
	dLon := (tolng - frlng) * (math.Pi / 180.0)

	lat1 := frlat * (math.Pi / 180.0)
	lat2 := tolat * (math.Pi / 180.0)

	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(lat1) * math.Cos(lat2)

	a := a1 + a2

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EARTH_RADIUS * c
}
