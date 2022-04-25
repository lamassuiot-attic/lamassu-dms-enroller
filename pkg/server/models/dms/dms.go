package dms

type DMS struct {
	Id                    string             `json:"id"`
	Name                  string             `json:"name"`
	SerialNumber          string             `json:"serial_number,omitempty"`
	KeyMetadata           PrivateKeyMetadata `json:"key_metadata"`
	Status                string             `json:"status"`
	CsrBase64             string             `json:"csr,omitempty"`
	CerificateBase64      string             `json:"crt,omitempty"`
	Subject               Subject            `json:"subject"`
	AuthorizedCAs         []string           `json:"authorized_cas,omitempty"`
	CreationTimestamp     string             `json:"creation_timestamp,omitempty"`
	ModificationTimestamp string             `json:"modification_timestamp,omitempty"`
}
type PrivateKeyMetadata struct {
	KeyType     string `json:"type"`
	KeyBits     int    `json:"bits"`
	KeyStrength string `json:"strength,omitempty"`
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
	Name        string             `json:"name"`
	Subject     Subject            `json:"subject"`
	KeyMetadata PrivateKeyMetadata `json:"key_metadata"`
	Url         string             `json:"url"`
}
type AuthorizedCAs struct {
	DmsId  string
	CaName string
}

const (
	PendingStatus  = "PENDING_APPROVAL"
	ApprovedStatus = "APPROVED"
	DeniedStatus   = "DENIED"
	RevokedStatus  = "REVOKED"
)
