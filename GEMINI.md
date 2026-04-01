# AyayaBot 開發指南 (Gemini CLI)

這是為 Gemini CLI 提供的專案上下文與規範指南。在執行任何開發任務前，請優先遵守此文件中的指令。

## 專案核心目標
這是一個 Discord 機器人，專門用於監控並發送《棕色塵埃 2》(Brown Dust 2) 的最新公告與兌換碼。

## 技術架構
- **語言**: Go (Golang)
- **主要依賴**: `discordgo` (Discord SDK), `robfig/cron/v3` (排程管理), `systray` (工作列管理), `yaml.v2` (設定檔)
- **目錄結構**:
  - `assets`: 存放圖示等靜態資源 (如 `icon.ico`)。
  - `config`: 設定檔載入邏輯。
  - `discord`: Discord 機器人連線與訊息發送邏輯。
  - `logger`: 檔案日誌系統，負責初始化與每日定時切換。
  - `model`: 專案通用的資料結構 (News, Redeem)。
  - `repository/bd2news`: 負責與 BD2 新聞 API 交互。
  - `repository/bd2redeem`: 負責與 BD2 兌換碼 API 交互。
  - `scheduler`: 負責排程管理與具體業務任務的執行 (`RunNewsTask`, `RunRedeemTask`)。

## 工程標準與慣例 (Mandates)
1.  **跨平台執行環境規範 (Windows & macOS)**:
    - **通用**: 程式必須具備 **單一執行個體檢查 (Single-Instance Check)** 並提供系統工作列 (System Tray) 操作介面。
    - **Windows**:
        - 必須支援以 `windowsgui` 模式編譯 (背景執行，無視窗)。
        - 單一實例檢查必須使用 Windows Named Mutex (`Local\AyayaBot-SingleInstance-Mutex`)。
        - 警告彈窗必須使用 Windows 原生 `MessageBoxW`。
        - 工作列圖示必須嵌入 `.ico` 格式以確保相容性。
    - **macOS (Darwin)**:
        - 單一實例檢查必須使用檔案鎖 (`syscall.Flock`) 作用於 `/tmp/AyayaBot-SingleInstance.lock`。
        - 警告彈窗必須透過 `osascript` (AppleScript) 實作。
        - 圖示嵌入邏輯應保留更換為 `.png` 的擴充性。
    - **實作要求**: 平台專屬邏輯必須嚴格使用 Go Build Tags (如 `//go:build windows` 或 `//go:build darwin`) 並拆分至 `main_{GOOS}.go` 檔案中。
2.  **命名規範**:
   - Package 名稱必須為單一單字的小寫格式 (如 `bd2news`, `bd2redeem`, `logger`)，禁止使用底線。
   - 檔案命名應簡潔且與功能直接相關 (如 `client.go`, `news.go`, `logger.go`)。
   - 務必遵循 Golang 的慣用法 (Idiomatic Go)。
3. **日誌規範**:
   - 系統日誌必須輸出至執行目錄下的 `log/` 資料夾。
   - 檔案命名格式為 `YYYYMMDD.log`。
   - 每日 `00:01` 必須執行日誌切換 (Rotate)，確保日誌按日期妥善分類。
4. **安全性**:
   - **嚴禁**將 `config.yaml` 納入 Git 追蹤。
   - 修改 `config` 相關邏輯時，務必確保不洩漏 Token。
5. **排程邏輯**:
   - 業務任務 (如抓取、通知) 應定義在 `scheduler` 目錄下的對應檔案中 (如 `RunNewsTask`)。
   - 排程註冊統一在 `main.go` 中透過 `AddJob(spec, func)` 進行。
   - **功能開關**: 所有功能模組 (如 News, Redeem) 必須支援透過 `config.yaml` 內的 `service: bool` 欄位控制啟動，且為確保向下相容性，該欄位在讀取失敗或不存在時應預設為 `true`。
   - **新聞通知順序**: 為了符合閱讀習慣，偵測到多筆新公告時，必須按發佈時間「由舊到新」傳送 (即對 `FetchNews` 回傳的降序列表進行反向遍歷)。
6. **測試規範 (Testing)**:
   - 所有外部 API 交互 (Repository 層) 必須具備使用 `httptest` 模擬回應的整合測試。
   - 核心業務任務 (Scheduler 層) 必須透過介面抽離 (如 `Messenger`) 進行邏輯驗證，確保過濾與發送順序正確。
   - 測試應避免產生持久性的副作用檔案（如 `news.json`），若需測試檔案讀寫應使用 Mock 或備份機制。

## 常用命令
- **編譯專案**: `go build -o ayayabot main.go`
- **執行**: `./ayayabot`

## 溝通偏好
- **語言**: 優先使用 **繁體中文** 進行交流。
- **回覆風格**: 保持專業、直接、高訊號比，減少冗餘的對話。
