# AyayaBot - 棕色塵埃 2 Discord 公告與兌換碼機器人

AyayaBot 是一個使用 Go 語言編寫的 Discord 機器人，專門用於自動監控《棕色塵埃 2》(Brown Dust 2) 的最新官方公告與兌換碼。它能定時抓取資訊並發送到指定的 Discord 頻道，確保你不會錯過任何重要的遊戲動態與獎勵。

## 核心功能

*   **自動監控官方公告**: 定期檢查《棕色塵埃 2》官方新聞 API，發現新公告時按發佈時間「由舊到新」依次通知。支援將 HTML 內容轉換為 Markdown 並傳送。
*   **自動獲取兌換碼**: 從 [BD2 Pulse API](https://thebd2pulse.com/) 獲取最新的兌換碼資訊。
*   **背景執行與系統工作列**:
    *   支援 Windows 背景執行，不顯示控制台視窗。
    *   提供系統工作列圖示，點擊右鍵選單即可「關閉 (Quit)」。
*   **防重複執行與智慧警告**:
    *   使用具名 Mutex 偵測程式是否重複啟動。
    *   重複執行時自動跳出 Windows 原生 `MessageBox` 警告視窗通知使用者。
*   **自動化日誌管理**: 
    *   日誌輸出至執行目錄下的 `log/` 資料夾。
    *   檔案依日期命名 (`YYYYMMDD.log`)，便於追蹤。
    *   每日 `00:01` 定時切換日誌檔案。
    *   啟動時自動檢查並建立日誌目錄。
*   **智慧過濾**: 
    *   自動過濾已過期的兌換碼向。
    *   自動記錄已發送過的公告與兌換碼，避免重複推送。
*   **完整測試覆蓋**: 
    *   包含 Repository 層與 Scheduler 層的整合測試。
    *   使用 `httptest` 模擬 API 與介面抽離 (Mocking) 驗證邏輯。
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
    *   `redeem.service` / `news.service`: (選填) 設為 `false` 可停用該功能，預設為 `true`。
    *   `redeem.hide_embed` / `news.hide_embed`: (選填) 設為 `true` 可隱藏 Discord 連結預覽 (嵌入內容)，預設為 `false`。
    *   `news.send_content`: (選填) 設為 `true` 可在通知中傳送新聞內容，預設為 `false`。

### 4. 編譯與執行 (Windows)

為了讓程式在背景執行且不顯示黑視窗，並啟用系統工作列與防重複執行功能，請使用以下命令編譯：

```powershell
# 編譯為背景執行應用程式
go build -o ayayabot.exe -ldflags="-H=windowsgui" main.go

# 執行
./ayayabot.exe
```

執行後，你可以在 Windows 系統工作列（右下角）找到 AyayaBot 圖示，右鍵點擊可選擇「關閉 (Quit)」。

## 專案結構 (Go Idiomatic)

```text
AyayaBot/
├── assets/               # 靜態資源 (如 icon.ico)
├── config/               # 設定檔載入與環境變數處理
├── discord/              # Discord Bot 連線與訊息發送邏輯
├── logger/               # 檔案日誌系統 (初始化與切換邏輯)
├── model/                # 專案通用資料結構 (NewsItem, RedeemCode)
├── repository/
│   ├── bd2news/          # 棕色塵埃 2 新聞 API 交互邏輯
│   └── bd2redeem/         # 棕色塵埃 2 兌換碼 API 交互邏輯
├── scheduler/            # 排程管理與業務任務執行 (RunNewsTask, RunRedeemTask)
├── log/                  # 自動生成的日誌儲存目錄
├── config.yaml.example   # 設定檔範本 (受 Git 追蹤)
├── GEMINI.md             # Gemini CLI 開發規範指引
├── main.go               # 程式入口點，負責初始化與註冊任務
└── README.md             # 本說明文件
```

## 開發指南

本專案遵循 **Idiomatic Go** 命名規範：
*   Package 名稱均為簡潔的小寫單字 (例如 `logger`, `config`)。
*   日誌操作統一由 `logger` 套件管理，確保全局 `log` 套件的輸出被正確導向檔案。
*   業務邏輯與排程邏輯分離，任務執行函數統一命名為 `Run...Task`。
*   新增任務時，只需在 `scheduler` 建立任務並在 `main.go` 註冊。

### 執行測試 (Testing)

本專案使用 Go 標準測試框架，包含對外部 API 的整合模擬測試：

```bash
# 執行所有測試
go test -v ./...

# 執行特定套件測試
go test -v ./repository/bd2news/...
```

---
*本機器人僅供學習與技術交流使用，所有資料來源歸原官方所有。*
