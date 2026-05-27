# AyayaBot 開發指南 (Gemini CLI)

這是為 Gemini CLI 提供的專案上下文與規範指南。在執行任何開發任務前，請優先遵守此文件中的指令。

## 專案核心目標
這是一個多功能的 Discord 機器人，旨在提供自動化的資訊監控、通知推送與系統管理功能。具備跨平台執行與自動更新能力。

## 技術架構
- **語言**: Go (Golang)
- **主要依賴**: `discordgo` (Discord SDK), `robfig/cron/v3` (排程管理), `systray` (工作列管理), `yaml.v2` (設定檔)
- **目錄結構**:
  - `assets`: 存放圖示等靜態資源。
  - `config`: 設定檔載入與邏輯處理。
  - `discord`: Discord 機器人連線與訊息發送邏輯。
  - `logger`: 檔案日誌系統，負責初始化與每日定時切換。
  - `model`: 專案通用的資料結構 (News, Redeem)。
  - `repository`: 外部 API 交互邏輯層 (目前包含 BD2 新聞與兌換碼)。
  - `scheduler`: 負責排程管理與業務任務的執行，包含內容格式化轉換與訊息截斷邏輯。
  - `updater`: 版本檢查與自動更新邏輯。

## 工程標準與慣例 (Mandates)
1.  **跨平台執行環境規範 (Windows & macOS)**:
    - **通用**: 程式必須具備 **單一執行個體檢查 (Single-Instance Check)** 並提供系統工作列 (System Tray) 操作介面。
    - **Windows**:
        - 必須支援以 `windowsgui` 模式編譯 (背景執行，無視窗)。
        - 單一實例檢查必須使用 Windows Named Mutex (`Local\AyayaBot-SingleInstance-Mutex`)。
        - 警告與對話框必須使用 Windows 原生 `MessageBoxW`。
    - **macOS (Darwin)**:
        - 單一實例檢查必須使用檔案鎖 (`syscall.Flock`) 作用於 `/tmp/AyayaBot-SingleInstance.lock`。
        - 警告與對話框必須透過 `osascript` (AppleScript) 實作。
    - **實作要求**: 平台專屬邏輯必須嚴格使用 Go Build Tags (如 `//go:build windows`) 並拆分至 `main_{GOOS}.go` 檔案中。
2.  **命名規範**:
   - Package 名稱必須為單一單字的小寫格式，禁止使用底線。
   - 檔案命名應簡潔且與功能直接相關。
   - 務必遵循 Golang 的慣用法 (Idiomatic Go)。
3. **日誌規範**:
   - 系統日誌必須輸出至執行目錄下的 `log/` 資料夾。
   - 檔案命名格式為 `YYYYMMDD.log`。
   - 每日 `00:01` 必須執行日誌切換 (Rotate)。
4. **安全性**:
   - **嚴禁**將 `config.yaml` 納入 Git 追蹤。
   - 修改 `config` 相關邏輯時，務必確保不洩漏 Token。
5. **排程與內容處理**:
   - 業務任務 (如抓取、通知) 應定義在 `scheduler` 目錄下。
   - 排程註冊統一在 `main.go` 中透過 `AddJob(spec, func)` 進行。
   - **訊息截斷**: 傳送至 Discord 的訊息若可能超過 2000 字元，必須使用 `scheduler.TruncateString` 進行安全截斷，確保不破壞 UTF-8 編碼與 Markdown 標籤。
6. **測試規範 (Testing)**:
   - 所有外部 API 交互 (Repository 層) 必須具備使用 `httptest` 模擬回應的整合測試。
   - 核心業務任務 (Scheduler 層) 必須透過介面抽離進行邏輯驗證。
   - 測試應避免產生持久性的副作用檔案。
7. **版本管理與自動更新**:
   - **版本號**: 定義在 `updater/updater.go` 中的 `CurrentVersion`。正式發佈時應透過 `ldflags` 注入 Git Tag 名稱。
   - **更新源**: 使用 GitHub Releases。
   - **自動重新啟動**: 更新完成後應自動啟動新版本的執行檔並結束舊進程。

## 常用命令
- **編譯 Windows (背景)**: `go build -ldflags="-H=windowsgui" -o ayayabot.exe .`
- **編譯 macOS**: `go build -o ayayabot .`
- **執行測試**: `go test -v ./...`

## 溝通偏好
- **語言**: 優先使用 **繁體中文** 進行交流。
- **回覆風格**: 專業、直接、高訊號比。
