package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func fetchPopularArticles() (*ApiResponse, error) {
	// Wikipedia API から人気記事を取得する処理を実装
	// 日本のWikipedia
	url := fmt.Sprintf("https://ja.wikipedia.org/w/api.php?action=query&list=mostviewed&format=json")

	// HTTPリクエストの実行
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// レスポンスの読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// レスポンスの表示
	fmt.Println(string(body))

	// JSONのパース
	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func main() {
	// 今日の日付を出力
	fmt.Println(time.Now().Format("2006-01-02"))

	// Wikipedia API から人気記事を取得
	articles, err := fetchPopularArticles()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 記事の表示
	for _, page := range articles.Query.MostViewed {
		fmt.Printf("Title: %s\nView Count: %d\n\n", page.Title, page.Count)
	}
}
