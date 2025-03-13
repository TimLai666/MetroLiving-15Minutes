package main

import (
	nearbysearch "MetroLiving-15Minutes/nearby_search"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load(".env")
}

func main() {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_MAPS_API_KEY is required")
	} else {
		log.Printf("GOOGLE_MAPS_API_KEY: %s", apiKey)
	}
	res, err := nearbysearch.NearbySearch(apiKey, nearbysearch.ReqData{
		IncludedTypes:  "restaurant",
		MaxResultCount: 20,
		LocationRestriction: nearbysearch.LocationRestriction{
			Circle: nearbysearch.Circle{
				Center: nearbysearch.Center{
					Latitude:  37.422057,
					Longitude: -122.08427,
				},
				Radius: 1000,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
}
