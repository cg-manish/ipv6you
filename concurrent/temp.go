package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func fetch(url string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()          // Mark this goroutine as done when it finishes
	defer func() { <-sem }() // Release the semaphore slot

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close() // Ensure the response body is closed
	fmt.Printf("Fetched %s with status %s\n", url, resp.Status)
}

func xx() {
	urls := make([]string, totalRequests)
	for i := 0; i < totalRequests; i++ {
		urls[i] = fmt.Sprintf("https://example.com/page/%d", i) // Example URLs
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentRequests) // Semaphore to limit concurrency

	start := time.Now()

	for _, url := range urls {
		sem <- struct{}{}       // Acquire a token
		wg.Add(1)               // Increment the WaitGroup counter
		go fetch(url, &wg, sem) // Call fetch in a goroutine
	}

	wg.Wait() // Wait for all goroutines to finish

	elapsed := time.Since(start)
	fmt.Printf("All requests completed in %v\n", elapsed)
}
