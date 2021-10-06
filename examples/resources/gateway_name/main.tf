terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
      version = "0.1.7"
    }
  }
}

provider "grid" {
}

# this data source is used to break circular dependency in cases similar to the following:
# vm: needs to know the domain in its init script
# gateway_name: needs the ip of the vm to use as backend.
# - the fqdn can be computed from grid_gateway_domain for the vm
# - the backend can reference the vm ip directly 
data "grid_gateway_domain" "domain" {
  node = 2
  name = "example2"
}

resource "grid_name_proxy" "p1" {
  node = 40
  name = "example2"
  backends = [format("http://137.184.106.152")]
  tls_passthrough = false
}
output "fqdn" {
    value = data.grid_gateway_domain.domain.fqdn
}
