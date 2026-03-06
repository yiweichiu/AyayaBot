package bd2redeem

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yiweichiu/AyayaBot/model"
)

func TestGetRedeemCodes(t *testing.T) {
	// ... (rest of the setup)

	tomorrow := time.Now().Add(24 * time.Hour).Format("2006/01/02")
	yesterday := time.Now().Add(-24 * time.Hour).Format("2006/01/02")

	mockCodes := []model.RedeemCode{
		{
			Code:   "PERMANENT_CODE",
			Status: "permanent",
			Reward: map[string]interface{}{"zh-Hant-TW": "永久獎勵"},
		},
		{
			Code:   "ACTIVE_CODE",
			Status: "active",
			Reward: map[string]interface{}{"zh-Hant-TW": "活動中獎勵"},
		},
		{
			Code:       "VALID_EXPIRY_CODE",
			Status:     "limited",
			ExpiryDate: tomorrow,
			Reward:     map[string]interface{}{"zh-Hant-TW": "未過期獎勵"},
		},
		{
			Code:       "EXPIRED_CODE",
			Status:     "limited",
			ExpiryDate: yesterday,
			Reward:     map[string]interface{}{"zh-Hant-TW": "已過期獎勵"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 驗證 API Key 是否正確傳遞
		if r.Header.Get("X-API-Key") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockCodes)
	}))
	defer server.Close()

	// 執行測試
	codes, err := GetRedeemCodes(context.Background(), server.URL, "test-api-key")

	// 驗證結果
	if err != nil {
		t.Fatalf("GetRedeemCodes failed: %v", err)
	}

	// 預期應該有 3 組代碼 (永久, 活動中, 未過期)
	if len(codes) != 3 {
		t.Errorf("Expected 3 active codes, got %d", len(codes))
	}

	expectedCodes := map[string]string{
		"PERMANENT_CODE":    "永久獎勵",
		"ACTIVE_CODE":       "活動中獎勵",
		"VALID_EXPIRY_CODE": "未過期獎勵",
	}

	for _, c := range codes {
		reward, ok := expectedCodes[c.Code]
		if !ok {
			t.Errorf("Unexpected code found: %s", c.Code)
			continue
		}
		if c.Reward != reward {
			t.Errorf("For code %s, expected reward %s, got %s", c.Code, reward, c.Reward)
		}
	}
}

func TestGetRedeemCodes_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	codes, err := GetRedeemCodes(context.Background(), server.URL, "wrong-key")
	if err == nil {
		t.Fatal("Expected error for unauthorized status, got nil")
	}
	if !strings.Contains(err.Error(), "API returned non-OK status") {
		t.Errorf("Expected error message to contain 'API returned non-OK status', got: %v", err)
	}
	if codes != nil {
		t.Errorf("Expected nil codes on error, got %v", codes)
	}
}
