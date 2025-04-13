package repodocs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMetadata(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantMeta    docMetadata
		wantContent string
	}{
		{
			name:        "No metadata",
			content:     "This is just content",
			wantMeta:    docMetadata{},
			wantContent: "This is just content",
		},
		{
			name: "With metadata",
			content: `---
description: Test document
globs: *.go
---
This is the content`,
			wantMeta: docMetadata{
				description: "Test document",
				globs:       "*.go",
			},
			wantContent: "This is the content",
		},
		{
			name: "Incomplete metadata",
			content: `---
description: Test document
---
This is the content`,
			wantMeta: docMetadata{
				description: "Test document",
			},
			wantContent: "This is the content",
		},
		{
			name: "No end delimiter",
			content: `---
description: Test document
This is the content`,
			wantMeta: docMetadata{},
			wantContent: `---
description: Test document
This is the content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMeta, gotContent := parseMetadata(tt.content)
			if gotMeta.description != tt.wantMeta.description {
				t.Errorf("parseMetadata() description = %v, want %v", gotMeta.description, tt.wantMeta.description)
			}
			if gotMeta.globs != tt.wantMeta.globs {
				t.Errorf("parseMetadata() globs = %v, want %v", gotMeta.globs, tt.wantMeta.globs)
			}
			if gotContent != tt.wantContent {
				t.Errorf("parseMetadata() content = %v, want %v", gotContent, tt.wantContent)
			}
		})
	}
}

// 追加のメタデータテスト
func TestParseMetadataEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantMeta    docMetadata
		wantContent string
	}{
		{
			name: "Metadata with empty fields",
			content: `---
description: 
globs: 
---
Content here`,
			wantMeta: docMetadata{
				description: "",
				globs:       "",
			},
			wantContent: "Content here",
		},
		{
			name: "Metadata with malformed line",
			content: `---
description: Test description
invalid line without colon
globs: *.txt
---
Content with malformed metadata`,
			wantMeta: docMetadata{
				description: "Test description",
				globs:       "*.txt",
			},
			wantContent: "Content with malformed metadata",
		},
		{
			name: "Metadata with additional fields",
			content: `---
description: Test description
globs: *.txt
author: Test Author
---
Content with extra fields`,
			wantMeta: docMetadata{
				description: "Test description",
				globs:       "*.txt",
			},
			wantContent: "Content with extra fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMeta, gotContent := parseMetadata(tt.content)
			if gotMeta.description != tt.wantMeta.description {
				t.Errorf("parseMetadata() description = %v, want %v", gotMeta.description, tt.wantMeta.description)
			}
			if gotMeta.globs != tt.wantMeta.globs {
				t.Errorf("parseMetadata() globs = %v, want %v", gotMeta.globs, tt.wantMeta.globs)
			}
			if gotContent != tt.wantContent {
				t.Errorf("parseMetadata() content = %v, want %v", gotContent, tt.wantContent)
			}
		})
	}
}

// テストのセットアップ：ドキュメントを含むテストディレクトリを作成
func setupDocumentTestDir(t *testing.T) string {
	// t.TempDirを使用して、テスト終了時に自動クリーンアップされる一時ディレクトリを作成
	testDir := t.TempDir()

	// テスト用のドキュメントを作成
	docs := map[string]string{
		"doc1.mdc": `---
description: First document
globs: *.go
---
This is the first document content.`,
		"doc2.mdc": `---
description: Second document
globs: *.md
---
This is the second document content.`,
		"notadoc.txt": "This is not a document file.",
	}

	for name, content := range docs {
		path := filepath.Join(testDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("テストファイル%sの作成に失敗しました: %v", name, err)
		}
	}

	return testDir
}

// docManagerの基本機能テスト
func TestDocManager(t *testing.T) {
	testDir := setupDocumentTestDir(t)

	// newDocManagerのテスト
	t.Run("newDocManager", func(t *testing.T) {
		dm, err := newDocManager(testDir)
		if err != nil {
			t.Fatalf("newDocManager() error = %v", err)
		}
		if dm == nil {
			t.Fatal("newDocManager() returned nil docManager")
		}
		if len(dm.documents) != 2 {
			t.Errorf("newDocManager() loaded %d documents, want 2", len(dm.documents))
		}
	})

	// 後続のテスト用にDocManagerを作成
	dm, err := newDocManager(testDir)
	if err != nil {
		t.Fatalf("docManagerの作成に失敗しました: %v", err)
	}

	// listDocumentsのテスト
	t.Run("listDocuments", func(t *testing.T) {
		docs := dm.listDocuments()
		if len(docs) != 2 {
			t.Errorf("listDocuments() returned %d documents, want 2", len(docs))
		}

		// 両方のドキュメントが正しい説明と共に存在することを確認
		foundDoc1 := false
		foundDoc2 := false
		for _, doc := range docs {
			if doc["filename"] == "doc1.mdc" {
				foundDoc1 = true
				if doc["description"] != "First document" {
					t.Errorf("listDocuments() doc1 description = %s, want 'First document'", doc["description"])
				}
			}
			if doc["filename"] == "doc2.mdc" {
				foundDoc2 = true
				if doc["description"] != "Second document" {
					t.Errorf("listDocuments() doc2 description = %s, want 'Second document'", doc["description"])
				}
			}
		}
		if !foundDoc1 {
			t.Error("listDocuments() did not include doc1.mdc")
		}
		if !foundDoc2 {
			t.Error("listDocuments() did not include doc2.mdc")
		}
	})

	// getDocumentのテスト
	t.Run("getDocument", func(t *testing.T) {
		// 存在するドキュメントのテスト
		content, err := dm.getDocument("doc1.mdc")
		if err != nil {
			t.Errorf("getDocument() error = %v", err)
		}
		if content != "This is the first document content." {
			t.Errorf("getDocument() content = %s, want 'This is the first document content.'", content)
		}

		// 存在しないドキュメントのテスト
		_, err = dm.getDocument("nonexistent.mdc")
		if err == nil {
			t.Error("getDocument() did not return error for non-existent file")
		}
	})

	// getDocumentContentのテスト
	t.Run("getDocumentContent", func(t *testing.T) {
		content, err := dm.getDocumentContent("doc1.mdc")
		if err != nil {
			t.Errorf("getDocumentContent() error = %v", err)
		}
		expectedContent := `---
description: First document
globs: *.go
---
This is the first document content.`
		if content != expectedContent {
			t.Errorf("getDocumentContent() = %s, want %s", content, expectedContent)
		}
	})
}

// ツール機能のテスト
func TestToolFunctionality(t *testing.T) {
	testDir := setupDocumentTestDir(t)

	// DocManagerを作成
	dm, err := newDocManager(testDir)
	if err != nil {
		t.Fatalf("DocManagerの作成に失敗しました: %v", err)
	}

	// list_documentsツールのシミュレーション
	t.Run("list_documents_tool", func(t *testing.T) {
		// ドキュメント一覧ツールの機能をシミュレート
		docs := dm.listDocuments()

		// JSONに変換
		jsonData, err := json.Marshal(docs)
		if err != nil {
			t.Fatalf("ドキュメントのJSONへの変換に失敗しました: %v", err)
		}

		// JSONの結果を検証
		var result []map[string]string
		err = json.Unmarshal(jsonData, &result)
		if err != nil {
			t.Fatalf("JSONのアンマーシャルに失敗しました: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("ドキュメント数が期待と異なります: got %d, want 2", len(result))
		}

		// 各ドキュメントを検証
		for _, doc := range result {
			filename := doc["filename"]
			if filename != "doc1.mdc" && filename != "doc2.mdc" {
				t.Errorf("予期しないドキュメントファイル名: %s", filename)
			}
		}
	})

	// get_documentツールのシミュレーション
	t.Run("get_document_tool", func(t *testing.T) {
		// 存在するドキュメントを取得
		content, err := dm.getDocument("doc1.mdc")
		if err != nil {
			t.Fatalf("ドキュメントの取得に失敗しました: %v", err)
		}

		if content != "This is the first document content." {
			t.Errorf("ドキュメントの内容が予期しないものです: %s", content)
		}

		// 存在しないドキュメントを取得
		_, err = dm.getDocument("nonexistent.mdc")
		if err == nil {
			t.Error("存在しないドキュメントでエラーが返されませんでした")
		}
	})
}

// ファイル操作エラーのテスト
func TestFileOperationErrors(t *testing.T) {
	// 存在しないディレクトリでDocManagerを作成
	t.Run("nonexistent_directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(os.TempDir(), "nonexistent-dir-"+t.Name())
		_, err := newDocManager(nonExistentDir)
		if err == nil {
			t.Error("存在しないディレクトリでエラーが返されませんでした")
		}
	})
}
