package main

import (
	"fmt"
	"time"

	"goscraping"
)

func main() {
	fmt.Println("Starting scrape...")

	// 1. Simple Fetch simulating Chrome
	resp, err := goscraping.Fetch("https://tls.peet.ws/api/all", &goscraping.Options{
		Method:        "GET",
		SessionID:     "session-1",
		Timeout:       30 * time.Second,
		HeaderProfile: "chrome",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body length: %d\n", len(resp.Body))
	// Print first 200 chars to see some JSON output
	if len(resp.Body) > 200 {
		fmt.Println(string(resp.Body[:200]))
	} else {
		fmt.Println(string(resp.Body))
	}

	// 2. Fetch with Safari profile
	resp2, err := goscraping.Fetch("https://httpbin.org/headers", &goscraping.Options{
		Method:        "GET",
		SessionID:     "session-safari", // Different session
		HeaderProfile: "safari",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nSafari Status: %d\n", resp2.StatusCode)
	fmt.Println(string(resp2.Body))
}
