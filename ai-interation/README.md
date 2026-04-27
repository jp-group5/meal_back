# ai-interation (mock scaffold)

## 起動
```bash
go mod tidy
go run ./cmd/api
```

## meal-analysis
`multipart/form-data` で `image` フィールドに JPEG を入れて送信します。

## recommendation
JSON で送信します。

```json
{
  "user_id": "u12345",
  "target_date": "2026-04-28",
  "condition": "home_cooking"
}
```

## Project Structure

ai-interaction/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── router/
│   │   └── router.go
│   ├── handler/
│   │   ├── meal_analysis_handler.go
│   │   └── recommendation_handler.go
│   ├── usecase/
│   │   ├── meal_analysis_usecase.go
│   │   └── recommendation_usecase.go
│   ├── dto/
│   │   ├── meal_analysis.go
│   │   └── recommendation.go
│   ├── port/
│   │   ├── ai_client.go
│   │   └── repository.go
│   ├── infrastructure/
│   │   ├── openai/
│   │   │   └── client.go
│   │   ├── repository/
│   │   │   └── repository_impl.go
│   │   └── mock/
│   │       ├── ai_client_mock.go
│   │       └── repository_mock.go
│   └── config/
│       └── config.go
└── go.mod

Directory Overview
cmd/api/main.go
Application entry point. Starts the Gin server and loads the router.
internal/router/router.go
Defines API routes and connects endpoints to handlers.
internal/handler/
Handles HTTP requests and responses for each endpoint.
internal/usecase/
Contains the main business logic for meal analysis and recommendation.
internal/dto/
Defines request and response data structures used by the API.
internal/port/
Defines interfaces for external dependencies such as AI clients and repositories.
internal/infrastructure/openai/
Implements communication with OpenAI.
internal/infrastructure/repository/
Implements database access logic.
internal/infrastructure/mock/
Provides mock implementations for development and testing before the real services are ready.
internal/config/
Stores configuration values such as environment variables and application settings.

cmd/api/main.go
アプリの起動ファイルです。Gin を立ち上げて、router を読み込みます。
最初に実行される入口です。

internal/router/router.go
URL と handler の対応をまとめます。
たとえば、/api/v1/meal-analysis と /api/v1/recommendation をここで登録します。

internal/handler/meal_analysis_handler.go
/api/v1/meal-analysis のHTTP処理を書きます。
画像ファイルを受け取り、usecase を呼び、レスポンスを返します。

internal/handler/recommendation_handler.go
/api/v1/recommendation のHTTP処理を書きます。
user_id や profile を受け取り、usecase を呼び、レスポンスを返します。

internal/usecase/meal_analysis_usecase.go
食事画像解析の中心処理です。
画像データの受け渡し、AIへの依頼、結果の整形を担当します。

internal/usecase/recommendation_usecase.go
おすすめ献立生成の中心処理です。
DBから必要データを取得し、整形してAIに渡し、結果をまとめます。

internal/dto/meal_analysis.go
meal-analysis 用のリクエスト・レスポンス型を置きます。
フロントエンドが受け取りやすいJSON構造をここで定義します。

internal/dto/recommendation.go
recommendation 用のリクエスト・レスポンス型を置きます。
ユーザー情報、条件、返却結果をここで定義します。

internal/port/ai_client.go
AIに必要な機能を interface で定義します。
OpenAI 実装でも mock 実装でも、同じ形で呼べるようにします。

internal/port/repository.go
DBやデータ取得の interface を定義します。
他の人が作るDB実装に差し替えやすくなります。

internal/infrastructure/openai/client.go
OpenAI と通信する実装を置きます。
実際にAPIへリクエストを送り、返答を受け取る部分です。

internal/infrastructure/repository/repository_impl.go
DB担当の実装が入る場所です。
行動ログや食事記録を取得する処理をここに置きます。

internal/infrastructure/mock/ai_client_mock.go
OpenAI がまだできていない間に使うダミー実装です。
固定のJSONや文字列を返して、全体の流れを確認できます。

internal/infrastructure/mock/repository_mock.go
DBがまだない間に使うダミー実装です。
仮データを返して、recommendation の流れを先に作れます。

internal/config/config.go
環境変数や設定値をまとめます。
APIキー、ポート番号、開発/本番の切り替えなどを管理します。
