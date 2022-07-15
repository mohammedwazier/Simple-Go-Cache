package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	fmt.Println("Starting the Project of Go Cache")

	api := NewApi()

	http.HandleFunc("/api", api.handleFunc)

	http.ListenAndServe(fmt.Sprintf(":%v", os.Getenv("PORT")), nil)
}

func (a *API) handleFunc(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	p := fmt.Sprintf("Some Request about: %v\n", q)
	fmt.Println(p)
	data, cache, err := a.GetData(r.Context(), q)
	if err != nil {
		fmt.Print("Error calling data source: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := ResponseApi{
		Cache: cache,
		Data:  data,
	}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Print("Error encoding response: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type API struct {
	cache *redis.Client
}

func NewApi() *API {
	redisAddress := fmt.Sprintf("%s:6379", os.Getenv("REDIS_URL"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	return &API{
		cache: rdb,
	}
}

func (a *API) GetData(ctx context.Context, q string) ([]NominatimData, bool, error) {
	// Simple Caching Process
	val, err := a.cache.Get(ctx, q).Result()
	if err == redis.Nil {
		fmt.Println("Get data From External Datasource")
		//Calling External Datasource
		escaped := url.PathEscape(q)
		address := fmt.Sprintf("https://nominatim.openstreetmap.org/search.php?q=%s&format=jsonv2", escaped)

		resp, err := http.Get(address)
		if err != nil {
			return nil, false, err
		}

		data := make([]NominatimData, 0)

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return nil, false, err
		}

		b, err := json.Marshal(data)
		if err != nil {
			return nil, false, err
		}

		err = a.cache.Set(ctx, q, bytes.NewBuffer(b).Bytes(), time.Second*60).Err()
		if err != nil {
			return nil, false, err
		}
		return data, false, nil
	} else if err != nil {
		fmt.Printf("Error calling redis: %v\n", err)
		return nil, false, err
	} else {
		//Get Cache Data
		fmt.Println("Get data From Cache")
		data := make([]NominatimData, 0)
		err := json.Unmarshal(bytes.NewBufferString(val).Bytes(), &data)
		if err != nil {
			return nil, false, err
		}
		return data, true, nil
	}
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
