# Leverage upstream CA signing authority store on disk

* Use-case - if validating entity doesn't have ability to integrate with SPIFFE bundle
* As ultimate signing authority is included SA the chain can be trusted

## Pre-requisites
* Load a secret called `ca-key-pair` into the spire namespace
* It should contain 2 files:
    * `tls.crt` - base64 encoded CA certificate
    * `tls.key` - base64 encoded CA key