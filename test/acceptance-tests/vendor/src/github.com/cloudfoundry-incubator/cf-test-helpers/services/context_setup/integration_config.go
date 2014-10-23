package context_setup

type IntegrationConfig struct {
	AppsDomain        string  `json:"apps_domain"`
	ApiEndpoint       string  `json:"api"`
	AdminUser         string  `json:"admin_user"`
	AdminPassword     string  `json:"admin_password"`
	SkipSSLValidation bool    `json:"skip_ssl_validation"`
	TimeoutScale      float64 `json:"timeout_scale"`
}
