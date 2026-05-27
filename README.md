# AyayaBot - 多功能 Discord 通知與管理機器人

AyayaBot 是一個使用 Go 語言編寫的高效能 Discord 機器人，旨在提供自動化的資訊監控、通知推送與系統管理功能。它具備跨平台支援（Windows & macOS），並提供直觀的系統工作列操作介面。

## 核心功能

*   **自動化資訊監控**: 支援多來源資訊抓取（如遊戲公告、兌換碼等）。具備智慧過濾與紀錄功能，確保資訊不重複推送。
    *   *目前內建支援《棕色塵埃 2》(Brown Dust 2) 的官方新聞與兌換碼監控。*
*   **強大的訊息處理**: 
    *   支援 HTML 轉 Markdown 格式轉換，讓通知內容在 Discord 上更易讀。
    *   具備智慧截斷功能，確保長訊息符合 Discord 的 2000 字元限制且不破壞 UTF-8 編碼與標籤格式。
*   **系統工作列整合 (GUI)**:
    *   提供系統工作列圖示與右鍵選單，可直接在介面上開關各項功能、調整通知標註對象、變更檢查頻率等。
*   **自動更新系統**: 
    *   內建版本檢查與自動更新機制，可直接透過 GitHub Releases 下載並更新執行檔，更新後自動重新啟動。
*   **跨平台與防重複執行**:
    *   **Windows**: 使用具名 Mutex 確保單一實例執行，支援 `windowsgui` 模式隱藏控制台視窗，並使用原生 `MessageBox` 提示資訊。
    *   **macOS**: 使用檔案鎖 (Flock) 確保單一實例執行，並透過 `osascript` 提供原生通知對話框。
*   **自動化日誌管理**: 
    *   檔案依日期命名 (`YYYYMMDD.log`)。
    *   每日 `00:01` 定時切換日誌檔案。
    *   支援追蹤系統狀態與任務執行紀錄。
*   **靈活排程與標註**: 
    *   支援 `cron` 表達式排程設定。
    *   可自訂各項任務的標註對象（如 `@everyone`, `@here` 或特定的身分組 ID）。

## 安裝與設定

### 1. 下載預編譯版本

請至 [Releases](https://github.com/yiweichiu/AyayaBot/releases) 頁面下載適用於您平台的壓縮包：
*   **Windows**: 下載 `.zip` 檔，解壓縮後執行 `ayayabot.exe`。
*   **macOS**: 下載 `.tar.gz` 檔，解壓縮後執行 `ayayabot`。

### 2. 設定設定檔 (config.yaml)

首次執行前，請參考 `config.yaml.example` 建立 `config.yaml` 並填入您的 Discord Bot Token 與相關設定。

### 3. 自行編譯

如果您想自行編譯，請確保系統已安裝 Go 1.21 或更高版本：

```bash
git clone https://github.com/yiweichiu/AyayaBot.git
cd AyayaBot
go mod tidy

# Windows (背景執行)
go build -o ayayabot.exe -ldflags="-H=windowsgui" .

# macOS
go build -o ayayabot .
```

## 專案結構 (Go Idiomatic)

```text
AyayaBot/
├── assets/               # 靜態資源 (如圖示)
├── config/               # 設定檔載入與邏輯
├── discord/              # Discord Bot 連線與訊息發送
├── logger/               # 自動化日誌系統
├── model/                # 專案通用資料結構
├── repository/           # 資料來源 API 交互邏輯 (可擴充)
├── scheduler/            # 排程管理與業務任務執行
├── updater/              # 自動更新邏輯
├── main.go               # 程式入口與工作列選單邏輯
├── main_windows.go       # Windows 平台專屬實作
├── main_darwin.go        # macOS 平台專屬實作
└── README.md             # 本說明文件
```

## 開發指南

請參考 [GEMINI.md](./GEMINI.md) 了解詳細的開發規範、命名慣例與技術標準。

---
*本機器人僅供學習與技術交流使用。*

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
