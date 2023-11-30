package main

import (
	"encoding/json"
	"flag"
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

// 引数によって日本語と英語を出し分ける
// Wikipedia API から人気記事を取得する処理を実装
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
	// 構造体のフィールドが見つからないこと自体はエラーではない
	// そのため、通常のレスポンスでも問題なく処理を続ける
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

	// HTTPリクエストの実行
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// レスポンスの読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}

	// エラーレスポンスの解析
	if err := checkAPIError(body); err != nil {
		return nil, err
	}

	// 通常のレスポンスが返ってきたとき：JSONのパース
	var apiResp ApiResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, err
	}

	return &apiResp, nil
}

// 今日の日付を出力
func printToday() {
	fmt.Println(time.Now().Format("2006-01-02"))
}

// 言語のコマンドライン引数をパース
func parseFlag() string {
	langFlag := flag.String("lang", "ja", "Specify the language (e.g., 'ja' for Japanese, 'en' for English)")
	flag.Parse()
	return *langFlag
}

func main() {
	printToday()

	// 言語の文字列を取得
	lang := parseFlag()

	// Wikipedia API から人気記事を取得
	articles, err := fetchPopularArticles(lang)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 記事の表示
	for _, page := range articles.Query.MostViewed {
		fmt.Printf("Title: %s\nView Count: %d\n\n", page.Title, page.Count)
	}
}
