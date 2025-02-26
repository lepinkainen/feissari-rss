package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// fetchImages fetches all images from a given URL's postbody div
func fetchImages(url string) ([]string, error) {
	// Create HTTP client with timeout and custom transport
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Create request with custom User-Agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "FeissariRSS/1.0 (https://github.com/lepinkainen/feissari-rss)")

	// Get the HTML page
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Find images only within the postbody div
	var images []string
	doc.Find("div.postbody img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			// Check if the URL is absolute or relative
			if !strings.HasPrefix(src, "http") {
				// Convert relative URLs to absolute
				baseURL := "https://static.feissarimokat.com"
				src = baseURL + src
			}
			images = append(images, src)
		}
	})

	return images, nil
}

// fetchRSS fetches and parses an RSS feed from a URL
func fetchRSS(url string) (*RSS, error) {
	// Create HTTP client with timeout and custom transport
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	// Create request with custom User-Agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "FeissariRSS/1.0 (https://github.com/lepinkainen/feissari-rss)")

	// Get the RSS feed
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Read the response body
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Parse the RSS feed
	var rss RSS
	if err := xml.Unmarshal(content, &rss); err != nil {
		return nil, err
	}

	return &rss, nil
}

func main() {
	rssURL := "https://static.feissarimokat.com/dynamic/latest/posts.rss"

	// Fetch and parse the RSS feed
	rss, err := fetchRSS(rssURL)
	if err != nil {
		log.Fatalf("Error fetching RSS feed: %v", err)
	}

	// Process each item
	for i, item := range rss.Channel.Items {
		fmt.Printf("Fetching images for item %d: %s\n", i+1, item.Title)

		images, err := fetchImages(item.Link)
		if err != nil {
			fmt.Printf("Error fetching images for %s: %v\n", item.Link, err)
			continue
		}

		// Update description to include images
		var imageHTML strings.Builder
		imageHTML.WriteString(item.Description)
		imageHTML.WriteString("\n\n")

		for _, img := range images {
			imageHTML.WriteString(fmt.Sprintf("<img src=\"%s\" alt=\"%s\">\n", img, item.Title))
		}

		rss.Channel.Items[i].Description = imageHTML.String()
	}

	// Generate new RSS feed
	output, err := xml.MarshalIndent(rss, "", "    ")
	if err != nil {
		log.Fatalf("Error generating RSS: %v", err)
	}

	// Add XML header
	xmlHeader := []byte(xml.Header)
	output = append(xmlHeader, output...)

	// Write to output file
	if err := os.WriteFile("feissarimokat.rss", output, 0644); err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	fmt.Println("\nNew RSS feed generated as feissarimokat.rss")
}
