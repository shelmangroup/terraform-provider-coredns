# terraform-coredns

## Example
```
provider "coredns" {
    etcd_endpoints = "http://127.0.0.1:2379"
    zones = "skydns.local"
}

resource "coredns_record" "foo" {
    fqdn = "foo.skydns.local"
    type = "A"
    rdata = [ "10.10.10.10", "10.10.10.20" ]
    ttl = "60"
}

resource "coredns_record" "bar" {
    fqdn = "bar.skydns.local"
    type = "CNAME"
    rdata = [ "${coredns_record.foo.hostname}" ]
    ttl = "60"
}
```
