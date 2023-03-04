package main

import (
	"api/internal/routes"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func ApiHandler(r *mux.Router) {
	// ResolveAddress
	r.HandleFunc("/resolve/chain/{chain}/address/{address}/limit/{limit}/{cursor}", routes.Resolve)
	r.HandleFunc("/resolve/chain/{chain}/address/{address}/limit/{limit}/", routes.Resolve)

	//Get Wallet NFTs 
	r.HandleFunc("/wallet/chain/{chain}/address/{address}/limit/{limit}/{cursor}", routes.GetWalletNfts) // if cursor param exists, match it
	r.HandleFunc("/wallet/chain/{chain}/address/{address}/limit/{limit}/", routes.GetWalletNfts)

	// Get Collection NFTs
	r.HandleFunc("/collection/chain/{chain}/address/{address}/limit/{limit}/{cursor}", routes.GetCollectionNfts)
	r.HandleFunc("/collection/chain/{chain}/address/{address}/limit/{limit}/", routes.GetCollectionNfts)
	
	// Get NFT Metadata
	r.HandleFunc("/nft/chain/{chain}/address/{address}/id/{tokenId}", routes.GetNft)

	// Search NFTs
	r.HandleFunc("/search/chain/{chain}/q/{q}/filter/{filter}/limit/{limit}/{cursor}", routes.SearchNfts)
	r.HandleFunc("/search/chain/{chain}/q/{q}/filter/{filter}/limit/{limit}/", routes.SearchNfts)

	// Get Random Wallet
	r.HandleFunc("/randomwallet", routes.GetRandomWallet)

	// Get Solana NFTs
	r.HandleFunc("/solana/nfts/wallet/network/{network}/address/{address}/", routes.GetSolanaNfts);

}

// set CORS origin
func corsOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if (os.Getenv("ENVIRONMENT") != "PROD") {
			(w).Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Development
		} else {
			(w).Header().Set("Access-Control-Allow-Origin", "http://productiondomain.com") // Production
		}
		
		next.ServeHTTP(w, r)
	})
}


func init() {

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set up logging 
	if (os.Getenv("ENVIRONMENT") == "PROD") {
		file, err := os.OpenFile("error.logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
	
		log.SetOutput(file) 
	}
	
	
}

func main() {

	r := mux.NewRouter() // gorilla/mux router

	r = r.PathPrefix("/api").Subrouter() // set all routes at domain.com/api/

	r.Use(corsOrigin) // apply middleware

	ApiHandler(r) // set routes

	// Run server
	if (os.Getenv("ENVIRONMENT") != "PROD") {
		log.Fatal(http.ListenAndServe("localhost:" + os.Getenv("GO_PORT"), r)) 	// Development; use localhost to prevent Windows firewall popup
	} else {
		log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), r))
	}
}
