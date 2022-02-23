package dms

type DMS struct {
	Id               int                `json:"id"`
	Name             string             `json:"dms_name"`
	SerialNumber     string             `json:"serial_number"`
	KeyMetadata      PrivateKeyMetadata `json:"key_metadata"`
	Status           string             `json:"status"`
	CsrBase64        string             `json:"csr,omitempty"`
	CerificateBase64 string             `json:"crt"`
	Subject 		 Subject            `json:"subject"`
}
type PrivateKeyMetadata struct {
	KeyType string `json:"type"`
	KeyBits int    `json:"bits"`
}
type Subject struct {
	CN string `json:"common_name"`
	O  string `json:"organization"`
	OU string `json:"organization_unit"`
	C  string `json:"country"`
	ST string `json:"state"`
	L  string `json:"locality"`
}
type DmsCreationForm struct {
	Name        string             `json:"dms_name"`
	Subject     Subject            `json:"subject"`
	KeyMetadata PrivateKeyMetadata `json:"key_metadata"`
	Url         string             `json:"url"`
}

const (
	PendingStatus  = "PENDIG_APPROVAL"
	ApprovedStatus = "APPROVED"
	DeniedStatus   = "DENIED"
	RevokedStatus  = "REVOKED"
)
