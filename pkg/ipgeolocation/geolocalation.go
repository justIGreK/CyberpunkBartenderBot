package ipgeolocation

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	UTC_lat  = 0.0
	UTC_long = 0.0
)

type TimezoneResponse struct {
	UTCtime  string `json:"converted_time"`
	UserTime string `json:"original_time"`
}

func GetTimeDiff(lat, lon float64, reminderTime string) (TimezoneResponse, error) {
	var tzResponse TimezoneResponse
	url := fmt.Sprintf("https://api.ipgeolocation.io/timezone/convert?apiKey=%s&lat_from=%f&long_from=%f&lat_to=%f&long_to=%f&time=%v",
		os.Getenv("TIMEZONE_API"), lat, lon, UTC_lat, UTC_long, reminderTime)
	fmt.Println("URL:" + url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error making request: %v", err)
		return tzResponse, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response: %v", err)
		return tzResponse, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("bad response from server: %s", string(body))
		return tzResponse, fmt.Errorf("bad response from server: %s", string(body))
	}

	if err := json.Unmarshal(body, &tzResponse); err != nil {
		log.Printf("error unmarshalling response: %v", err)
		return tzResponse, fmt.Errorf("error unmarshalling response: %v", err)
	}
	fmt.Println(tzResponse)
	return tzResponse, nil
}
