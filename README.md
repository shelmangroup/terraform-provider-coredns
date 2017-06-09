# terraform-coredns

## Example
```
provider "coredns" {
    etcd_endpoints = "http://127.0.0.1:2379"
    zones = "skydns.local"
}

resource "coredns_record" "covfefe" {
    fqdn = "covfefe.skydns.local"
    type = "A"
    rdata = [ "10.10.10.10" ]
    ttl = "60"
}
```
