package request

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func Request(url string) (string, error) {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("url", url)
		return url, errors.New("Get() response error")
	}

	// close once body is returned
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("couldn't read body")
	}

	data := string(body)

	return data, nil
}

// Make requests to Moralis API
func APIRequest(url string) (string, error) {

	requestUrl := os.Getenv("MORALIS_API_URL") + url

	// Create GET request
	req, _ := http.NewRequest("GET", requestUrl, nil)

	// Set headers
	req.Header = http.Header{
		"x-api-key": []string{os.Getenv("MORALIS_API_KEY")},
	}

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return "", errors.New("Request failed")
	}

	// Format body
	defer resp.Body.Close()
	resBody, _ := ioutil.ReadAll(resp.Body)
	response := string(resBody)

	fmt.Println("resp", resp.StatusCode)

	// if Moralis API request fails, send Moralis's response 
	if (resp.StatusCode != 200) {
		errorMessage := "ERROR: Moralis API request failed for " + requestUrl

		log.Println(errorMessage)
		return response, errors.New(errorMessage)
	}

	return response, nil
}

// Make requests to Moralis Solana API
func SolanaAPIRequest(url string) (string, error) {
	req, _ := http.NewRequest("GET", os.Getenv("MORALIS_SOLANA_API_URL") + url, nil)

	// Set headers
	req.Header = http.Header{
		"x-api-key": []string{os.Getenv("MORALIS_API_KEY")},
	}

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// Format body
	defer resp.Body.Close()
	resBody, _ := ioutil.ReadAll(resp.Body)
	response := string(resBody)

	return response, nil
}
