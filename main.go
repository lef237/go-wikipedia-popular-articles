package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func printToday() {
	fmt.Println(time.Now().Format("2006-01-02"))
}

func parseFlag() string {
	langFlag := flag.String("lang", "ja", "Specify the language (e.g., 'ja' for Japanese, 'en' for English)")
	flag.Parse()
	return *langFlag
}

func buildURL(lang string) string {
	baseURL := "https://%s.wikipedia.org/w/api.php"
	return fmt.Sprintf(baseURL, lang) + "?action=query&list=mostviewed&format=json"
}

func checkAPIError(body []byte) error {
	var errResp struct {
		Error struct {
			Code string `json:"code"`
			Info string `json:"info"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return err
	}
	if errResp.Error.Code != "" {
		return fmt.Errorf("API error: %s - %s", errResp.Error.Code, errResp.Error.Info)
	}
	return nil
}

func fetchPopularArticles(lang string) (*ApiResponse, error) {
	url := buildURL(lang)

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

	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func fetchArticleDetailURL(lang string, title string) (string, error) {
	baseURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php", lang)
	query := fmt.Sprintf("%s?action=query&format=json&titles=%s&prop=info&inprop=url", baseURL, url.QueryEscape(title))

	resp, err := http.Get(query)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body failed: %w", err)
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

func main() {
	printToday()

	lang := parseFlag()

	articles, err := fetchPopularArticles(lang)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, page := range articles.Query.MostViewed {
		fmt.Printf("%d: Title: %s, View Count: %d\n", i, page.Title, page.Count)
	}

	var index int
	fmt.Println("Enter the index of the article to view its details (0-9):")
	_, err = fmt.Scanf("%d", &index)
	if err != nil || index < 0 || index >= len(articles.Query.MostViewed) {
		fmt.Println("Invalid input. Exiting.")
		return
	}

	title := articles.Query.MostViewed[index].Title
	url, err := fetchArticleDetailURL(lang, title)
	if err != nil {
		fmt.Println("Failed to fetch article details:", err)
		return
	}

	fmt.Printf("Details for \"%s\": %s\n", title, url)
}
