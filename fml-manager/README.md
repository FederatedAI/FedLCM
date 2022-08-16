# FATE FML Manager

A service to manage federations between different FATE sites

## Build
```
make all
```
The generated deliverables are placed in the `output` folder.

## Build & Run Docker Image
* Modify `.env` file to change the image name, and then
```
set -a; source .env; set +a
make docker-build
```
* Optionally push the image
```
make docker-push
```
* Start the service
```
docker-compose up
```
## Run Pre-built Images from remote registry
* Modify `.env` file to change the image names, and then
```
docker-compose pull
docker-compose up
```
## Enable HTTPS using docker-compose
### Create FML Manager Server and Client certificates
The server cert is used for accepting connection from site portals and the client cert is used for connecting to site portals.

**As an example, this guide uses StepCA CLI to sign and get the certificate.**

1. Make sure your StepCA CA server is running. Put your CA `ca.crt` into `tls/cert` folder.
Use command below to get your CA cert.
```
step ca root
```
2. Then `cd` to `tls/cert` folder, run commands below to create certificates and keys (replace `<CommonName>`(e.g. fmlmanager.fate.org) and `<CertValidTime>`(e.g. 8760h) with your configuration):
```
step ca certificate <CommonName> --san localhost server.crt server.key --not-after=<CertValidTime>

step ca certificate <CommonName> client.crt client.key --not-after=<CertValidTime>
```
* For the server cert, the `localhost` SAN name is required because our sever may call itself via the localhost address.
* You can optionally add your other address and dns names as SANs in the command line.

**If you want to use other methods to generate the certificates and keys**
* For server.crt server.key, make sure to include `<CommonName>` and <SAN> `localhost`.
  1. If you have a usable FQDN, you can use it as your `<CommonName>`, and set <SAN> `localhost`
  2. If you don't have a usable FQDN, set your `<CommonName>` with your preference. Then append <SAN> with `localhost`.
* For client.crt client.key, make sure to include `<CommonName>`

### Run FML Manager
* `cd` to the project root folder
```
docker-compose -f docker-compose-https.yml up
```

## Deploy into Kubernetes
The are helms chart developed for installing fml-manager with the FATE exchange components together. Currently, it is used by the lifecycle-manager service. Refer to the documents in the lifecycle-manager.