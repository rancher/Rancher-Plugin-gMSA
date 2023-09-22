# Rancher gMSA Account Provider

The account provider is a `hostProcess` container which runs an HTTP/s API using [Gin](https://github.com/gin-gonic/gin). The account provider only listens on `localhost` and does _not_ use a static port. The API offers a single endpoint (`/provider`), which is responsible for retrieving the contents of Kubernetes secrets and relaying them to the caller. In normal operation, the server enforces [mTLS](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/#:~:text=Mutual%20TLS%2C%20or%20mTLS%20for,TLS%20certificates%20provides%20additional%20verification) (using TLSv1.2) and automatically writes the required certificates to disk as well as imports them into the Windows host. 

# Detailed Host Modifications

During initialization the following directories are written to disk: 

+ `/var/lib/rancher/gmsa`
  + This is the root directory which contains all certificates and metadata for every instance of the Rancher gMSA Account Provider
+ `/var/lib/rancher/gmsa/<NAMESPACE>`
  + This is a subdirectory named according to the namespace the Account Provider is deployed to in Kubernetes. Each instance of the Account Provider will be allocated its own subdirectory. 
+ `/var/lib/rancher/gmsa/<NAMESPACE>/port.txt`
  + This is a simple text file containing the port number the Account Provider is currently listening on. It is used by the Rancher CCG Plugin to craft its HTTP/s requests. 
+ `/var/lib/rancher/gmsa/<NAMESPACE>/ssl`
  + This is a base directory containing all certificates used by both the Account Provider and the Rancher gMSA CCG Plugin 
+ `/var/lib/rancher/gmsa/<NAMESPACE>/ssl/client`
  + This directory contains the `pfx` file used by the Rancher CCG Plugin when crafting requests using mTLs. This file is automatically imported into the Windows certificate store.
+ `/var/lib/rancher/gmsa/<NAMESPACE>/ssl/server`
  + This directory contains the server certificate and certificate authority certificate used by the Account Provider API. These files are automatically imported into the Windows certificate store.
+ `/var/lib/rancher/gmsa/<NAMESPACE>/ssl/ca`
  + This directory contains the root certificate authority certificates. These files are automatically imported into the Windows certificate store. 

# Testing Steps

## End to End 
The Account Provider is complemented by the CCG Plugin, and as such end-to-end testing will need to have both components deployed to the cluster. Additionally, in order for the CCG Plugin to function correctly, a `GMSACredntialSpec` will need to be configured and modified to cover specific test cases.   

### Scenarios

#### Valid `GMSACredentialSpec`

`a-test-namespace:this-is-a-test`

0. Ensure that the GMSA webhook and Rancher gMSA Plugin Installer charts are both installed onto your cluster
1. If not yet deployed, deploy the Account Provider chart into the `a-test-namespace` namespace
2. If not yet created, create a new secret within the `a-test-namespace` titled `this-is-a-test`, ensure the secret is `Opaque` and has `username`, `password`, and `domainName` fields. Each field should have a valid value given your Active Directory instance, and should be base64 encoded. 
3. Configure a GMSACredentialSpec for the desired gMSA account and ensure that the `PluginInput` field has the following content: `a-test-namespace:this-is-a-test`
4. Deploy a workload onto the cluster which leverages a gMSA account 
5. Ensure that the workload successfully becomes ready
6. Ensure that the Account Provider API logs the request made by the CCG plugin, and that no errors are observed 


#### Incorrect namespace in `GMSACredentialSpec`

0. Ensure that the GMSA webhook and Rancher gMSA Plugin Installer charts are both installed onto your cluster
1. If not yet deployed, deploy the Account Provider chart into the `a-test-namespace` namespace
3. If not yet created, create a new secret within the `a-test-namespace` titled `this-is-a-test`, ensure the secret is `Opaque` and has `username`, `password`, and `domainName` fields. Each field should have a valid value given your Active Directory instance, and should be base64 encoded.
4. Configure / reconfigure a GMSACredentialSpec, ensure that the following is used for the `PluginInput`: `bad-namespace:this-is-a-test`
5. Deploy a workload onto the cluster which leverages the GMSACredentialSpec
6. Ensure that the Account Provider API logs the request
7. Ensure that the Account Provider API logs the error message encountered when retrieving a non-existent secret
8. Ensure that the workload does not become ready 

#### Missing Secret name in `GMSACredentialSpec`

0. Ensure that the GMSA webhook and Rancher gMSA Plugin Installer charts are both installed onto your cluster
1. If not yet deployed, deploy the Account Provider chart into the `a-test-namespace` namespace
3. If not yet created, create a new secret within the `a-test-namespace` titled `this-is-a-test`, ensure the secret is `Opaque` and has `username`, `password`, and `domainName` fields. Each field should have a valid value given your Active Directory instance, and should be base64 encoded.
4. Configure / reconfigure a GMSACredentialSpec, ensure that the following is used for the `PluginInput`: `a-test-namespace:`
5. Deploy a workload onto the cluster which leverages the GMSACredentialSpec
6. Ensure that the Account Provider API logs a 404 response
7. Ensure that the workload does not become ready 
8. Check the event log on the windows worker, ensure that the CCG event log includes error messages relating to the CCG plugin

#### Incorrect Plugin Input Format name in `GMSACredentialSpec`

0. Ensure that the GMSA webhook and Rancher gMSA Plugin Installer charts are both installed onto your cluster
1. If not yet deployed, deploy the Account Provider chart into the `a-test-namespace` namespace
3. If not yet created, create a new secret within the `a-test-namespace` titled `this-is-a-test`, ensure the secret is `Opaque` and has `username`, `password`, and `domainName` fields. Each field should have a valid value given your Active Directory instance, and should be base64 encoded.
4. Configure / reconfigure a GMSACredentialSpec, ensure that the following is used for the `PluginInput`: `not-a-real-format`
5. Deploy a workload onto the cluster which leverages the GMSACredentialSpec
6. Ensure that the workload does not become ready
7. Check the event log on the windows worker, ensure that the CCG event log includes error messages relating to the CCG plugin

#### Incorrect / Missing Plugin GUID in `GMSACredentialSpec`

0. Ensure that the GMSA webhook and Rancher gMSA Plugin Installer charts are both installed onto your cluster
1. If not yet deployed, deploy the Account Provider chart into the `a-test-namespace` namespace
3. If not yet created, create a new secret within the `a-test-namespace` titled `this-is-a-test`, ensure the secret is `Opaque` and has `username`, `password`, and `domainName` fields. Each field should have a valid value given your Active Directory instance, and should be base64 encoded.
4. Configure / reconfigure a GMSACredentialSpec, ensure that the following is used for the `PluginInput`: `a-test-namespace:this-is-a-test`, and that the `PluginGUID` field is left empty
5. Deploy a workload onto the cluster which leverages the GMSACredentialSpec
6. Ensure that the workload does not become ready
7. Check the event log on the windows worker, ensure that the CCG event log includes error messages relating to the CCG plugin