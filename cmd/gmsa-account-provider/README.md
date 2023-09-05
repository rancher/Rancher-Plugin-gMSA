`rancher-gmsa-account-provider`
===

The gMSA Account Provider is responsible for hosting an HTTP/s API which enables a Container Credential Guard plugin to retrieve Kubernetes secrets containing the login information for an Active Directory Domain. In normal operation, this application is expected to run as a Kubernetes Pod; running `rancher-gmsa-account-provider` as a standalone binary on non-windows systems comes with functional limitations. 

## Flags and Environment Variables

+ `--namespace`, `NAMESPACE`
  + Type: `string`, `required`
  + Description: The namespace which contains the impersonation account secret
+ `--disable-mtls`, `DISABLE_MTLS`
  + Type: `bool`, `optional`
  + Description: Disables mTLS checks for the API. This can be useful for debugging the API, but should not be used in production. 
+ `--kubeconfig`, `KUBECONFIG`
  + Type: `string`, `optional`
  + Description: The KubeConfig which should be used to retrieve the Kubernetes secret containing the impersonation account credentials. Must be manually set if running outside a Kubernetes pod.
+ `--skip-artifacts`, `SKIP_ARTIFACTS`
  + TYPE: `bool`, `optional`
  + Description: Prevents the API from writing any files to disk, this is useful for development environments only. Implicitly disables mTLS support. Must be enabled if running outside a Kubernetes pod, or in an environment where artifacts cannot be written to the host. 
