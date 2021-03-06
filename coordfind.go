package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// getResp(url string) []byte
// Returns a byte slice from the body of the http GET
func getResp(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Cannot evaluate the requested url")
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Cannot evaulate response body")
		os.Exit(1)
	}
	jsonString := fmt.Sprintf("%s", string(body))
	return []byte(jsonString)
}

// processPosition returns latitude and longitude as float64
// from a query result
func processPosition(m map[string]interface{}) (float64, float64, bool) {
	l0, ok := m["results"].([]interface{})
	if !ok {
		return 0, 0, false
	}
	l1, ok := l0[0].(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	l2, ok := l1["geometry"].(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	l3, ok := l2["location"].(map[string]interface{})
	if !ok {
		return 0, 0, false
	}
	latitude, ok := l3["lat"].(float64)
	if !ok {
		return 0, 0, false
	}
	longitude, ok := l3["lng"].(float64)
	if !ok {
		return 0, 0, false
	}
	return latitude, longitude, true
}

func main() {

	appName := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Printf("%s retrievs coordinates of a given address from Google Maps API.\n\nUsage: %s ADDRESS\n", appName, appName)
		flag.PrintDefaults()
	}

	flag.Parse()
	rawArgs := flag.Args()
	if len(rawArgs) == 0 {
		fmt.Println("Error: missing location")
		os.Exit(1)
	}
	location := strings.Join(rawArgs, "+")

	apiKey := "AIzaSyDTaE7gzghBlk_7dV-rgurL9yJbx-7IK3E"
	url := "https://maps.googleapis.com/maps/api/geocode/json?address=" + location + "&key=" + apiKey

	// Calling getResp(url) using a goroutine and channel asynchronous communication
	jsonByte := make(chan []byte)

	go func(url string) {
		jsonByte <- getResp(url)
	}(url)

	// Defie a map of string to empty interfaces to
	// host data whose structure is unknown
	m := make(map[string]interface{})

	// json.Unmarshal waits for data from jsonByte channel
	err := json.Unmarshal(<-jsonByte, &m)
	if err != nil {
		fmt.Println("Cannot unmarshal output")
		os.Exit(1)
	}

	// Extract the location fields
	status := m["status"]
	if status == "ZERO_RESULTS" {
		fmt.Println("The search produced zero results")
		os.Exit(1)
	}

	// If twe have found something, we start to
	// unpack the map using type assertion.
	// We want to print out the coordinates in the
	// followint format: (lat, lng).
	lat, lng, ok := processPosition(m)
	if !ok {
		fmt.Println("Error extracting latitude/longitude")
		os.Exit(1)
	}

	// Print results
	fmt.Printf("(%f, %f)\n", lat, lng)
}
