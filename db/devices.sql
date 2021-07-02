CREATE TABLE device_information (
    id TEXT PRIMARY KEY,
    alias TEXT,
    status TEXT,
    dms_id int,
    country TEXT,
    state TEXT,
    locality TEXT,
    organization TEXT,
    organization_unit TEXT,
    common_name TEXT,
    key_stregnth TEXT,
    key_type TEXT,
    key_bits int,
    creation_ts timestamp,
    current_cert_serial_number TEXT
);

CREATE TABLE device_logs (
    id TEXT PRIMARY KEY,
    creation_ts timestamp,
    device_uuid TEXT,
    log_type TEXT,
    log_message TEXT
);

CREATE TABLE device_certificates_history (
    serial_number TEXT PRIMARY KEY,
    device_uuid TEXT,
    issuer_serial_number TEXT,
    issuer_name TEXT,
    status TEXT,
    creation_ts timestamp
);