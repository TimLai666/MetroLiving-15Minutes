package nearby_search

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ReqData struct {
	IncludedTypes       string              `json:"includedTypes"`
	MaxResultCount      int                 `json:"maxResultCount"`
	LocationRestriction LocationRestriction `json:"locationRestriction"`
}

type LocationRestriction struct {
	Circle Circle `json:"circle"`
}

type Circle struct {
	Center Center `json:"center"`
	Radius int    `json:"radius"`
}

type Center struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RespData struct {
	Places []Place `json:"places"`
}

// Place 表示單一地點的資訊
type Place struct {
	Types            []string    `json:"types"`
	FormattedAddress string      `json:"formattedAddress"`
	WebsiteUri       string      `json:"websiteUri"`
	DisplayName      DisplayName `json:"displayName"`
}

// DisplayName 表示顯示名稱資訊
type DisplayName struct {
	LanguageCode string `json:"languageCode"`
	Text         string `json:"text"`
}

func NearbySearch(apiKey string, reqData ReqData) (responseData map[string]any, err error) {
	var jsonData []byte
	jsonData, err = json.Marshal(reqData)
	if err != nil {
		return nil, err
	}
	nearbySearchUrl := "https://places.googleapis.com/v1/places:searchNearby"
	req, err := http.NewRequest("POST", nearbySearchUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "*")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}
