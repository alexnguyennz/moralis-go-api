package ipfsurl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func ParseMetadata(response []byte) string {
	var metadata map[string]interface{}

	err := json.Unmarshal(response, &metadata)
	if err != nil {
		test, _ := json.Marshal(response)
		fmt.Println("test", string(test))
		fmt.Println("Couldn't unmarshal (ParseMetadata)", err)
	}

	// changeIpfsUrl
	if metadata != nil {

		// some NFTs use image_url, set it back to image
		if metadata["image_url"] != nil {
			metadata["image"] = metadata["image_url"]
		}

		if metadata["image"] != nil {

			// change URL if it uses IPFS protocol
			metadata["image"] = ChangeIpfsUrl(metadata["image"].(string))

			metadata["original_image"] = metadata["image"] // set original image to use as backup

			// Get contentType
			client := &http.Client{
				Timeout: 1 * time.Second,
			}

			resp, err := client.Head(metadata["image"].(string))
			if err != nil {
				// if getting Content-Type takes too long, return a default type and length
				fmt.Println("contentType HEAD error:", err)

				metadata["content_length"] = "0"
				metadata["content_type"] = "image/png"

				// Format into JSON
				jsonByte, _ := json.Marshal(metadata)
				json := string(jsonByte)

				return json
			}

			metadata["content_length"] = resp.ContentLength
			metadata["content_type"] = resp.Header.Get("Content-Type")

			// Format into JSON
			jsonByte, _ := json.Marshal(metadata)
			json := string(jsonByte)

			//data.Result[i].Metadata = json // update original response
			return json
		}
	}

	return ""
}

func ChangeIpfsUrl(nftUrl string) string {

	// parse image URL
	u, _ := url.Parse(nftUrl)

	if strings.Contains(nftUrl, "ipfs") {
		if strings.HasPrefix(nftUrl, "ipfs://ipfs/") {
			u.Path = "/ipfs" + u.Path
		} else if strings.HasPrefix(nftUrl, "ipfs://") {
			u.Path = "/ipfs/" + u.Host + u.Path
		}

		u.Scheme = "https"
		u.Host = "ipfs.io"

	}

	return u.String()

}