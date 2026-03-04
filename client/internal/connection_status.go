package internal

type ConnectionStatus struct {
	Type      string                 `json:"type"`
	Connected bool                   `json:"connected"`
	Details   map[string]interface{} `json:"details"`
}
