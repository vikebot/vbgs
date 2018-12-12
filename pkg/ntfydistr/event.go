package ntfydistr

type event struct {
	Type  string      `json:"type"`
	Obj   interface{} `json:"obj"`
	Unixn int64       `json:"unixn"`
}
