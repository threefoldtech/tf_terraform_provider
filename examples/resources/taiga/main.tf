terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
    }
  }
}

provider "grid" {
}

resource "grid_network" "net2" {
  nodes         = [2]
  ip_range      = "10.1.0.0/16"
  name          = "network1"
  description   = "newer network"
  add_wg_access = true
}

resource "grid_deployment" "node1" {
  node         = 2
  network_name = grid_network.net2.name
  ip_range     = lookup(grid_network.net2.nodes_ip_range, 2, "")
  disks {
    name        = "data0"
    # will hold images, volumes etc. modify the size according to your needs
    size        = 100
    description = "volume holding docker data"
  }
  vms {
    name        = "taiga"
    flist       = "https://hub.grid.tf/tf-official-apps/grid3_taiga_docker-latest.flist"
    entrypoint  = "/sbin/zinit init"
    cpu         = 0
    memory      = 8096
    rootfs_size = 51200
    mounts {
      disk_name   = "data0"
      mount_point = "/var/lib/docker"
    }
    env_vars = {
      SSH_KEY     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCnJiB+sXPfqMKx6g67qOONqD0Kme08OKPtvacRwP6o0T8b404YaBvJaAEnsTgNE5A8vl15LBHdI94MdNoNwV7xT9HWWlw6hQ8PzN0e6z5M5bsKH6R6cKVbg8iYUESWWkRBr8iN10KDmbysyIbpr1QVIUoAbET4XpkJMxw46L8ClBsfPY5YFd1Bdd1oLwHJD4+cZQdZf9iSV4EtVfOfgpmOqk5mTzJGEnVf2/NnwzvTeuiezqY9QeIpigHvCKuj4JMyxLYk7zz6/5qY85v1yIUlMQ7xO3OWQFboNYr8E1O6w3wNGp3kGzbI8YrXankz3jfR2tFQBk7f4uWFzjYeaFv04QP830I0l/OSNrM4xBQ8JAQ20PxG2xznfY45g/gDTA2KxKEHLcpxZvq1aLTiqXOay0a270QMVIRIbK69Pov4y94TAZnDqf0DJpDo+dauH/TfDbtA/xelProl7CncE8ZG+HKrkYaNQef8YTql+9jLZwY9IMViwGrKJky6B5lzhQc= khaled@khaled-Inspiron-3576",
      DOMAIN_NAME = data.grid_gateway_domain.domain.fqdn,
      ADMIN_USERNAME = "khaled",
      ADMIN_PASSWORD = "password",
      ADMIN_EMAIL = "samehabouelsaad@gmail.com",
      # configure smtp settings bellow only If you have an working smtp service and you know what you’re doing.
      # otherwise leave these settings empty. gives wrong smtp settings will cause issues/server errors in taiga.
      DEFAULT_FROM_EMAIL = "",
      EMAIL_USE_TLS = "", # either "True" or "False"
      EMAIL_USE_SSL = "", # either "True" or "False"
      EMAIL_HOST = "",
      EMAIL_PORT = "",
      EMAIL_HOST_USER = "",
      EMAIL_HOST_PASSWORD = "",
    }
    planetary = true
    publicip = true
  }
}

data "grid_gateway_domain" "domain" {
  node = 2
  name = "grid3taiga"
}
resource "grid_name_proxy" "p1" {
  node            = 2
  name            = "grid3taiga"
  backends        = [format("http://%s:9000", grid_deployment.node1.vms[0].ygg_ip)]
  tls_passthrough = false
}

output "node1_zmachine1_ip" {
  value = grid_deployment.node1.vms[0].ip
}


output "node1_zmachine1_ygg_ip" {
  value = grid_deployment.node1.vms[0].ygg_ip
}

output "fqdn" {
  value = data.grid_gateway_domain.domain.fqdn
}
