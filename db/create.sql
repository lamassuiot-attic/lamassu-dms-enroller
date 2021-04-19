CREATE TABLE csr_store (
    id SERIAL,
    c TEXT,
    st TEXT,
    l TEXT,
    o TEXT,
    ou TEXT,
    cn TEXT,
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