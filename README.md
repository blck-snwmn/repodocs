# repodocs

リポジトリ内のドキュメントを管理・提供するためのシンプルなMCPサーバー実装です。

## 概要

このツールは、指定したディレクトリ内の`.mdc`形式のドキュメントファイルを読み込み、MCP（Model Context Protocol）サーバーとして提供します。LLMアプリケーションからのアクセスを可能にします。

主な機能：
- ディレクトリ内の`.mdc`ファイルの自動スキャン
- ドキュメント一覧の取得
- ドキュメント内容の取得
- YAMLフロントマターによるメタデータのサポート

## インストール

```bash
go get github.com/youruser/repodocs
```

## 使い方

### サーバーの実行

```bash
repodocs --dir=/path/to/your/docs
```

オプション:
- `--dir`: ドキュメントが格納されているディレクトリのパス（デフォルト: カレントディレクトリ）

### ドキュメントファイル形式

ドキュメントは`.mdc`拡張子を持つファイルで、オプションのYAMLフロントマターを含むことができます：

```md
---
description: ドキュメントの説明
globs: 関連するファイルパターン
---

# 実際のドキュメント内容

これはマークダウン形式のドキュメントです。
```

## MCPツール

このサーバーは以下のMCPツールを提供します：

1. `list_documents` - 利用可能なすべてのドキュメントの一覧を取得
2. `get_document` - 指定したファイル名のドキュメント内容を取得

## 依存ライブラリ

- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Go言語用MCPプロトコル実装

## ライセンス

MIT 