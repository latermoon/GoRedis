/**
 * copy from https://github.com/broady/gogeohash
 */
package geo

import (
	"strings"
)

const (
	precision = 12
	base32    = "0123456789bcdefghjkmnpqrstuvwxyz"
)

func Encode(lat, lng float64) string {
	geohash := ""
	lats := [2]float64{-90, 90}
	lngs := [2]float64{-180, 180}

	even := true
	bit := 0
	n := 0

	for len(geohash) < precision {
		n <<= 1
		// interleave bits
		if even {
			n ^= constrict(&lngs, lng)
		} else {
			n ^= constrict(&lats, lat)
		}

		if bit == 4 {
			geohash += string(base32[n])
			bit = 0
			n = 0
		} else {
			bit++
		}
		even = !even
	}
	return geohash
}

func Decode(geohash string) ([2]float64, [2]float64) {
	lats := [2]float64{-90, 90}
	lngs := [2]float64{-180, 180}
	even := true

	for _, r := range geohash {
		i := strings.Index(base32, string(r))
		for j := 16; j != 0; j >>= 1 {
			if even {
				refine(&lngs, i&j)
			} else {
				refine(&lats, i&j)
			}
			even = !even
		}
	}
	return lats, lngs
}

func constrict(pair *[2]float64, coord float64) int {
	m := mid(pair)
	if coord > m {
		pair[0] = m
		return 1
	}
	pair[1] = m
	return 0
}

func refine(pair *[2]float64, bit int) {
	if bit != 0 {
		pair[0] = mid(pair)
	} else {
		pair[1] = mid(pair)
	}
}

func mid(pair *[2]float64) float64 {
	return (pair[0] + pair[1]) / 2
}
