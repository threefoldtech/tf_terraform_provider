terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
    }
  }
}
resource "random_bytes" "mycelium_ip_seed" {
  length = 6
}

resource "random_bytes" "mycelium_key" {
  length = 32
}

locals {
  solution_type = "Mattermost"
  name          = "ashrafmattermost"
}

provider "grid" {
}

resource "grid_scheduler" "sched" {
  requests {
    name      = "node"
    cru       = 2
    sru       = 1 * 1024
    mru       = 4 * 1024
    yggdrasil = false
    wireguard = true
  }

  requests {
    name             = "gateway"
    public_config    = true
    public_ips_count = 1
    yggdrasil        = false
    wireguard        = false
  }
}

resource "grid_network" "net1" {
  solution_type = local.solution_type
  name          = local.name
  nodes         = [grid_scheduler.sched.nodes["node"]]
  ip_range      = "10.1.0.0/16"
  description   = "newer network"
  add_wg_access = true
  mycelium_keys = {
    format("%s", grid_scheduler.sched.nodes["node"]) = random_bytes.mycelium_key.hex
  }
}

resource "grid_deployment" "d1" {
  solution_type = local.solution_type
  name          = local.name
  node          = grid_scheduler.sched.nodes["node"]
  network_name  = grid_network.net1.name
  vms {
    name       = "vm1"
    flist      = "https://hub.grid.tf/tf-official-apps/mattermost-latest.flist"
    cpu        = 2
    entrypoint = "/sbin/zinit init"
    memory     = 4096
    env_vars = {
      SSH_KEY      = file("~/.ssh/id_rsa.pub")
      DB_PASSWORD  = "ashroof"
      SITE_URL     = format("https://%s", data.grid_gateway_domain.domain.fqdn)
      SMTPPASSWORD = "password"
      SMTPUSERNAME = "Ashraf"
      SMTPSERVER   = "smtp.gmail.com"
      SMTPPORT     = 587
    }
    mycelium_ip_seed = random_bytes.mycelium_ip_seed.hex
  }
}

# this data source is used to break circular dependency in cases similar to the following:
# vm: needs to know the domain in its init script
# gateway_name: needs the ip of the vm to use as backend.
# - the fqdn can be computed from grid_gateway_domain for the vm
# - the backend can reference the vm ip directly 
data "grid_gateway_domain" "domain" {
  node = grid_scheduler.sched.nodes["gateway"]
  name = local.name
}

resource "grid_name_proxy" "p1" {
  solution_type   = local.solution_type
  name            = local.name
  node            = grid_scheduler.sched.nodes["gateway"]
  backends        = [format("http://[%s]:8000", grid_deployment.d1.vms[0].mycelium_ip)]
  tls_passthrough = false
}

output "fqdn" {
  value = data.grid_gateway_domain.domain.fqdn
}

output "node1_zmachine1_ip" {
  value = grid_deployment.d1.vms[0].ip
}

output "mycelium_ip" {
  value = grid_deployment.d1.vms[0].mycelium_ip
}
