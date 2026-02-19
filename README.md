# AyayaBot - 棕色塵埃 2 Discord 公告與兌換碼機器人

AyayaBot 是一個使用 Go 語言編寫的 Discord 機器人，專門用於自動監控《棕色塵埃 2》(Brown Dust 2) 的最新官方公告與兌換碼。它能定時抓取資訊並發送到指定的 Discord 頻道，確保你不會錯過任何重要的遊戲動態與獎勵。

## 核心功能

*   **自動監控官方公告**: 定期檢查《棕色塵埃 2》官方新聞 API，發現新公告時立即通知。
*   **自動獲取兌換碼**: 從 [BD2 Pulse API](https://thebd2pulse.com/) 獲取最新的兌換碼資訊。
*   **智慧過濾**: 
    *   自動過濾已過期的兌換碼。
    *   自動記錄已發送過的公告與兌換碼，避免重複推送。
*   **多語言支援**: 優先提取繁體中文獎勵說明。
*   **靈活排程**: 支援基於 `cron` 表達式的多時段排程設定。
*   **安全管理**: 支援範本化設定檔管理，保護敏感的 API Key 與 Token。

## 安裝與設定

### 1. 前提條件

確保系統已安裝 Go 1.21 或更高版本。

### 2. 下載與安裝

```bash
git clone https://github.com/yiweichiu/AyayaBot.git
cd AyayaBot
go mod tidy
```

### 3. 設定設定檔 (重要)

本專案使用 `config.yaml` 進行管理，但為了安全起見，真實的設定檔已被加入 `.gitignore`。請依照以下步驟設定：

1.  複製範本檔：
    ```bash
    cp config.yaml.example config.yaml
    ```
2.  編輯 `config.yaml` 並填入你的真實資訊：
    *   `discord.token`: 你的 Discord Bot Token (需包含 `Bot ` 前綴)。
    *   `discord.channel_id`: 接收通知的頻道 ID。
    *   `redeem.api.api_key`: BD2 Pulse 的 API Key。

### 4. 編譯與執行

```bash
# 編譯
go build -o ayayabot main.go

# 執行
./ayayabot
```

## 專案結構 (Go Idiomatic)

```text
AyayaBot/
├── config/               # 設定檔載入與環境變數處理
├── discord/              # Discord Bot 連線與訊息發送邏輯
├── model/                # 專案通用資料結構 (NewsItem, RedeemCode)
├── repository/
│   ├── bd2news/          # 棕色塵埃 2 新聞 API 交互邏輯
│   └── bd2redeem/         # 棕色塵埃 2 兌換碼 API 交互邏輯
├── scheduler/            # 排程管理與業務任務執行 (RunNewsTask, RunRedeemTask)
├── config.yaml.example   # 設定檔範本 (受 Git 追蹤)
├── GEMINI.md             # Gemini CLI 開發規範指引
├── main.go               # 程式入口點，負責初始化與註冊任務
└── README.md             # 本說明文件
```

## 開發指南

本專案遵循 **Idiomatic Go** 命名規範：
*   Package 名稱均為簡潔的小寫單字。
*   業務邏輯與排程邏輯分離，任務執行函數統一命名為 `Run...Task`。
*   新增任務時，只需在 `scheduler` 建立任務並在 `main.go` 註冊。

---
*本機器人僅供學習與技術交流使用，所有資料來源歸原官方所有。*
