# Leverage upstream CA signing authority store on disk

* Use-case - if validating entity doesn't have ability to integrate with SPIFFE bundle
* As ultimate signing authority is included SA the chain can be trusted

## Pre-requisites
* Load a secret called `ca-key-pair` into the spire namespace
* Assume this is an intermediate signing authority
* It should contain 3 files:
    * `tls.crt` - base64 encoded CA certificate
    * `tls.key` - base64 encoded CA key (intermediate)
    * `ca.crt` - base64 encoded root CA certificate