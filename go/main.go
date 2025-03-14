package main

import (
	"MetroLiving-15Minutes/geocoding"
	nbs "MetroLiving-15Minutes/nearby_search"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/HazelnutParadise/insyra/isr"
	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

func init() {
	godotenv.Load(".env")
}

type PlaceWithCoordinate struct {
	Place nbs.Place `json:"place"`
	Lat   float64   `json:"lat"`
	Lng   float64   `json:"lng"`
}

type PlaceFlatten struct {
	Types            string  `json:"types"`
	FormattedAddress string  `json:"formattedAddress"`
	WebsiteUri       string  `json:"websiteUri"`
	Name             string  `json:"name"`
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
}

var POI_TYPE_TO_GET = []string{
	"accounting", "airport", "amusement_park", "aquarium", "art_gallery", "atm", "bakery", "bank",
	"bar", "beauty_salon", "bicycle_store", "book_store", "bowling_alley", "bus_station", "cafe",
	"campground", "car_dealer", "car_rental", "car_repair", "car_wash", "casino", "cemetery", "church",
	"city_hall", "clothing_store", "convenience_store", "courthouse", "dentist", "department_store",
	"doctor", "drugstore", "electrician", "electronics_store", "embassy", "fire_station", "florist",
	"funeral_home", "furniture_store", "gas_station", "gym", "hair_care", "hardware_store", "hindu_temple",
	"home_goods_store", "hospital", "insurance_agency", "jewelry_store", "laundry", "lawyer", "library",
	"light_rail_station", "liquor_store", "local_government_office", "locksmith", "lodging", "meal_delivery",
	"meal_takeaway", "mosque", "movie_rental", "movie_theater", "moving_company", "museum", "night_club",
	"painter", "park", "parking", "pet_store", "pharmacy", "physiotherapist", "plumber", "police",
	"post_office", "primary_school", "real_estate_agency", "restaurant", "roofing_contractor", "rv_park",
	"school", "secondary_school", "shoe_store", "shopping_mall", "spa", "stadium", "storage", "store",
	"subway_station", "supermarket", "synagogue", "taxi_stand", "tourist_attraction", "train_station",
	"transit_station", "travel_agency", "university", "veterinary_care", "zoo",
}

func main() {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_MAPS_API_KEY is required")
	} else {
		log.Printf("GOOGLE_MAPS_API_KEY: %s", apiKey)
	}

	mapsClient, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	_ = mapsClient

	// ** 取得捷運站附近poi **
	// getNearbyPOI(apiKey)

	// ** 取得地址座標 **
	// getPOICoordinate(mapsClient)

	// ** 將poi_nearby_per_station_exit資料夾下的json檔案進行扁平化 **
	flattenPOIJson()
}

func getNearbyPOI(apiKey string) {
	var poiNearbyMap = make(map[string]nbs.RespData)
	dt := isr.DT{}.From(isr.CSV{FilePath: path.Join("..", "data", "臺北捷運車站出入口座標.csv"), LoadOpts: isr.CSV_inOpts{FirstRow2ColNames: true}})
	rowCount, _ := dt.Size()
	for i := range rowCount {
		lat := dt.At(i, isr.Name("緯度")).(float64)
		lng := dt.At(i, isr.Name("經度")).(float64)
		res, err := nbs.NearbySearch(apiKey, nbs.ReqData{
			IncludedTypes:  "restaurant",
			MaxResultCount: 20,
			LocationRestriction: nbs.LocationRestriction{
				Circle: nbs.Circle{
					Center: nbs.Center{
						Latitude:  lat,
						Longitude: lng,
					},
					Radius: 1000,
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		poiNearbyMap[dt.At(i, isr.Name("出入口名稱")).(string)] = *res
		log.Println(res)
	}
	b, err := json.Marshal(poiNearbyMap)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile(path.Join("..", "data", "poi_nearby.json"), b, 0644)
}

func getPOICoordinate(mapsClient *maps.Client) {
	poiNearbyJson, err := os.ReadFile(path.Join("..", "data", "poi_nearby.json"))
	if err != nil {
		log.Fatal(err)
	}
	var poiNearbyMap = make(map[string]nbs.RespData)
	json.Unmarshal(poiNearbyJson, &poiNearbyMap)

	var placeWithCoordinateMap = make(map[string][]PlaceWithCoordinate)

	for k, v := range poiNearbyMap {
		placeWithCoordinateMap[k] = []PlaceWithCoordinate{}
		for _, p := range v.Places {
			lat, lng, err := geocoding.GetCoordinate(mapsClient, p.FormattedAddress)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s, %s: %f, %f\n", k, p.FormattedAddress, lat, lng)
			placeWithCoordinateMap[k] = append(placeWithCoordinateMap[k], PlaceWithCoordinate{
				Place: p,
				Lat:   lat,
				Lng:   lng,
			})
		}
		jsonData, err := json.Marshal(placeWithCoordinateMap[k])
		if err != nil {
			log.Fatal(err)
		}
		filename := strings.ReplaceAll(k, "/", "_")
		err = os.WriteFile(path.Join("..", "data", "poi_nearby_per_station_exit", filename+".json"), jsonData, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func flattenPOIJson() {
	dir, err := os.ReadDir(path.Join("..", "data", "poi_nearby_per_station_exit"))
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if !strings.HasSuffix(filename, ".json") {
			continue
		}
		file, err := os.ReadFile(path.Join("..", "data", "poi_nearby_per_station_exit", filename))
		if err != nil {
			log.Fatal(err)
		}
		var placeWithCoordinateMap = make([]PlaceWithCoordinate, 0)
		var placeWithCoordinateMapFlatten = make([]PlaceFlatten, 0)
		json.Unmarshal(file, &placeWithCoordinateMap)

		for _, p := range placeWithCoordinateMap {
			pf := PlaceFlatten{
				Types:            strings.Join(p.Place.Types, ","),
				FormattedAddress: p.Place.FormattedAddress,
				WebsiteUri:       p.Place.WebsiteUri,
				Name:             p.Place.DisplayName.Text,
				Lat:              p.Lat,
				Lng:              p.Lng,
			}
			placeWithCoordinateMapFlatten = append(placeWithCoordinateMapFlatten, pf)
		}

		jsonData, err := json.Marshal(placeWithCoordinateMapFlatten)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(path.Join("..", "data", "poi_nearby_per_station_exit_flatten", filename), jsonData, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
