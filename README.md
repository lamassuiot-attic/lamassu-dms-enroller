<a href="https://www.lamassu.io/">
  <img src="logo.png" alt="Lamassu logo" title="Lamassu" align="right" height="80" />
</a>

Lamassu
=======
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](http://www.mozilla.org/MPL/2.0/index.txt)

[Lamassu](https://www.lamassu.io) project is a Public Key Infrastructure (PKI) for the Internet of Things (IoT).

## Enroller

Enroller implements Simple Certificate Enrollment Protocol (SCEP). A protocol designed to make the issuing of digital certificates as scalable as possible.

### Project Structure
The Enroller is composed of two services:
1. Enroller: Main service of the project. Performs the pairing operations with a [Device Manufacturing System](https://github.com/lamassuiot/device-manufacturing-system). The Device Manufacturing System submmits a CSR (Certificate Signing Request) and the Enroller admin manually accepts (creating a signed certificate), denys the CSR or revokes a previously signed certificate.
2. SCEP: This service provides some useful operations (list and revoke) to check the lifecycle of the certificates signed by Lamassu PKI and provided to devices via SCEP protocol.

Each service has its own application directory in `cmd/` and libraries in `pkg/`.

## Installation
To compile the Enroller follow the next steps:
1. Clone the repository: `go get github.com/lamassuiot/enroller`.
2. Run the Enroller service compilation script: `cd src/github.com/lamassuiot/enroller/cmd/enroller && ./release.sh`
3. Run the SCEP service compilation script: `cd src/github.com/lamassuiot/enroller/cmd/scep && ./release.sh`

The binaries will be compiled in the `build/` directory.

## Usage
Each service of the Enroller should be configured with some environment variables.
**Enroller service**
```
ENROLLER_PORT=8085 //Enroller service port.
ENROLLER_POSTGRESUSER=<POSTGRESUSER> //Enroller DB user.
ENROLLER_POSTGRESPASSWORD=<POSTGRESPASSWORD> //Enroller DB password.
ENROLLER_POSTGRESDB=enrollerdb // Enroller DB name.
ENROLLER_POSTGRESHOSTNAME=enrollerdb //Enroller DB server hostname.
ENROLLER_POSTGRESPORT=5432 //Enroller DB port.
ENROLLER_CONSULPROTOCOL=https //Consul server protocol.
ENROLLER_CONSULHOST=consul //Consul server host.
ENROLLER_CONSULPORT=8501 //Consul server port.
ENROLLER_CONSULCA=consul.crt //Consul server certificate CA to trust it.
ENROLLER_HOMEPATH=/var/lib/csrs //File system path to store CSR files.
ENROLLER_ENROLLERUIHOST=enrollerui //UI host (for CORS 'Access-Control-Allow-Origin' header).
ENROLLER_EROLLERUIPORT=443 //UI port (for CORS 'Access-Control-Allow-Origin' header).
ENROLLER_ENROLLERUIPROTOCOL=https //UI protocol (for CORS 'Access-Control-Allow-Origin' header).
ENROLLER_KEYCLOAKHOSTNAME=keycloak //Keycloak server hostname.
ENROLLER_KEYCLOAKPORT=8443 //Keycloak server port.
ENROLLER_KEYCLOAKPROTOCOL=https //Keycloak server protocol.
ENROLLER_KEYCLOAKREALM=<KEYCLOAK_REALM> //Keycloak realm configured.
ENROLLER_KEYCLOAKCA=keycloak.crt //Keycloak server certificate CA to trust it.
ENROLLER_CACERTFILE=enroller_admin.crt //Enroller admin certificate used to sign Device Manufacturing Systems' CSRs.
ENROLLER_CAKEYFILE=enroller_admin.key //Enroller admin key used to sign Device Manufacturing Systems' CSRs.
ENROLLER_CERTFILE=enroller.crt //Enroller service certificate.
ENROLLER_KEYFILE=enroller.key //Enroller service key.
ENROLLER_OCSPSERVER=https://ocsp:9098 //OCSP Server address for including it in signed certificates.
```
**SCEP service**
```
SCEP_PORT=8086
SCEP_POSTGRESUSER=<POSTGRESUSER> //SCEP DB user.
SCEP_POSTGRESDB=scepdb //SCEP DB name.
SCEP_POSTGRESPORT=5432 //SCEP DB port.
SCEP_POSTGRESPASSWORD=<POSTGRESPASSWORD> //SCEP DB password.
SCEP_POSTGRESHOSTNAME=scepdb //SCEP DB hostname.
SCEP_CONSULPROTOCOL=https //Consul server protocol.
SCEP_CONSULHOST=consul //Consul server host.
SCEP_CONSULPORT=8501 //Consul server port.
SCEP_CONSULCA=consul.crt //Consul server certificate CA to trust it.
SCEP_ENROLLERUIHOST=enrollerui //UI host (for CORS 'Access-Control-Allow-Origin' header).
SCEP_ENROLLERUIPORT=443 //UI port (for CORS 'Access-Control-Allow-Origin' header).
SCEP_ENROLLERUIPROTOCOL=https //UI protocol (for CORS 'Access-Control-Allow-Origin' header).
SCEP_KEYCLOAKHOSTNAME=keycloak //Keycloak server hostname.
SCEP_KEYCLOAKPORT=8443 //Keycloak server port.
SCEP_KEYCLOAKPROTOCOL=https //Keycloak server protocol.
SCEP_KEYCLOAKREALM=<KEYCLOAK_REALM> //Keycloak realm configured.
SCEP_KEYCLOAKCA=keycloak.crt //Keycloak server certificate CA to trust it.
SCEP_CERTFILE=scep.crt //SCEP service certificate.
SCEPKEYFILE=scep.key //SCEP service key.
```
The prefixes `(ENROLLER_)` and `(SCEP_)` used to declare the environment variables can be changed in `cmd/enroller/main.go` and `cmd/scep/main.go`:
```
cfg, err := configs.NewConfig("enroller")
cfg, err := configs.NewConfig("scep")
```
For more information about the environment variables declaration check `pkg/enroller/configs` and  `pkg/scep/configs`.

## Docker
The recommended way to run [Lamassu](https://www.lamassu.io) is following the steps explained in [lamassu-compose](https://github.com/lamassuiot/lamassu-compose) repository. However, each component can be run separately in Docker following the next steps.
**Enroller service**
```
docker image build -t lamassuiot/lamassu-enroller:latest -f Dockerfile.enroller .
docker run -p 8085:8085
  --env ENROLLER_PORT=8085 
  --env ENROLLER_POSTGRESUSER=<POSTGRESUSER>
  --env ENROLLER_POSTGRESPASSWORD=<POSTGRESPASSWORD>
  --env ENROLLER_POSTGRESDB=enrollerdb
  --env ENROLLER_POSTGRESHOSTNAME=enrollerdb
  --env ENROLLER_POSTGRESPORT=5432
  --env ENROLLER_CONSULPROTOCOL=https
  --env ENROLLER_CONSULHOST=consul
  --env ENROLLER_CONSULPORT=8501
  --env ENROLLER_CONSULCA=consul.crt
  --env ENROLLER_HOMEPATH=/var/lib/csrs
  --env ENROLLER_ENROLLERUIHOST=enrollerui
  --env ENROLLER_EROLLERUIPORT=443
  --env ENROLLER_ENROLLERUIPROTOCOL=https
  --env ENROLLER_KEYCLOAKHOSTNAME=keycloak
  --env ENROLLER_KEYCLOAKPORT=8443
  --env ENROLLER_KEYCLOAKPROTOCOL=https
  --env ENROLLER_KEYCLOAKREALM=<KEYCLOAK_REALM>
  --env ENROLLER_KEYCLOAKCA=keycloak.crt
  --env ENROLLER_CACERTFILE=enroller_admin.crt
  --env ENROLLER_CAKEYFILE=enroller_admin.key
  --env ENROLLER_CERTFILE=enroller.crt
  --env ENROLLER_KEYFILE=enroller.key
  --env ENROLLER_OCSPSERVER=https://ocsp:9098
  lamassuiot/enroller:latest
```
**SCEP service**
```
docker image build -t lamassuiot/lamassu-enroller-scep:latest -f Dockerfile.scep .
docker run -p 8086:8086
  --env SCEP_PORT=8086
  --env SCEP_POSTGRESUSER=<POSTGRESUSER>
  --env SCEP_POSTGRESDB=scepdb
  --env SCEP_POSTGRESPORT=5432
  --env SCEP_POSTGRESPASSWORD=<POSTGRESPASSWORD>
  --env SCEP_POSTGRESHOSTNAME=scepdb
  --env SCEP_CONSULPROTOCOL=https
  --env SCEP_CONSULHOST=consul
  --env SCEP_CONSULPORT=8501
  --env SCEP_CONSULCA=consul.crt
  --env SCEP_ENROLLERUIHOST=enrollerui
  --env SCEP_ENROLLERUIPORT=443
  --env SCEP_ENROLLERUIPROTOCOL=https
  --env SCEP_KEYCLOAKHOSTNAME=keycloak 
  --env SCEP_KEYCLOAKPORT=8443
  --env SCEP_KEYCLOAKPROTOCOL=https
  --env SCEP_KEYCLOAKREALM=<KEYCLOAK_REALM>
  --env SCEP_KEYCLOAKCA=keycloak.crt
  --env SCEP_CERTFILE=scep.crt
  --env SCEPKEYFILE=scep.key
  lamassuiot/lamassu-enroller-scep:latest
```

## Kubernetes
[Lamassu](https://www.lamassu.io) can be run in Kubernetes deploying the objects defined in `k8s/` directory. `provision-k8s.sh` script provides some useful guidelines and commands to deploy the objects in a local [Minikube](https://github.com/kubernetes/minikube) Kubernetes cluster.
