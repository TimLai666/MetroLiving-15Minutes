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
	// getCoordinate(mapsClient)
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

func getCoordinate(mapsClient *maps.Client) {
	poiNearbyJson, err := os.ReadFile(path.Join("..", "data", "poi_nearby.json"))
	if err != nil {
		log.Fatal(err)
	}
	var poiNearbyMap = make(map[string]nbs.RespData)
	json.Unmarshal(poiNearbyJson, &poiNearbyMap)

	type PlaceWithCoordinate struct {
		Place nbs.Place `json:"place"`
		Lat   float64   `json:"lat"`
		Lng   float64   `json:"lng"`
	}
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
