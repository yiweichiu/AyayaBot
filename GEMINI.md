# AyayaBot 開發指南 (Gemini CLI)

這是為 Gemini CLI 提供的專案上下文與規範指南。在執行任何開發任務前，請優先遵守此文件中的指令。

## 專案核心目標
這是一個 Discord 機器人，專門用於監控並發送《棕色塵埃 2》(Brown Dust 2) 的最新公告與兌換碼。

## 技術架構
- **語言**: Go (Golang)
- **主要依賴**: `discordgo` (Discord SDK), `robfig/cron/v3` (排程管理), `yaml.v2` (設定檔)
- **目錄結構**:
  - `config`: 設定檔載入邏輯。
  - `discord`: Discord 機器人連線與訊息發送邏輯。
  - `model`: 專案通用的資料結構 (News, Redeem)。
  - `repository/bd2news`: 負責與 BD2 新聞 API 交互。
  - `repository/bd2redeem`: 負責與 BD2 兌換碼 API 交互。
  - `scheduler`: 負責排程管理與具體業務任務的執行 (`RunNewsTask`, `RunRedeemTask`)。

## 工程標準與慣例 (Mandates)
1. **命名規範**:
   - Package 名稱必須為單一單字的小寫格式 (如 `bd2news`, `bd2redeem`)，禁止使用底線。
   - 檔案命名應簡潔且與功能直接相關 (如 `client.go`, `news.go`)。
   - 務必遵循 Golang 的慣用法 (Idiomatic Go)。
2. **安全性**:
   - **嚴禁**將 `config.yaml` 納入 Git 追蹤。
   - 修改 `config` 相關邏輯時，務必確保不洩漏 Token。
3. **排程邏輯**:
   - 業務任務 (如抓取、通知) 應定義在 `scheduler` 目錄下的對應檔案中 (如 `RunNewsTask`)。
   - 排程註冊統一在 `main.go` 中透過 `AddJob(spec, func)` 進行。

## 常用命令
- **編譯專案**: `go build -o ayayabot main.go`
- **執行**: `./ayayabot`

## 溝通偏好
- **語言**: 優先使用 **繁體中文** 進行交流。
- **回覆風格**: 保持專業、直接、高訊號比，減少冗餘的對話。
