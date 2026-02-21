package scheduler

import (
	"os"
	"strings"
	"testing"

	"github.com/yiweichiu/AyayaBot/model"
)

func TestProcessRedeemTask(t *testing.T) {
	// 暫時修改檔名以避免影響開發環境的 redeem.json
	originalFileName := redeemFilePath
	const testRedeemFile = "redeem_test.json"
	// 這裡有點棘手，因為 redeemFilePath 是 package 層級的常數且在 redeem.go 定義
	// 為了測試，我們暫時確保測試結束後刪除產生的檔案
	defer os.Remove(testRedeemFile)

	mockBot := &MockMessenger{}
	
	fetchedCodes := []model.RedeemCodeInfo{
		{Code: "NEW_CODE_1", Reward: "Reward 1"},
		{Code: "OLD_CODE", Reward: "Old Reward"},
	}
	
	previouslySent := []string{"OLD_CODE"}

	// 執行測試邏輯
	// 注意：由於 redeemFilePath 在 redeem.go 是 const，
	// 我們在測試中會真的寫入一個名為 "redeem.json" 的檔案（除非重構為變數）。
	// 為了安全起見，我們先備份現有的 redeem.json (如果有的話)。
	backupName := "redeem.json.bak"
	hasBackup := false
	if _, err := os.Stat(originalFileName); err == nil {
		os.Rename(originalFileName, backupName)
		hasBackup = true
	}
	defer func() {
		os.Remove(originalFileName) // 刪除測試產生的
		if hasBackup {
			os.Rename(backupName, originalFileName)
		}
	}()

	err := processRedeemTask(mockBot, fetchedCodes, previouslySent)
	if err != nil {
		t.Fatalf("processRedeemTask failed: %v", err)
	}

	// 應只有一個新代碼的訊息
	if len(mockBot.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(mockBot.Messages))
	}

	if !strings.Contains(mockBot.Messages[0], "NEW_CODE_1") {
		t.Errorf("Expected message to contain NEW_CODE_1, got: %s", mockBot.Messages[0])
	}

	if strings.Contains(mockBot.Messages[0], "OLD_CODE") {
		t.Error("Message should not contain OLD_CODE")
	}

	// 驗證檔案是否正確儲存 (應包含所有目前的代碼)
	savedCodes, err := loadRedeemCodesFromFile(originalFileName)
	if err != nil {
		t.Fatalf("Failed to load saved codes: %v", err)
	}
	if len(savedCodes) != 2 {
		t.Errorf("Expected 2 saved codes, got %d", len(savedCodes))
	}
}

func TestProcessRedeemTask_NoNewCodes(t *testing.T) {
	originalFileName := redeemFilePath
	backupName := "redeem.json.bak"
	hasBackup := false
	if _, err := os.Stat(originalFileName); err == nil {
		os.Rename(originalFileName, backupName)
		hasBackup = true
	}
	defer func() {
		os.Remove(originalFileName)
		if hasBackup {
			os.Rename(backupName, originalFileName)
		}
	}()

	mockBot := &MockMessenger{}
	fetchedCodes := []model.RedeemCodeInfo{
		{Code: "OLD_CODE", Reward: "Old Reward"},
	}
	previouslySent := []string{"OLD_CODE"}

	err := processRedeemTask(mockBot, fetchedCodes, previouslySent)
	if err != nil {
		t.Fatalf("processRedeemTask failed: %v", err)
	}

	// 不應該發送訊息
	if len(mockBot.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(mockBot.Messages))
	}
}
