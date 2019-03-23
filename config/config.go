package config

type Config struct{
	ActiveBiddingServices int `json:"active_bidding_services"`
	CircuitBreaker BreakerConfig	`json:"circuit_breaker"`
}

type BreakerConfig struct{
	Enabled    bool      `json:"is_enabled"`
	ErrorThreshold int	 `json:"error_threshold"`
	SuccessThreshold int	`json:"success_threshold"`
	Timeout	int `json:"timeout"`
}