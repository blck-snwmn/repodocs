package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/blck-snwmn/repodocs"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// コマンドライン引数からディレクトリを取得
	dirPath := flag.String("dir", ".", "ドキュメントディレクトリのパス")
	flag.Parse()

	// ライブラリを使用してMCPサーバーを作成
	s, err := repodocs.CreateMCPServer(*dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "サーバーの初期化に失敗: %v\n", err)
		os.Exit(1)
	}

	// サーバーを起動
	fmt.Fprintf(os.Stderr, "ドキュメントサーバーを起動しました。ディレクトリ: %s\n", *dirPath)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "サーバーエラー: %v\n", err)
	}
}
