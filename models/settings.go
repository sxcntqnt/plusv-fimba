package models

type Settings struct {
	SettingsId                 uint   `json:"s_id"`
	SettingsCompanyName        string `json:"s_companyname"`
	SettingsAddress            string `json:"s_address"`
	SettingsInvoicePrefix      string `json:"s_inovice_prefix"`
	SettingsLogo               string `json:"s_logo"`
	SettingsPricePrefix        string `json:"s_price_prefix"`
	SettingsTermsAndCond       string `json:"s_inovice_termsandcondition"`
	SettingsInvoiceServiceName string `json:"s_inovice_servicename"`
	SettingsGoogleApiKey       string `json:"s_googel_api_key"`
}
