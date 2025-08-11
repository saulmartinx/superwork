package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type geocodeLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type geocodeGeometry struct {
	Location geocodeLocation `json:"location"`
}

type geocodeResult struct {
	Geometry geocodeGeometry `json:"geometry"`
}

type geocodeResponse struct {
	Results []geocodeResult `json:"results"`
}

func geocode(address string) (latitude, longitude float64, err error) {
	key := config.GeocodeAPIKey
	encodedAddress := url.QueryEscape(address)
	uri := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", encodedAddress, key)
	response, err := http.Get(uri)
	if err != nil {
		return 0, 0, err
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, 0, err
	}
	var geo geocodeResponse
	if err := json.Unmarshal(b, &geo); err != nil {
		return 0, 0, err
	}
	if len(geo.Results) > 0 {
		location := geo.Results[0].Geometry.Location
		return location.Lat, location.Lng, nil
	}
	return 0, 0, nil
}
