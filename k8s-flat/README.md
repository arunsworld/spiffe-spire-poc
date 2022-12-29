# k8s-flat

## Characteristics

* One Spire server at the root - running in K8S
* Spire Controller Manager sidecar that keeps the entries up to date
* Spire agent daemonset for Workload API & workload attestation

## Configurations

* Namespace (`spire`)
* `06-spire-server.yml`
    * server.conf
        * trust_domain
        * ca_subject
        * clusters
        * use of k8s_psat NodeAttestor
        * use of sqlite3 database
    * spire-controller-manager - config.json
        * trustDomain
        * clusterName
        * ignoreNamespaces
* `08-spire-agent`
    * agent.conf
        * trust_domain
        * use of k8s_psat NodeAttestor
        * cluster
* `09-cluster-spiffe-config`
    * spiffeIDTemplate - Identity Naming Scheme

