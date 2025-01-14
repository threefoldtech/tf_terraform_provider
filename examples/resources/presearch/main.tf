terraform {
  required_providers {
    grid = {
      source = "threefoldtech/grid"
    }
  }
}

provider "grid" {
}

resource "random_bytes" "mycelium_ip_seed" {
  length = 6
}

resource "random_bytes" "mycelium_key" {
  length = 32
}

locals {
  solution_type = "Presearch"
  name          = "presearch"
}


resource "grid_scheduler" "sched" {
  requests {
    name             = "presearch"
    cru              = 1
    sru              = 1024 * 10
    mru              = 1024
    public_ips_count = 1
    public_config    = true
    yggdrasil        = false
    wireguard        = true
  }
}

resource "grid_network" "net1" {
  solution_type = local.solution_type
  name          = local.name
  nodes         = [grid_scheduler.sched.nodes["presearch"]]
  ip_range      = "10.1.0.0/16"
  description   = "presearch network"
  add_wg_access = true
  mycelium_keys = {
    format("%s", grid_scheduler.sched.nodes["presearch"]) = random_bytes.mycelium_key.hex
  }
}

# Deployment specs
resource "grid_deployment" "d1" {
  solution_type = local.solution_type
  name          = local.name
  node          = grid_scheduler.sched.nodes["presearch"]
  network_name  = grid_network.net1.name

  disks {
    name        = "data"
    size        = 10
    description = "volume holding docker data"
  }

  vms {
    name             = local.name
    flist            = "https://hub.grid.tf/tf-official-apps/presearch-v2.2.flist"
    entrypoint       = "/sbin/zinit init"
    publicip         = true
    cpu              = 1
    memory           = 1024
    mycelium_ip_seed = random_bytes.mycelium_ip_seed.hex

    mounts {
      name        = "data"
      mount_point = "/var/lib/docker"
    }

    env_vars = {
      SSH_KEY                     = file("~/.ssh/id_rsa.pub"),
      PRESEARCH_REGISTRATION_CODE = "",

      # COMMENT the two env vars below to create a new node. 
      # or uncomment and fill them from your old node to restore it. #

      # optional keys pair from the old node 
      # important to follow the schema ` <<-EOF ... EOF ` with no indentation
      PRESEARCH_BACKUP_PRI_KEY = <<EOF
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDQjfuZ3uIGOXUP
Qqpw1K85LV6sZWOAntUnhL73GXTWcwBer06yPI1ush8Vj6tdP94hmUFfWW85vYRU
...
-----END PRIVATE KEY-----
      EOF
      PRESEARCH_BACKUP_PUB_KEY = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0I37md7iBjl1D0KqcNSv
OS1erGVjgJ7VJ4S+9xl01nMAXq9OsjyNbrIfFY+rXT/eIZlBX1lvOb2EVJ93o1mz
...
-----END PUBLIC KEY-----
      EOF
    }
  }
}


# Print deployment info
output "node1_zmachine1_ip" {
  value = grid_deployment.d1.vms[0].ip
}

output "computed_public_ip" {
  value = split("/", grid_deployment.d1.vms[0].computedip)[0]
}

output "mycelium_ip" {
  value = grid_deployment.d1.vms[0].mycelium_ip
}
