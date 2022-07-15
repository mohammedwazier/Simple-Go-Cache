package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	fmt.Println("Starting the Project of Go Cache")

	http.HandleFunc("/api", handleFunc)

	http.ListenAndServe(":3000", nil)
}

func handleFunc(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	p := fmt.Sprintf("Some Request about: %v\n", q)
	fmt.Println(p)
	data, err := GetData(q)
	if err != nil {
		fmt.Print("Error calling data source: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := ResponseApi{
		Cache: false,
		Data:  data,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Print("Error encoding response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetData(q string) ([]NominatimData, error) {
	escaped := url.PathEscape(q)
	address := fmt.Sprintf("https://nominatim.openstreetmap.org/search.php?q=%s&format=jsonv2", escaped)

	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}

	data := make([]NominatimData, 0)

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type ResponseApi struct {
	Cache bool            `json:"cache"`
	Data  []NominatimData `json:"data"`
}

type NominatimData struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmID       int      `json:"osm_id"`
	Boundingbox []string `json:"boundingbox"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	DisplayName string   `json:"display_name"`
	PlaceRank   int      `json:"place_rank"`
	Category    string   `json:"category"`
	Type        string   `json:"type"`
	Importance  float64  `json:"importance"`
	Icon        string   `json:"icon"`
}
