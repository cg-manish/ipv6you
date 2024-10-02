package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	maxConcurrentRequests = 300 // Limit of concurrent requests
	// totalRequests         = 100 // Total number of requests
)

var http_client *http.Client

func init() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http_client = &http.Client{Transport: transport, Timeout: 3 * time.Second}
}

func convertIpToFileName(ip string) string {
	return strings.Replace(ip, ":", "-", -1)
}

func convertFileNameToIP(filename string) string {
	return strings.Replace(filename, "-", ":", -1)
}

func sendRequest(ip string, port int, wg *sync.WaitGroup, channelSemaphore chan struct{}, outputFile *os.File) {
	defer wg.Done()
	defer func() { <-channelSemaphore }()

	protocol := "http"
	if port == 443 {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://[%s]:%d", protocol, ip, port)
	fmt.Println(url)
	resp, err := http_client.Get(url)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	status_code := resp.StatusCode

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Chek if dir exists for that status code.
	_, err = os.Stat(fmt.Sprintf("response/%d", status_code))
	if err != nil {
		os.Mkdir(fmt.Sprintf("response/%d", status_code), os.ModePerm)
	}

	os.WriteFile(fmt.Sprintf("response/%d/%s.html", status_code, convertIpToFileName(ip)), body, os.ModePerm)

	// write to csv file with ip ,status and port
	csv_row := fmt.Sprintf("%s,%d,%d\n", ip, port, status_code)

	if _, err := outputFile.WriteString(csv_row); err != nil {
		fmt.Println(err)
	}
}

func main() {

	var wg sync.WaitGroup
	channelSemaphore := make(chan struct{}, maxConcurrentRequests)
	start := time.Now()

	buff, err := os.ReadFile("./80.txt")
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	ips := strings.Split(string(buff), "\n")

	outputCSV, err := os.OpenFile("output.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Error", err)
	}
	defer outputCSV.Close()

	if err != nil {
		fmt.Println("Error", err)
	}

	for _, ip := range ips {
		if ip == "" {
			continue
		}

		channelSemaphore <- struct{}{} // add a semaphore to the channel
		wg.Add(1)

		go sendRequest(ip, 80, &wg, channelSemaphore, outputCSV)
	}
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("All requests completed in %v\n", elapsed)
}
