package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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

type APIClient interface {
	Fetch(url string) ([]byte, error)
}

type WikipediaAPIClient struct{}

func (WikipediaAPIClient) Fetch(apiURL string) ([]byte, error) {
	resp, err := http.Get(apiURL)
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

func printToday() {
	fmt.Println(time.Now().Format("2006-01-02"))
}

func parseLangFlag() string {
	langFlag := flag.String("lang", "ja", "Specify the language (e.g., 'ja' for Japanese, 'en' for English)")
	flag.Parse()
	return *langFlag
}

func buildAPIURL(lang string, params url.Values) string {
	return fmt.Sprintf("https://%s.wikipedia.org/w/api.php?%s", lang, params.Encode())
}

func buildMostViewedURL(lang string) string {
	return buildAPIURL(lang, url.Values{
		"action": {"query"},
		"list":   {"mostviewed"},
		"format": {"json"},
	})
}

func buildArticleDetailURL(lang, title string) string {
	return buildAPIURL(lang, url.Values{
		"action": {"query"},
		"format": {"json"},
		"titles": {title},
		"prop":   {"info"},
		"inprop": {"url"},
	})
}

func buildArticleSummaryURL(lang, title string) string {
	return buildAPIURL(lang, url.Values{
		"action":      {"query"},
		"format":      {"json"},
		"titles":      {title},
		"prop":        {"extracts"},
		"exintro":     {"true"},
		"explaintext": {"true"},
	})
}

func fetchJSON(client APIClient, apiURL, operation string, v any) error {
	body, err := client.Fetch(apiURL)
	if err != nil {
		return fmt.Errorf("%s failed: %w", operation, err)
	}
	return json.Unmarshal(body, v)
}

func getArticleDetails(client APIClient, lang, title string) (string, error) {
	var result struct {
		Query struct {
			Pages map[string]struct {
				FullURL string `json:"fullurl"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := fetchJSON(client, buildArticleDetailURL(lang, title), "fetching article details", &result); err != nil {
		return "", err
	}

	for _, page := range result.Query.Pages {
		return page.FullURL, nil
	}
	return "", errors.New("article URL not found")
}

func getArticleSummary(client APIClient, lang, title string) (string, error) {
	var result struct {
		Query struct {
			Pages map[string]struct {
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := fetchJSON(client, buildArticleSummaryURL(lang, title), "fetching article summary", &result); err != nil {
		return "", err
	}

	for _, page := range result.Query.Pages {
		return page.Extract, nil
	}
	return "", errors.New("article summary not found")
}

func fetchPopularArticles(client APIClient, lang string) (ApiResponse, error) {
	var apiResp ApiResponse
	if err := fetchJSON(client, buildMostViewedURL(lang), "fetching popular articles", &apiResp); err != nil {
		return ApiResponse{}, err
	}
	return apiResp, nil
}

func promptForArticleIndex(max int) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the index of the article to view its details (0-9):")

	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, fmt.Errorf("failed to read input: %w", err)
	}
	input = convertFullWidthDigitsToHalfWidth(strings.TrimSpace(input))

	index, err := strconv.Atoi(input)
	if err != nil || index < 0 || index >= max {
		return -1, errors.New("invalid input")
	}
	return index, nil
}

func convertFullWidthDigitsToHalfWidth(input string) string {
	var builder strings.Builder
	for _, r := range input {
		if r >= '０' && r <= '９' {
			builder.WriteRune(r - '０' + '0')
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func main() {
	printToday()
	lang := parseLangFlag()

	client := WikipediaAPIClient{}
	articles, err := fetchPopularArticles(client, lang)
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
	articleURL, err := getArticleDetails(client, lang, title)
	if err != nil {
		fmt.Printf("Failed to fetch article details: %s\n", err)
		return
	}

	summary, err := getArticleSummary(client, lang, title)
	if err != nil {
		fmt.Printf("Failed to fetch article summary: %s\n", err)
		return
	}

	fmt.Printf("Details for \"%s\": %s\n\n", title, articleURL)
	fmt.Println("Article Summary:", summary)
}
