terraform {
  required_providers {
    grid = {
      source = "threefoldtechdev.com/providers/grid"
    }
  }
}

provider "grid" {
}


resource "grid_fqdn_proxy" "p1" {
  node = 7 
  name = "workloadname"
  fqdn = "remote.omar.grid.tf"
  backends = [format("https://137.184.106.152:443")]
  tls_passthrough = true
}

output "fqdn" {
    value = grid_fqdn_proxy.p1.fqdn
}
