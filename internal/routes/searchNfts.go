package routes

import (
	"api/pkg/ipfsurl"
	"api/pkg/request"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

func SearchNfts(w http.ResponseWriter, r *http.Request) {

	type Result struct {
		Token_Id            string `json:"token_id"`
		Token_Address       string `json:"token_address"`
		Token_Uri           string `json:"token_uri"`
		Metadata            string `json:"metadata"`
		AppMetadata         interface{} `json:"appMetadata"`
		Contract_Type       string `json:"contract_type"`
		Token_Hash          string `json:"token_hash"`
		Minter_Address      string `json:"minter_address"`
		Block_Number_Minted string `json:"block_number_minted"`
	  Transaction_Minted  string `json:"transaction_minted"`
	  Last_Token_Uri_Sync string `json:"last_token_uri_sync"`
	  Last_Metadata_Sync  string `json:"last_metadata_sync"`
	  Created_At          string `json:"created_at"`
	}

	type Response struct {
		Total     int                 `json:"total"`
		Page      int                 `json:"page"`
		Page_Size int                 `json:"page_size"` 
		Cursor    string              `json:"cursor,omitempty"`
		Result    []Result            `json:"result"`
		Data      map[string][]Result `json:"data"`
	}

	// get URL params
	vars := mux.Vars(r)
	chain := vars["chain"]
	q := url.QueryEscape(vars["q"])
	filter := vars["filter"]
	limit := vars["limit"]
	cursor := vars["cursor"]

	if cursor != "" {
		cursor = "&cursor=" + cursor
	}

	// Moralis searchNFTs https://docs.moralis.io/reference/searchnfts
	response, err := request.APIRequest(`/nft/search?chain=` + chain + `&q=` + q + `&filter=` + filter + `&limit=` + limit + cursor)
	if err != nil {
		fmt.Println("Error - ", err)
	}

	var data Response

	err = json.Unmarshal([]byte(response), &data)
	if err != nil {
		fmt.Println("Couldn't unmarshal: ", err)
		return
	}

	var wg sync.WaitGroup // Create WaitGroup to wait for all goroutines to finish

	// Loop through each NFT's results
	for i, nft := range data.Result {

		wg.Add(1)

		// Fetch each NFT's metadata in parallel
		go func(i int, nft Result) {

			// Decrease WaitGroup when goroutine has finished
			defer wg.Done()

			updatedMetadata := ""

			if nft.Metadata != "" {
				updatedMetadata = ipfsurl.ParseMetadata([]byte(nft.Metadata))
			} else {
				// token_uri must exist and be fetchable
				if nft.Token_Uri != "" {
					if !strings.HasPrefix(nft.Token_Uri, "data:application/json") {
						response, err := request.Request(nft.Token_Uri)
						if err != nil {
							fmt.Println("Error fetching NFT Token URI", err)
							return
						}

						updatedMetadata = ipfsurl.ParseMetadata([]byte(response))

					} else {

						base64 := strings.TrimPrefix(nft.Token_Uri, "data:application/json;base64,")
						base64Decoded, _ := b64.StdEncoding.DecodeString(base64)

						if json.Valid(base64Decoded) {
							updatedMetadata = ipfsurl.ParseMetadata(base64Decoded)
						}
					}

				} else {
					fmt.Println("No metadata.")
					return
				}
			}

			// Add updated metadata to app result
			var appMetadata interface{}
			err = json.Unmarshal([]byte(updatedMetadata), &appMetadata)
			if err != nil {
				fmt.Println("Couldn't unmarshal: ", err)
				return
			}

			data.Result[i].AppMetadata = appMetadata

		}(i, nft) // End goroutine
	} // End for range loop

	wg.Wait() // Block execution until all goroutines are done

	// Group NFTs by collection address
	grouped := make(map[string][]Result)
	for i, collection := range data.Result {

		// if metadata is empty, don't add to final result
		if data.Result[i].Metadata != "" {
			grouped[collection.Token_Address] = append(grouped[collection.Token_Address], collection)
		}
	}

	data.Data = grouped // add grouped collections to Data field
	jsonByte, _ := json.Marshal(data)

	// Send HTTP Response
	w.WriteHeader(http.StatusOK)
	w.Write(jsonByte)
}
