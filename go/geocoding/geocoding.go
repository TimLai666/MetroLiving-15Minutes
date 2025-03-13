package geocoding

import (
	"context"

	"googlemaps.github.io/maps"
)

func GetCoordinate(c *maps.Client, address string) (lat, lng float64, err error) {
	result, err := c.Geocode(context.Background(), &maps.GeocodingRequest{
		Address: address,
	})
	if err != nil {
		return 0, 0, err
	}

	lat = result[0].Geometry.Location.Lat
	lng = result[0].Geometry.Location.Lng
	return lat, lng, nil
}
