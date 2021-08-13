package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	res, err := http.Get("https://s3-eu-west-1.amazonaws.com/snowplow-hosted-assets/third-party/referer-parser/referers-latest.json")
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, res.Body)
		res.Body.Close()
		os.Exit(1)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		log.Fatal(err)
	}
	err = os.WriteFile("pkg/refparse/referrer.json", b, 0600)
}
