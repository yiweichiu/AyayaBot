# AyayaBot - 棕色塵埃2 Discord 兌換碼機器人

AyayaBot 是一個 Go 語言編寫的 Discord 機器人，用於自動獲取《棕色塵埃2》的兌換碼，並將最新的有效兌換碼發送到指定的 Discord 頻道。它會自動過濾已過期的兌換碼，並僅發送新發現的兌換碼。

## 功能

*   **定時獲取兌換碼**: 定期從指定的 API (`https://api.thebd2pulse.com/redeem`) 獲取最新的《棕色塵埃2》兌換碼資訊。
*   **自動過濾過期兌換碼**: 根據兌換碼的有效期限和當前時間，自動排除已過期的兌換碼。
*   **僅發送新兌換碼**: 機器人會記錄已發送過的兌換碼，只將新發現的有效兌換碼發送到 Discord 頻道，避免重複發送。
*   **包含獎勵資訊**: 發送的訊息中會包含兌換碼及其對應的繁體中文獎勵內容，格式為 `{兌換碼}:{獎勵內容}`。
*   **靈活的配置**: 所有關鍵資訊（Discord Bot Token、頻道 ID、API URL、API Key、排程）均透過 YAML 配置文件進行管理。
*   **可配置的排程**: 支援基於 `cron` 表達式的靈活排程設定。

## 安裝與設定

### 1. 前提條件

在運行本專案之前，請確保您的系統已安裝 Go 語言環境（版本 1.25.5 或更高）。

### 2. 下載專案

```bash
git clone https://github.com/yiweichiu/AyayaBot.git
cd AyayaBot
```

### 3. 安裝依賴

進入專案目錄後，運行以下命令安裝所需的 Go 模組：

```bash
go mod tidy
```

### 4. 配置 `config.yaml`

在專案根目錄下，您會找到一個 `config.yaml` 檔案。請根據您的需求編輯此檔案：

```yaml
discord:
  token: "Bot YOUR_DISCORD_BOT_TOKEN" # 替換為您的 Discord Bot Token。請務必在 Token 前加上 "Bot "
  channel_id: "YOUR_DISCORD_CHANNEL_ID" # 替換為您希望機器人發送訊息的 Discord 頻道 ID
api:
  url: "https://api.thebd2pulse.com/redeem" # 棕色塵埃2兌換碼 API URL
  api_key: "pulse-key-abc123-xyz789-very-secret" # 替換為您的 API Key
schedule:
  - "0 8 * * *" # 每日上午 08:00
  - "0 18 * * *" # 每日下午 18:00
  # 您可以根據 cron 格式添加更多排程，例如每分鐘一次為 "- "* * * * *""
```

*   **獲取 Discord Bot Token**: 請參閱 [Discord 開發者門戶](https://discord.com/developers/applications) 創建一個 Bot 應用程式並獲取 Token。
*   **獲取 Discord 頻道 ID**: 在 Discord 客戶端中啟用開發者模式（使用者設定 -> 高級），然後右鍵點擊您想要發送訊息的頻道，選擇「複製 ID」。

### 5. `redeem.json` 檔案

機器人會在運行目錄下自動創建和維護 `redeem.json` 檔案，用於記錄已發送過的兌換碼。您無需手動創建或修改它。

## 運行

完成上述設定後，您可以在專案根目錄下運行機器人：

```bash
go run main.go
```

機器人將會啟動，連接到 Discord，並根據 `config.yaml` 中定義的排程獲取並發送兌換碼。您可以在終端中看到日誌輸出。

要停止機器人，請在終端中按 `CTRL-C`。

## 文件結構

```
AyayaBot/
├── api/                  # 處理兌換碼 API 請求的邏輯
│   └── bd2pulse.go
├── config/               # 配置加載邏輯
│   └── config.go
├── discord/              # Discord 機器人相關功能（發送訊息）
│   └── discord.go
├── scheduler/            # 任務排程邏輯
│   └── scheduler.go
├── go.mod                # Go 模組文件
├── go.sum                # Go 模組依賴的總和校驗文件
├── config.yaml           # 機器人配置檔案
├── main.go               # 機器人主程式入口
└── README.md             # 本說明文件
```

## 未來擴展

*   **Webhook 功能**: 專案設計時已考慮到未來可以集成 Webhook 功能，以支持更靈活的訊息推送方式或與其他服務的集成。
*   **更多 API 支援**: 擴展支持更多遊戲或服務的兌換碼 API。
*   **命令功能**: 為 Discord 機器人添加命令功能，允許用戶手動觸發獲取兌換碼或查詢狀態。
