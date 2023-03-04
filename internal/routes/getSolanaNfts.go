package routes

import (
	"api/pkg/ipfsurl"
	"api/pkg/request"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

func GetSolanaNfts(w http.ResponseWriter, r *http.Request) {

	type SolanaNFTOwners struct {
		Address string `json:"address"`
		Verified int `json:"verified"`
		Share int `json:"share"`
	}

	type SolanaNFTMetaplex struct {
		MetadataUri string `json:"metadataUri"` 
		UpdateAuthority string `json:"updateAuthority"` 
		SellerFeeBasisPoints int `json:"sellerFeeBasisPoints"` 
		PrimarySaleHappened int `json:"primarySaleHappened"`
		Owners []SolanaNFTOwners `json:"owners"`
		IsMutable bool `json:"isMutable"`
		MasterEdition bool `json:"masterEdition"`
	}

	type SolanaNFTResponse struct {
		Mint string `json:"mint"`
		Standard string `json:"standard"`
		Name string `json:"name"`
		Symbol string `json:"symbol"`
		Metaplex SolanaNFTMetaplex `json:"metaplex"`
	}

	type SolanaWalletNFTsResponse struct {
		Associated_Token_Address      string `json:"associatedTokenAddress"`
		Mint      										string `json:"mint"`
		Metadata string `json:"metadata"`
		Data SolanaNFTResponse `json:"data"`
		
	}

	// PARAMS
	vars := mux.Vars(r)
	address := vars["address"]
	network := vars["network"]

	// Moralis GetNfts https://docs.moralis.io/reference/getnfts-5
	response, err := request.SolanaAPIRequest("account/" + network + "/" + address + "/nft")
	if err != nil {
		fmt.Println("Error - ", err)
	}

	var solanaData[] SolanaWalletNFTsResponse

	err = json.Unmarshal([]byte(response), &solanaData)
	if err != nil {
		fmt.Println("Couldn't unmarshal first response for NFT portfolio: ", err)
		return
	}

	var wg sync.WaitGroup // Create WaitGroup to wait for all goroutines to finish

	// Loop through each NFT's results
	for i, nft := range solanaData {

		wg.Add(1)

		// Fetch each NFT's metadata in parallel
		go func(i int, nft SolanaWalletNFTsResponse) {

			// Decrease WaitGroup when goroutine has finished
			defer wg.Done()

			// Fetch each NFT's global metadata from Moralis API
			response, _:= request.SolanaAPIRequest("nft/" + network + "/" + nft.Mint + "/metadata")

			// Unmarshal
			var nftMetadata SolanaNFTResponse
			err = json.Unmarshal([]byte(response), &nftMetadata);
			if err != nil {
				fmt.Println("Error fetching NFT global metadata from Moralis", err)
			}

			// Store metadata and NFT data
			solanaData[i].Data = nftMetadata

			fmt.Println("global metadata response", response)

			// Get metadata from tokenUri
			response, err = request.Request(nftMetadata.Metaplex.MetadataUri)
			if err != nil {
				fmt.Println("Error fetching NFT Token URI", err)
			}
			solanaData[i].Metadata = ipfsurl.ParseMetadata([]byte(response))


		}(i, nft) // End goroutine
	} // End for range loop

	wg.Wait() // Block execution until all goroutines are done

	jsonByte, _ := json.Marshal(solanaData)

	// Send HTTP Response
	w.WriteHeader(http.StatusOK)
	w.Write(jsonByte)
}
