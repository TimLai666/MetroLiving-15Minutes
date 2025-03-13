package main

import (
	nbs "MetroLiving-15Minutes/nearby_search"
	"log"
	"os"
	"path"

	"github.com/HazelnutParadise/insyra/isr"
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
	dt := isr.DT{}.From(isr.CSV{FilePath: path.Join("..", "data", "臺北捷運車站出入口座標.csv"), LoadOpts: isr.CSV_inOpts{FirstRow2ColNames: true}})
	rowCount, _ := dt.Size()
	for i := 0; i < rowCount; i++ {
		lat := dt.At(i, isr.Name("緯度")).(float64)
		lon := dt.At(i, isr.Name("經度")).(float64)
		res, err := nbs.NearbySearch(apiKey, nbs.ReqData{
			IncludedTypes:  "restaurant",
			MaxResultCount: 20,
			LocationRestriction: nbs.LocationRestriction{
				Circle: nbs.Circle{
					Center: nbs.Center{
						Latitude:  lat,
						Longitude: lon,
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

}
