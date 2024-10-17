package ipgeolocation

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)
//go:generate mockgen -source=geolocalation.go -destination=mocks/mock.go
const (
	UTC_lat  = 0.0
	UTC_long = 0.0
)
type TimeDiffGetter interface{
	GetTimeDiff(lat, lon float64) (int, error)
}

type TimezoneResponse struct {
	UTCtime  string `json:"converted_time"`
	UserTime string `json:"original_time"`
}

func GetTimeDiff(lat, lon float64) (int, error) {
	var tzResponse TimezoneResponse
	url := fmt.Sprintf("https://api.ipgeolocation.io/timezone/convert?apiKey=%s&lat_from=%f&long_from=%f&lat_to=%f&long_to=%f",
		os.Getenv("TIMEZONE_API"), lat, lon, UTC_lat, UTC_long)
	log.Println("URL: " + url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error making request: %v", err)
		return 0, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response: %v", err)
		return 0, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("bad response from server: %s", string(body))
		return 0, fmt.Errorf("bad response from server: %s", string(body))
	}

	if err := json.Unmarshal(body, &tzResponse); err != nil {
		log.Printf("error unmarshalling response: %v", err)
		return 0, fmt.Errorf("error unmarshalling response: %v", err)
	}
	fmt.Println(tzResponse)
	layout := "2006-01-02 15:04:05"

	parsedUTC, err := time.Parse(layout, tzResponse.UTCtime)
	if err != nil {
		log.Println(err)
	}
	parsedOriginal, err := time.Parse(layout, tzResponse.UserTime)
	if err != nil {
		log.Println(err)
	}
	log.Println(parsedOriginal, parsedUTC)
	diff := parsedOriginal.Sub(parsedUTC)
	diffhour := int(diff.Hours())
	log.Printf("timediff is %v \n", diff.Hours())
	return diffhour, nil
}
