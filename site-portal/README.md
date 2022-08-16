# FML Site Portal

Managing FATE jobs from each site.

## Build
```
make all
```
The generated deliverables are placed in the `output` folder.

## Build & Run Docker Images
* Modify `.env` file to change the image names, and then
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
## Run Pre-built Images
* Modify `.env` file to change the image names, and then
```
docker-compose pull
docker-compose up
```

> You may need to run `chmod a+w ./output/data/server/uploaded` to allow write permission to the uploaded data folder. Otherwise you may not be able to upload your local data.

## Enable HTTPS using docker-compose
### Create Site Portal Server and Client certificates
The server cert is used for accepting connection from users and fml-manager and the client cert is used for connecting to fml-manager.

**As an example, this guide uses StepCA CLI to sign and get the certificate.**

1. Make sure your StepCA CA server is running. Put your CA `ca.crt` into `tls/cert` folder. 
Use command below to get your CA cert.
```
step ca root
```
2. Then `cd` to `tls/cert` folder, run commands below to create certificates and keys (replace `<CommonName>`(e.g. 1.fate.org), `<CertValidTime>`(e.g. 8760h) with your site configuration):
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

### Run Site Portal
* `cd` to the project root folder
```
docker-compose -f docker-compose-https.yml up
```
* Open Site Portal with URL `https://<address>:8443`

## Deploy into Kubernetes
There are helms chart developed for installing Site Portal with the FATE exchange components together. Currently, it is used by the FedLCM service. Refer to the documents in the FedLCM.

## Getting Started

### 0. Prerequisites

* A FATE instance using Spark and Pulsar as backend.
* An existing FML Manager service to help to coordinate between other Site Portals.

### 1. Login Site Portal with predefined credentials
* If this portal is deployed by FedLCM service, then we can open it from the FedLCM's FATE cluster detail page.
* The predefined credentials for Site Portal are defined in the environment variables:
  * `SITEPORTAL_INITIAL_ADMIN_PASSWORD` for user "Admin", by default, `admin`.
  * `SITEPORTAL_INITIAL_USER_PASSWORD` for user "User", by default, `user`.
* You can change the password of you current user after login. Once changed, the above environment variable will no longer take effect.

### 2. Configure site information
Firstly, in the "Site Configuration" page, we need to configure the basic information of the site.
* Party ID should be what is configured during deploying of the FATE cluster
* Site Address and Port is the exposed service address and port that can be accessed from other service, like from FML Manager.
  * This depends on the type of the exposed service. When using FedLCM, such information can be found from the FedLCM page under the exposed service section.
* FML Manager endpoint: select to use `http` or `https`, then input the IP Address or FQDN for accessing the fml-manager
  * For example，the completed url is： `http://<fml-manager-service-ip>:<fml-manager-service-port>`
* FML Manager Server Name: When FML manager endpoint uses 'https', `Server Name` may be needed if 'FML Manager Endpoint' is an IP address. It's typical to use the Common Name or FQDN(if valid) of FML Manager as 'Server Name'. Site Portal uses 'FML Manager Server Name' to verify FML Manager's server certificate. If you leave it empty, server name will be 'FML Manager Endpoint' address.
  * For example： `fmlmanager.server.fate.org`
  * When FedLCM is used, such information can be found from the Federation's exchange detail page.
* FATE-Flow configuration is the service address of the fate flow service.
  * When FedLCM is used, we can simply input the follow info:
    * FATE-Flow address should be `fateflow`
    * HTTP port should be `9380`
  * If Site Portal is not deployed along with FATE cluster via KubeFATE, then users need to find the exposed ip and port of the FATE-Flow service. Currently it can only works with FATE v1.6.1 using Spark and Pulsar as backend.
* The Kubeflow configuration is for deploying horizontal models to KFServing system. They are optional and require an installation of MinIO and KFServing. Currently they are not used.
* After saving the configuration, click "register" button next to the FML Manager configuration section to register this service to the FML Manager.
  * In the future, if we have changed some site settings, the connection status to FML Manager will change to not connected. We need to register again to update our new site information to FML Manager.
  * This means if you want to disconnect ourselves from FML Manager, then you can clean the FML Manager endpoint settings and save again.

### 3. Upload local data in the "data management" page
* CSV files can be uploaded to the system via this page.
* Site Portal needs to upload the data to the FATE cluster, so the "Upload Job Status" field of each data is the status of the "uploading to FATE" job.
  * Only data in "Succeeded" status can be used in future FATE training and predicting jobs.

### 4. Create project in the "project management" page
Once created, this project can be viewed as a "local" project. Its information will not be sent to FML Manager, until we invite other parties to join this project.

### 5. Invite other parties
* Make sure other parties' FATE cluster and Site Portal are deployed and configured, in the same way above.
* In the participant management tab, view other registered participant and send invitation to others.
* The invited party's site portal page will list the invitation and users of that party can choose to join or reject.

### 6. Associate data
* All parties can associate their local data into the current project from the data management page.
* When data association is changed, all participating parties should see the updated information.

### 7. Start jobs
* In the "job management" tab, one can create new FATE job with other parties.
* Any joined party can initiate new jobs.
* We provide two modes to create FATE jobs: Drag-n-Drop and Json template.

### 8. Work with trained models
* Models can be viewed in the "model management" tab in project or "model management" page in the main page.
* They can be used in "prediction" type of job.
* Currently the publish function will return error as FATE v1.6.1 does not support such operation. This is a placeholder for future integration with newer FATE versions.

### 9. Other operations
* All parties can dismiss their own data association so it won't be used for futurre jobs.
* Project joining participant can leave the joined project if it no longer want to participate.
* Project managing participant can close the project if it is no longer need.
* The "User Management" page provides some configurations to set user permissions for accessing FATE Jupyter Notebook and FATEBoard. But currently it is not implemented yet. It is a placeholder for future integrations.
