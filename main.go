package main

import (
	"flag"
	"fmt"
	"time"
)

type ApiResponse struct {
	// TODO
}

func fetchPopularArticles(date string) (*ApiResponse, error) {
	// TODO
	return nil, nil
}

func main() {
	// コマンドライン引数の処理
	dateFlag := flag.String("date", time.Now().Format("2006-01-02"), "Specify the date in YYYY-MM-DD format")
	flag.Parse()

	// ここまでの結果を出力
	fmt.Println(*dateFlag)
}
