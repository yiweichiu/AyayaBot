package model



// RedeemCode represents the structure of a single redeem code from the API.
type RedeemCode struct {
	Code       string                 `json:"code"`
	Reward     map[string]interface{} `json:"reward"`
	Status     string                 `json:"status"`
	ExpiryDate string                 `json:"expiry_date"`
	ImageURL   interface{}            `json:"image_url"` // Can be null or string
}

// RedeemCodeInfo represents a simplified structure of a single active redeem code.
type RedeemCodeInfo struct {
	Code   string
	Reward string
}
