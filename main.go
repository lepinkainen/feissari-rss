package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// Version is set during build via ldflags
var Version = "dev"

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

	req.Header.Set("User-Agent", fmt.Sprintf("FeissariRSS/%s (https://github.com/lepinkainen/feissari-rss)", Version))

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

	req.Header.Set("User-Agent", fmt.Sprintf("FeissariRSS/%s (https://github.com/lepinkainen/feissari-rss)", Version))

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
	// Define command line flags
	outputDir := flag.String("outdir", ".", "Directory where the RSS file will be saved")
	flag.Parse()

	//fmt.Printf("FeissariRSS %s starting up\n", Version)

	rssURL := "https://static.feissarimokat.com/dynamic/latest/posts.rss"

	// Fetch and parse the RSS feed
	rss, err := fetchRSS(rssURL)
	if err != nil {
		log.Fatalf("Error fetching RSS feed: %v", err)
	}

	// Create a new feed using gorilla/feeds
	feed := &feeds.Feed{
		Title:       rss.Channel.Title,
		Link:        &feeds.Link{Href: rss.Channel.Link},
		Description: rss.Channel.Description,
		Updated:     time.Now(),
		Created:     time.Now(),
		Author:      &feeds.Author{Name: "Feissarimokat"},
	}

	// Process each item
	for _, item := range rss.Channel.Items {
		//fmt.Printf("Fetching images for item %d: %s\n", i+1, item.Title)

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

		// Create a new feed item
		feedItem := &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: item.Link},
			Description: imageHTML.String(),
			Created:     time.Now(), // Since we don't have the original date, use current time
			Id:          item.Link,  // Use the link as a unique ID
		}

		feed.Items = append(feed.Items, feedItem)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Create the full output file path
	outputPath := filepath.Join(*outputDir, "feissarimokat.xml")

	// Create the RSS feed
	rssOutput, err := feed.ToAtom()
	if err != nil {
		log.Fatalf("Error generating RSS: %v", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, []byte(rssOutput), 0644); err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	//fmt.Printf("\nNew RSS feed generated at %s\n", outputPath)
}
