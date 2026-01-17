package main

import (
	"fmt"
	"time"

	"github.com/MrAhmedElkady/goscraping"
	"github.com/MrAhmedElkady/goscraping/types"
)

func main() {
	// 1. Configure Options
	opts := types.DefaultOptions()
	opts.Timeout = 10 * time.Second
	opts.Method = "GET"
	// Optional: Enable Debug to see what's happening (clean logs now)
	opts.Debug = true

	// 2. Target URL
	url := "https://httpbin.org/get"
	fmt.Printf("Fetching %s...\n", url)

	// 3. Execute Fetch
	resp, err := goscraping.Fetch(url, opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	// Note: Response Body is []byte, so no Close() needed.

	// 4. Output Results
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Identity: %s (%s)\n", opts.Identity.Browser, opts.Identity.OS)

	// Print a snippet of the body
	bodyPreview := string(resp.Body)
	if len(bodyPreview) > 200 {
		bodyPreview = bodyPreview[:200] + "..."
	}
	fmt.Printf("Body: %s\n", bodyPreview)
}
