package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type ApiResponse struct {
	Query struct {
		MostViewed []struct {
			Title string `json:"title"`
			Count int    `json:"count"`
		} `json:"mostviewed"`
	} `json:"query"`
}

type APIError struct {
	Error struct {
		Code string `json:"code"`
		Info string `json:"info"`
	} `json:"error"`
}

func printToday() {
	fmt.Println(time.Now().Format("2006-01-02"))
}

func parseLangFlag() string {
	langFlag := flag.String("lang", "ja", "Specify the language (e.g., 'ja' for Japanese, 'en' for English)")
	flag.Parse()
	return *langFlag
}

func fetchAPI(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	if err := checkAPIError(body); err != nil {
		return nil, err
	}

	return body, nil
}

func checkAPIError(body []byte) error {
	var errResp APIError
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}
	if errResp.Error.Code != "" {
		return fmt.Errorf("API error: %s - %s", errResp.Error.Code, errResp.Error.Info)
	}
	return nil
}

func buildMostViewedURL(lang string) string {
	return fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=query&list=mostviewed&format=json", lang)
}

func buildArticleDetailURL(lang, title string) string {
	return fmt.Sprintf("https://%s.wikipedia.org/w/api.php?action=query&format=json&titles=%s&prop=info&inprop=url", lang, url.QueryEscape(title))
}

func getArticleDetails(lang, title string) (string, error) {
	url := buildArticleDetailURL(lang, title)
	body, err := fetchAPI(url)
	if err != nil {
		return "", fmt.Errorf("fetching article details failed: %w", err)
	}

	var result struct {
		Query struct {
			Pages map[string]struct {
				FullURL string `json:"fullurl"`
			} `json:"pages"`
		} `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	for _, page := range result.Query.Pages {
		return page.FullURL, nil
	}

	return "", fmt.Errorf("article URL not found")
}

func fetchPopularArticles(lang string) (*ApiResponse, error) {
	url := buildMostViewedURL(lang)
	body, err := fetchAPI(url)
	if err != nil {
		return nil, fmt.Errorf("fetching popular articles failed: %w", err)
	}

	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func containsFullWidthDigit(input string) bool {
	for _, r := range input {
		// ASCII characters fall within the range 0-127, so characters over 128 can be considered full-width characters.
		if unicode.IsDigit(r) && r > 127 {
			return true
		}
	}
	return false
}

func promptForArticleIndex(max int) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the index of the article to view its details (0-9):")

	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, fmt.Errorf("failed to read input: %w", err)
	}
	input = strings.TrimSpace(input)

	if containsFullWidthDigit(input) {
		return -1, fmt.Errorf("全角数字が含まれています。半角数字で入力してください")
	}

	index, err := strconv.Atoi(input)
	if err != nil || index < 0 || index >= max {
		return -1, fmt.Errorf("invalid input")
	}
	return index, nil
}

func main() {
	printToday()
	lang := parseLangFlag()

	articles, err := fetchPopularArticles(lang)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	for i, page := range articles.Query.MostViewed {
		fmt.Printf("%d: Title: %s, View Count: %d\n", i, page.Title, page.Count)
	}

	index, err := promptForArticleIndex(len(articles.Query.MostViewed))
	if err != nil {
		fmt.Println("Exiting due to:", err)
		return
	}

	title := articles.Query.MostViewed[index].Title
	url, err := getArticleDetails(lang, title)
	if err != nil {
		fmt.Printf("Failed to fetch article details: %s\n", err)
		return
	}

	fmt.Printf("Details for \"%s\": %s\n", title, url)
}
