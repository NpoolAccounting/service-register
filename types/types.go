package types

// ServiceRegisterInput
type ServiceRegisterInput struct {
	UserName   string `json:"UserName"`
	Password   string `json:"password"`
	DomainName string `json:"DomainName"`
	PublicIP   string `json:"PublicIP"`
	IP         string `json:"IP"`
	Port       string `json:"Port"`
}

type ServiceRegisterOutput struct {
	IP   string `json:"IP"`
	Port string `json:"Port"`
}
