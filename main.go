package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var http_client *http.Client

func init() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http_client = &http.Client{Transport: transport, Timeout: 2 * time.Second}
}

func convertIpToFileName(ip string) string {
	return strings.Replace(ip, ":", "-", -1)
}

func convertFileNameToIP(filename string) string {
	return strings.Replace(filename, "-", ":", -1)
}

func sendRequest(ip string, port int, channel chan string) {

	protocol := "http"
	if port == 443 {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://[%s]:%d", protocol, ip, port)
	fmt.Println(url)
	resp, err := http_client.Get(url)

	if err != nil {
		fmt.Println(err)
		channel <- fmt.Sprintf("Error %s", err)
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

	// append to file
	os.WriteFile("output.csv", []byte(csv_row), os.ModeAppend)
	channel <- "Done"

}

func main() {

	buff, err := os.ReadFile("./80.txt")
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}

	ips := strings.Split(string(buff), "\n")

	channel := make(chan string)

	for _, ip := range ips {
		if ip == "" {
			continue
		}

		go sendRequest(ip, 80, channel)
	}

}
