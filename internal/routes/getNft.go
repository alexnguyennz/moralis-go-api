package routes

import (
	"api/pkg/ipfsurl"
	"api/pkg/request"
	b64 "encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func GetNft(w http.ResponseWriter, r *http.Request) {

	type Data struct {
		Token_Address       string 			`json:"token_address"`
		Token_Id            string 			`json:"token_id"`
		Amount              string 			`json:"amount"`
		Owner_Of            string 			`json:"owner_of"`
		Token_Hash					string 			`json:"token_hash"`
		Block_Number_Minted string 			`json:"block_number_minted"`
		Block_Number        string 			`json:"block_number"`
		Transfer_Index   		[]int				`json:"transfer_index"`
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

	
	// get URL params
	vars := mux.Vars(r) 
	chain := vars["chain"]
	address := vars["address"]
	tokenId := vars["tokenId"]


	// Moralis getNFTMetadata https://docs.moralis.io/reference/getnftmetadata
	response, err := request.APIRequest(`/nft/` + address + `/` + tokenId + `?chain=` + chain)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(response))
		return 
	}


	// unmarshal Moralis API response
	var data Data

	err = json.Unmarshal([]byte(response), &data)
	if err != nil {
		log.Println("Couldn't unmarshal: ", err)
	}


	// if metadata exists, then parse it
	updatedMetadata := ""
	if data.Metadata != "" {
		updatedMetadata = ipfsurl.ParseMetadata([]byte(data.Metadata))

	// use token_uri instead of metadata is null
	} else {
		if data.Token_Uri != "" {
			if !strings.HasPrefix(data.Token_Uri, "data:application/json") {
				response, err := request.Request(data.Token_Uri)
				if err != nil {
					log.Println("Error fetching NFT Token URI", err)
					return
				}

				updatedMetadata = ipfsurl.ParseMetadata([]byte(response))

			} else {
				base64 := strings.TrimPrefix(data.Token_Uri, "data:application/json;base64,")
				base64Decoded, _ := b64.StdEncoding.DecodeString(base64)

				if json.Valid(base64Decoded) { // make sure JSON is valid
					updatedMetadata = ipfsurl.ParseMetadata(base64Decoded)
				}
			}
		}
	}


	// Add updated metadata to app result 
	var appMetadata interface{}
	err = json.Unmarshal([]byte(updatedMetadata), &appMetadata)
	if err != nil {
		log.Println("Couldn't unmarshal: ", err)
		return
	}

	data.AppMetadata = appMetadata


	// format data for response
	jsonByte, _ := json.Marshal(data)


	// send HTTP response
	w.WriteHeader(http.StatusOK)
	w.Write(jsonByte)

}