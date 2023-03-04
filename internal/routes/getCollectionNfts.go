package routes

import (
	"api/pkg/ipfsurl"
	"api/pkg/request"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

func GetCollectionNfts(w http.ResponseWriter, r *http.Request) {

	type Result struct {
		Token_Address       string 			`json:"token_address"`
		Token_Id            string 			`json:"token_id"`
		Amount              string 			`json:"amount"`
		Token_Hash					string 			`json:"token_hash"`
		Block_Number_Minted string 			`json:"block_number_minted"`
		Contract_Type       string 			`json:"contract_type"`
		Name                string 			`json:"name"`
		Symbol              string 			`json:"symbol"`
		Token_Uri           string 			`json:"token_uri"`
		Metadata            string 			`json:"metadata"`
		AppMetadata         interface{} `json:"appMetadata"`
		Last_Token_Uri_Sync string 			`json:"last_token_uri_sync"`
		Last_Metadata_Sync  string			`json:"last_metadata_sync"`
		Minter_Address      string      `json:"minter_address"`
	}

	type Response struct {
		Total     int      `json:"total"`
		Page      int      `json:"page"`
		Page_Size int      `json:"page_size"`
		Cursor    string   `json:"cursor,omitempty"`
		Result    []Result `json:"result"`
	}

	// get URL params
	vars := mux.Vars(r)
	chain := vars["chain"]
	address := vars["address"]
	limit := vars["limit"]
	cursor := vars["cursor"]

	if cursor != "" {
		cursor = "&cursor=" + cursor
	}

	// Moralis getContractNFTs https://docs.moralis.io/reference/getcontractnfts
	response, err := request.APIRequest(`/nft/` + address + `/?chain=` + chain + `&limit=` + limit + cursor)
	if err != nil {
		fmt.Println("Error - ", err)
	}

	var data Response

	// unmarshal Moralis API response
	err = json.Unmarshal([]byte(response), &data)
	if err != nil {
		fmt.Println("Couldn't unmarshal: ", err)
		return
	}

	var wg sync.WaitGroup

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
							fmt.Println("Error fetching token_uri", err)

							updatedMetadata = "{}"
							return
						}
						// invalid character issue happens here
						fmt.Println("metadata ", nft.Metadata)
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

		}(i, nft)

	} // end of for loop

	wg.Wait()

	// format data for response
	jsonByte, _ := json.Marshal(data)

	// send HTTP response
	w.WriteHeader(http.StatusOK)
	w.Write(jsonByte)
}
