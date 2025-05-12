package rate_limiter

type TokenBucket struct {
	Tokens     int   `json:"tokens"`
	LastUpdate int64 `json:"last_update"`
	Capacity   int   `json:"capacity"`
	Rate       int   `json:"rate"`
}
