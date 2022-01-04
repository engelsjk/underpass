package underpass

import (
	"strings"

	geojson "github.com/paulmach/go.geojson"
)

// Note: We use [1] here instead of [2] because [2] does not seem to
// properly use lowercase JSON struct tags in a Feature type.
// [1] https://github.com/paulmach/orb
// [2] https://github.com/twpayne/go-geom

// decode converts a range of (assumed) GeoJSON strings
// into an interface{} of a []geojson.Feature
func decode(r []interface{}) interface{} {
	o := make([]interface{}, len(r), len(r))
	for i, j := range r {
		if f, err := geojson.UnmarshalFeature([]byte(j.(string))); err != nil {
			o[i] = f
		}
	}
	return o
}

func stringify(r []interface{}) string {
	var sb strings.Builder

	n := 0
	for _, s := range r {
		n += len(s.(string)) + 1
	}
	n -= 1 // remove 1  byte for unused last "," delimiter in GeoJSON Feature array
	n += 2 // add 2 bytes for "[" and "]"
	sb.Grow(n)

	sb.WriteString("[")
	for i, s := range r {
		sb.WriteString(s.(string))
		if i != len(r)-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString("]")
	return sb.String()
}
