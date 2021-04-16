package types

// ServiceRegisterInput
type ServiceRegisterInput struct {
	DomainName string `json:"DomainName"`
	IP         string `json:"IP"`
	Port       string `json:"Port"`
}

type ServiceRegisterOutput struct {
	IP   string `json:"IP"`
	Port string `json:"Port"`
}
