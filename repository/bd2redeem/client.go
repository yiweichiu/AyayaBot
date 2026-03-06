package bd2redeem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yiweichiu/AyayaBot/model" // Import model package
)

// GetRedeemCodes fetches redeem codes from the BD2 Pulse API.
func GetRedeemCodes(ctx context.Context, apiURL, apiKey string) ([]model.RedeemCodeInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response body: %w", err)
	}

	var redeemCodes []model.RedeemCode
	err = json.Unmarshal(body, &redeemCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	var activeCodesInfo []model.RedeemCodeInfo
	currentDate := time.Now()
	for _, rc := range redeemCodes {
		// Determine reward string
		rewardStr := ""
		if rewardVal, ok := rc.Reward["zh-Hant-TW"]; ok {
			if s, isString := rewardVal.(string); isString {
				rewardStr = s
			}
		}

		// If status is "permanent" or "active", always include it
		if rc.Status == "permanent" || rc.Status == "active" {
			activeCodesInfo = append(activeCodesInfo, model.RedeemCodeInfo{Code: rc.Code, Reward: rewardStr})
			continue
		}

		// Parse expiry date if available
		if rc.ExpiryDate != "" {
			expiry, err := time.Parse("2006/01/02", rc.ExpiryDate)
			if err != nil {
				// Try parsing with single-digit month/day format
				expiry, err = time.Parse("2006/1/2", rc.ExpiryDate)
				if err != nil {
					log.Printf("Error parsing expiry date %s for code %s: %v", rc.ExpiryDate, rc.Code, err)
					continue // Skip if date cannot be parsed
				}
			}
			// Compare expiry date with current date. Add if not expired yet (or expires today)
			// Adding 23 hours, 59 minutes, 59 seconds to expiry to ensure it includes the whole day
			if expiry.Add(23*time.Hour + 59*time.Minute + 59*time.Second).After(currentDate) {
				activeCodesInfo = append(activeCodesInfo, model.RedeemCodeInfo{Code: rc.Code, Reward: rewardStr})
			}
		}
	}

	return activeCodesInfo, nil
}
