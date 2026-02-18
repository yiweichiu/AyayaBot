package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type RedeemCode struct {
	Code        string                 `json:"code"`
	Reward      map[string]interface{} `json:"reward"`
	Status      string                 `json:"status"`
	ExpiryDate  string                 `json:"expiry_date"`
	ImageURL    interface{}            `json:"image_url"` // Can be null or string
}

type RedeemCodeInfo struct {
	Code   string
	Reward string
}

func GetRedeemCodes(apiURL, apiKey string) ([]RedeemCodeInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response body: %w", err)
	}


	var redeemCodes []RedeemCode
	err = json.Unmarshal(body, &redeemCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	var activeCodesInfo []RedeemCodeInfo
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
			activeCodesInfo = append(activeCodesInfo, RedeemCodeInfo{Code: rc.Code, Reward: rewardStr})
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
				activeCodesInfo = append(activeCodesInfo, RedeemCodeInfo{Code: rc.Code, Reward: rewardStr})
			}
		}
	}

	return activeCodesInfo, nil
}
