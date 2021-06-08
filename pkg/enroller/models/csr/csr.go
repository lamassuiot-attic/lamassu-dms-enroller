package csr

type CSR struct {
	Id                     int    `json:"id"`
	Name                   string `json:"dms_name"`
	CountryName            string `json:"country"`
	StateOrProvinceName    string `json:"state"`
	LocalityName           string `json:"locality"`
	OrganizationName       string `json:"organization"`
	OrganizationalUnitName string `json:"organization_unit,omitempty"`
	CommonName             string `json:"common_name"`
	EmailAddress           string `json:"mail,omitempty"`
	Status                 string `json:"status"`
	CsrFilePath            string `json:"csrpath,omitempty"`
}

type CSRForm struct {
	Name                   string `json:"dms_name"`
	CountryName            string `json:"country"`
	StateOrProvinceName    string `json:"state"`
	LocalityName           string `json:"locality"`
	OrganizationName       string `json:"organization"`
	OrganizationalUnitName string `json:"organization_unit,omitempty"`
	CommonName             string `json:"common_name"`
	EmailAddress           string `json:"mail,omitempty"`
	KeyType                string `json:"key_type"`
	KeyBits                int    `json:"key_bits"`
}

type CSRs struct {
	CSRs []CSR `json:"-"`
}

const (
	PendingStatus  = "NEW"
	ApprovedStatus = "APPROVED"
	DeniedStatus   = "DENIED"
	RevokedStatus  = "REVOKED"
)
