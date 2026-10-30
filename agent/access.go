package agent

type AccessPrivateRequest struct {
	ShareToken      string   `dd:"share_token"`
	BindAddress     string   `dd:"bind_address"`
	AutoMode        bool     `dd:"auto_mode"`
	AutoAddress     string   `dd:"auto_address"`
	AutoStartPort   uint16   `dd:"auto_start_port"`
	AutoEndPort     uint16   `dd:"auto_end_port"`
	ResponseHeaders []string `dd:"response_headers"`
}

type access struct {
	frontendToken   string
	token           string
	bindAddress     string
	autoMode        bool
	autoAddress     string
	autoStartPort   uint16
	autoEndPort     uint16
	responseHeaders []string
}
