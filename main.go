package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"io/fs"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ドキュメントのメタデータ
type DocMetadata struct {
	Description string
	Globs       string
	Filename    string
}

// ドキュメントマネージャー
type DocManager struct {
	root      *os.Root
	documents map[string]DocMetadata
}

// メタデータをパースする
func parseMetadata(content string) (DocMetadata, string) {
	meta := DocMetadata{}

	if !strings.HasPrefix(content, "---\n") {
		return meta, content
	}

	endMeta := strings.Index(content[4:], "---\n")
	if endMeta == -1 {
		return meta, content
	}

	metaStr := content[4 : endMeta+4]
	body := content[endMeta+8:]

	lines := strings.Split(metaStr, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "description":
			meta.Description = value
		case "globs":
			meta.Globs = value
		}
	}

	return meta, body
}

// 新しいドキュメントマネージャーを作成
func NewDocManager(dirPath string) (*DocManager, error) {
	root, err := os.OpenRoot(dirPath)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリを開けませんでした: %w", err)
	}

	dm := &DocManager{
		root:      root,
		documents: make(map[string]DocMetadata),
	}

	// ドキュメントをスキャンして読み込む
	err = dm.scanDocuments()
	if err != nil {
		return nil, err
	}

	return dm, nil
}

// ディレクトリをスキャンしてドキュメントを読み込む
func (dm *DocManager) scanDocuments() error {
	// os.Root の制約内で安全にファイル一覧を取得する
	files, err := dm.getFilesInDirectory()
	if err != nil {
		return fmt.Errorf("ディレクトリの読み込みに失敗: %w", err)
	}

	for _, filename := range files {
		if !strings.HasSuffix(filename, ".mdc") {
			continue
		}

		content, err := dm.getDocumentContent(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: %sの読み込みに失敗: %v\n", filename, err)
			continue
		}

		meta, _ := parseMetadata(content)
		meta.Filename = filename
		dm.documents[filename] = meta
	}

	return nil
}

// ディレクトリ内のファイル一覧を取得
func (dm *DocManager) getFilesInDirectory() ([]string, error) {
	files := []string{}

	// root.FSを使用してfsパッケージ互換のファイルシステムを取得
	fsys := dm.root.FS()

	// ファイルパス取得のためにfs.WalkDirを使用
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリは無視
		if d.IsDir() {
			return nil
		}

		// .mdcファイルのみ収集
		if strings.HasSuffix(path, ".mdc") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ディレクトリ内のファイル取得エラー: %w", err)
	}

	return files, nil
}

// ドキュメントの一覧を取得
func (dm *DocManager) listDocuments() []map[string]string {
	result := make([]map[string]string, 0, len(dm.documents))

	for filename, meta := range dm.documents {
		result = append(result, map[string]string{
			"filename":    filename,
			"description": meta.Description,
		})
	}

	return result
}

// ドキュメントの内容を取得
func (dm *DocManager) getDocumentContent(filename string) (string, error) {
	file, err := dm.root.Open(filename)
	if err != nil {
		return "", fmt.Errorf("ファイルを開けません: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	return string(content), nil
}

// ドキュメントを取得
func (dm *DocManager) getDocument(filename string) (string, error) {
	if _, exists := dm.documents[filename]; !exists {
		return "", fmt.Errorf("ドキュメントが見つかりません: %s", filename)
	}

	content, err := dm.getDocumentContent(filename)
	if err != nil {
		return "", err
	}

	_, body := parseMetadata(content)
	return body, nil
}

func main() {
	// コマンドライン引数からディレクトリを取得
	dirPath := flag.String("dir", ".", "ドキュメントディレクトリのパス")
	flag.Parse()

	// 絶対パスに変換
	absPath, err := filepath.Abs(*dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "パスの変換に失敗: %v\n", err)
		os.Exit(1)
	}

	// ドキュメントマネージャーを作成
	docManager, err := NewDocManager(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ドキュメントマネージャーの初期化に失敗: %v\n", err)
		os.Exit(1)
	}

	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Document Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// ドキュメント一覧ツール
	listDocsTool := mcp.NewTool("list_documents",
		mcp.WithDescription("ドキュメント一覧を取得"),
	)

	// ドキュメント一覧ツールのハンドラ
	s.AddTool(listDocsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		docs := docManager.listDocuments()
		jsonData, err := json.Marshal(docs)
		if err != nil {
			return nil, fmt.Errorf("JSONへの変換に失敗: %w", err)
		}
		return mcp.NewToolResultText(string(jsonData)), nil
	})
	// ドキュメント取得ツール
	getDocTool := mcp.NewTool("get_document",
		mcp.WithDescription("ドキュメントの内容を取得"),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("取得するドキュメントのファイル名"),
		),
	)
	// ドキュメント取得ツールのハンドラ
	s.AddTool(getDocTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filename := request.Params.Arguments["filename"].(string)

		content, err := docManager.getDocument(filename)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResultText(content), nil
	})

	// サーバーを起動
	fmt.Fprintf(os.Stderr, "ドキュメントサーバーを起動しました。ディレクトリ: %s\n", absPath)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "サーバーエラー: %v\n", err)
	}
}
