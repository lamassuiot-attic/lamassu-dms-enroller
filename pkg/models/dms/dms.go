package dms

type DMS struct {
	Id           int    `json:"id"`
	Name         string `json:"dms_name"`
	SerialNumber string `json:"serial_number"`
	KeyType      string `json:"key_type"`
	KeyBits      int    `json:"key_bits"`
	Status       string `json:"status"`
	CsrBase64    string `json:"csr,omitempty"`
}

type DmsCreationForm struct {
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
	Url                    string `json:"url"`
}

const (
	PendingStatus  = "PENDIG_APPROVAL"
	ApprovedStatus = "APPROVED"
	DeniedStatus   = "DENIED"
	RevokedStatus  = "REVOKED"
)
