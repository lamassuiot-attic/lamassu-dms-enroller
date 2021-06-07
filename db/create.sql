CREATE TABLE csr_store (
    id SERIAL,
    name TEXT,
    country TEXT,
    state TEXT,
    locality TEXT,
    organization TEXT,
    organization_unit TEXT,
    common_name TEXT,
    email TEXT,
    status TEXT,
    csrPath TEXT
);

CREATE TABLE ca_store (
    id INTEGER,
    status CHAR(1),
    expirationDate TEXT,
    revocationDate TEXT,
    serial TEXT,
    dn TEXT,
    certPath TEXT
);