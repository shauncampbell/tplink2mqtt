package homeassistant

type connection []string

type deviceConfiguration struct {
	Name         string     `json:"name"`
	CommandTopic string     `json:"command_topic"`
	StateTopic   string     `json:"state_topic"`
	Device       deviceInfo `json:"device"`
	UniqueID     string     `json:"unique_id"`
}

type deviceInfo struct {
	Manufacturer string       `json:"manufacturer"`
	Connections  []connection `json:"connections"`
	Identifiers  []string     `json:"identifiers"`
	Model        string       `json:"model"`
	Name         string       `json:"name"`
}
