  terraform {
    required_providers {
      grid = {
        source = "threefoldtech/grid"
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
    name = "khaledpeertube"
  }
  resource "grid_network" "net1" {
      nodes = [2]
      ip_range = "10.1.0.0/16"
      name = "network"
      description = "newer network"
      add_wg_access = true
  }
  resource "grid_deployment" "d1" {
    node = 2
    network_name = grid_network.net1.name
    ip_range = lookup(grid_network.net1.nodes_ip_range, 2, "")
    vms {
      name = "vm1"
      flist = "https://hub.grid.tf/tf-official-apps/peertube-v3.1.1.flist"
      cpu = 2 
      # publicip = true
      entrypoint = "/sbin/zinit init"
      memory = 4096
      env_vars = {
        SSH_KEY = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCnJiB+sXPfqMKx6g67qOONqD0Kme08OKPtvacRwP6o0T8b404YaBvJaAEnsTgNE5A8vl15LBHdI94MdNoNwV7xT9HWWlw6hQ8PzN0e6z5M5bsKH6R6cKVbg8iYUESWWkRBr8iN10KDmbysyIbpr1QVIUoAbET4XpkJMxw46L8ClBsfPY5YFd1Bdd1oLwHJD4+cZQdZf9iSV4EtVfOfgpmOqk5mTzJGEnVf2/NnwzvTeuiezqY9QeIpigHvCKuj4JMyxLYk7zz6/5qY85v1yIUlMQ7xO3OWQFboNYr8E1O6w3wNGp3kGzbI8YrXankz3jfR2tFQBk7f4uWFzjYeaFv04QP830I0l/OSNrM4xBQ8JAQ20PxG2xznfY45g/gDTA2KxKEHLcpxZvq1aLTiqXOay0a270QMVIRIbK69Pov4y94TAZnDqf0DJpDo+dauH/TfDbtA/xelProl7CncE8ZG+HKrkYaNQef8YTql+9jLZwY9IMViwGrKJky6B5lzhQc= khaled@khaled-Inspiron-3576"
        PEERTUBE_DB_SUFFIX = "_prod"
        PEERTUBE_DB_USERNAME = "peertube"
        PEERTUBE_DB_PASSWORD = "peertube"
        PEERTUBE_ADMIN_EMAIL = "support@threefold.com"
        PEERTUBE_WEBSERVER_HOSTNAME = data.grid_gateway_domain.domain.fqdn
        PEERTUBE_WEBSERVER_PORT = 443
        PEERTUBE_SMTP_HOSTNAME = "https://app.sendgrid.com"
        PEERTUBE_SMTP_USERNAME = "sendgridusername"
        PEERTUBE_SMTP_PASSWORD = "sendgridpassword"
        PEERTUBE_BIND_ADDRESS = "::",
      }
      planetary = true
    }
  }
  resource "grid_name_proxy" "p1" {
    node = 2
    name = "khaledpeertube"
    backends = [format("http://[%s]:9000", grid_deployment.d1.vms[0].ygg_ip)]
    tls_passthrough = false
  }
  output "fqdn" {
      value = data.grid_gateway_domain.domain.fqdn
  }
  output "node1_zmachine1_ip" {
      value = grid_deployment.d1.vms[0].ip
  }
  output "public_ip" {
      value = split("/",grid_deployment.d1.vms[0].computedip)[0]
  }

  output "ygg_ip" {
      value = grid_deployment.d1.vms[0].ygg_ip
  }

