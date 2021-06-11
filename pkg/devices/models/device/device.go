package device

type Device struct {
	Id                string `json:"id"`
	Alias             string `json:"alias"`
	Status            string `json:"status,omitempty"`
	DmsId             string `json:"dms_id"`
	Country           string `json:"country"`
	State             string `json:"state"`
	Locality          string `json:"locality"`
	Organization      string `json:"organization"`
	OrganizationUnit  string `json:"organization_unit"`
	CommonName        string `json:"common_name"`
	KeyType           string `json:"key_type"`
	KeyBits           int    `json:"key_bits"`
	CreationTimestamp string `json:"creation_timestamp,omitempty"`
}

type DeviceCertHistory struct {
	DeviceId           string `json:"device_id"`
	SerialNumber       string `json:"serial_number"`
	IssuerSerialNumber string `json:"issuer_serial_number"`
	IsuuerName         string `json:"issuer_name"`
	Status             string `json:"status"`
	CreationTimestamp  string `json:"creation_timestamp"`
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
	DeviceCreated       = "DEVICE_CREATED"
	DeviceDecomissioned = "DEVICE_DECOMISSIONED"
)

const ( // Device Logs types
	LogPendingProvisionStatus = "PENDING_PROVISION"
	LogProvisionedStatus      = "PROVISIONED"
)

const ( // Cert History status
	CertHistoryActive  = "ACTIVE"
	CertHistoryRevoked = "REVOKED"
)
