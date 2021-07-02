package device

type Device struct {
	Id                      string `json:"id"`
	Alias                   string `json:"alias"`
	Status                  string `json:"status,omitempty"`
	DmsId                   int    `json:"dms_id"`
	Country                 string `json:"country"`
	State                   string `json:"state"`
	Locality                string `json:"locality"`
	Organization            string `json:"organization"`
	OrganizationUnit        string `json:"organization_unit"`
	CommonName              string `json:"common_name"`
	KeyType                 string `json:"key_type"`
	KeyBits                 int    `json:"key_bits"`
	KeyStrength             string `json:"key_strength"`
	CreationTimestamp       string `json:"creation_timestamp,omitempty"`
	CurrentCertSerialNumber string `json:"current_cert_serial_number"`
}

type DeviceCertHistory struct {
	DeviceId           string `json:"device_id"`
	SerialNumber       string `json:"serial_number"`
	IssuerSerialNumber string `json:"issuer_serial_number"`
	IsuuerName         string `json:"issuer_name"`
	Status             string `json:"status"`
	CreationTimestamp  string `json:"creation_timestamp"`
}

type DeviceCert struct {
	DeviceId     string `json:"device_id"`
	SerialNumber string `json:"serial_number"`
	CAName       string `json:"issuer_name"`
	Status       string `json:"status"`
	CRT          string `json:"crt"`
	Country      string `json:"country"`
	State        string `json:"state"`
	Locality     string `json:"locality"`
	Org          string `json:"organization"`
	OrgUnit      string `json:"organization_unit"`
	CommonName   string `json:"common_name"`
	ValidFrom    string `json:"valid_from"`
	ValidTo      string `json:"valid_to"`
}

type DeviceLog struct {
	Id         string `json:"id"`
	DeviceId   string `json:"device_id"`
	LogType    string `json:"log_type"`
	LogMessage string `json:"log_message"`
	Timestamp  string `json:"timestamp"`
}

type Devices struct {
	Devices []Device `json:"-"`
}

type DeviceLogs struct {
	Logs []DeviceLog `json:"-"`
}
type DeviceCertsHistory struct {
	DeviceCertHistory []DeviceCertHistory `json:"-"`
}

const ( // Device status
	DevicePendingProvision = "PENDING_PROVISION"
	DeviceProvisioned      = "DEVICE_PROVISIONED"
	DeviceCertRevoked      = "CERT_REVOKED"
	DeviceCertExpired      = "CERT_EXPIRED"
	DeviceDecommisioned    = "DEVICE_DECOMMISIONED"
)

const ( // Device Logs types
	LogDeviceCreated       = "LOG_DEVICE_CREATED"
	LogPendingProvision    = "LOG_PENDING_PROVISION"
	LogProvisioned         = "LOG_PROVISIONED"
	LogCertRevoked         = "LOG_CERT_REVOKED"
	LogCertExpired         = "LOG_CERT_EXPIRED"
	LogDeviceDecommisioned = "LOG_DEVICE_DECOMMISIONED"
)

const ( // Cert History status
	CertHistoryActive  = "ACTIVE"
	CertHistoryExpired = "EXPIRED"
	CertHistoryRevoked = "REVOKED"
)
