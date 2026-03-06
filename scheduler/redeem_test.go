package scheduler

import (
	"os"
	"strings"
	"testing"

	"github.com/yiweichiu/AyayaBot/model"
)

func TestProcessRedeemTask(t *testing.T) {
	const testRedeemFile = "redeem_test.json"
	defer os.Remove(testRedeemFile)

	mockBot := &MockMessenger{}
	
	fetchedCodes := []model.RedeemCodeInfo{
		{Code: "NEW_CODE_1", Reward: "Reward 1"},
		{Code: "OLD_CODE", Reward: "Old Reward"},
	}
	
	previouslySent := []string{"OLD_CODE"}

	err := processRedeemTask(mockBot, "test-channel", fetchedCodes, previouslySent, false, testRedeemFile)
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

	if strings.Contains(mockBot.Messages[0], "(<https://thebd2pulse.com/>)") {
		t.Errorf("Message should contain visible URL, got: %s", mockBot.Messages[0])
	}

	if strings.Contains(mockBot.Messages[0], "OLD_CODE") {
		t.Error("Message should not contain OLD_CODE")
	}

	// 驗證檔案是否正確儲存 (應包含所有目前的代碼)
	savedCodes, err := loadRedeemCodesFromFile(testRedeemFile)
	if err != nil {
		t.Fatalf("Failed to load saved codes: %v", err)
	}
	if len(savedCodes) != 2 {
		t.Errorf("Expected 2 saved codes, got %d", len(savedCodes))
	}
}

func TestProcessRedeemTask_HideEmbed(t *testing.T) {
	const testRedeemFile = "redeem_hide_test.json"
	defer os.Remove(testRedeemFile)

	mockBot := &MockMessenger{}
	fetchedCodes := []model.RedeemCodeInfo{
		{Code: "NEW_CODE_1", Reward: "Reward 1"},
	}
	previouslySent := []string{}

	err := processRedeemTask(mockBot, "test-channel", fetchedCodes, previouslySent, true, testRedeemFile)
	if err != nil {
		t.Fatalf("processRedeemTask failed: %v", err)
	}

	if !strings.Contains(mockBot.Messages[0], "(<https://thebd2pulse.com/>)") {
		t.Errorf("Message should contain hidden URL: %s", mockBot.Messages[0])
	}
}

func TestProcessRedeemTask_NoNewCodes(t *testing.T) {
	const testRedeemFile = "redeem_none_test.json"
	defer os.Remove(testRedeemFile)

	mockBot := &MockMessenger{}
	fetchedCodes := []model.RedeemCodeInfo{
		{Code: "OLD_CODE", Reward: "Old Reward"},
	}
	previouslySent := []string{"OLD_CODE"}

	err := processRedeemTask(mockBot, "test-channel", fetchedCodes, previouslySent, false, testRedeemFile)
	if err != nil {
		t.Fatalf("processRedeemTask failed: %v", err)
	}

	// 不應該發送訊息
	if len(mockBot.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(mockBot.Messages))
	}
}
