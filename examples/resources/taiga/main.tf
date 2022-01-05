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
  nodes         = [1]
  ip_range      = "10.1.0.0/16"
  name          = "network1"
  description   = "newer network"
  add_wg_access = true
}

resource "grid_deployment" "node1" {
  node         = 1
  network_name = grid_network.net2.name
  ip_range     = lookup(grid_network.net2.nodes_ip_range, 1, "")
  vms {
    name        = "taiga"
    flist       = "https://hub.grid.tf/samehabouelsaad.3bot/abouelsaad-taiga-test.flist"
    entrypoint  = "/sbin/zinit init"
    cpu         = 4
    memory      = 8096
    rootfs_size = 51200
    env_vars = {
      SSH_KEY     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC9MI7fh4xEOOEKL7PvLvXmSeRWesToj6E26bbDASvlZnyzlSKFLuYRpnVjkr8JcuWKZP6RQn8+2aRs6Owyx7Tx+9kmEh7WI5fol0JNDn1D0gjp4XtGnqnON7d0d5oFI+EjQQwgCZwvg0PnV/2DYoH4GJ6KPCclPz4a6eXrblCLA2CHTzghDgyj2x5B4vB3rtoI/GAYYNqxB7REngOG6hct8vdtSndeY1sxuRoBnophf7MPHklRQ6EG2GxQVzAOsBgGHWSJPsXQkxbs8am0C9uEDL+BJuSyFbc/fSRKptU1UmS18kdEjRgGNoQD7D+Maxh1EbmudYqKW92TVgdxXWTQv1b1+3dG5+9g+hIWkbKZCBcfMe4nA5H7qerLvoFWLl6dKhayt1xx5mv8XhXCpEC22/XHxhRBHBaWwSSI+QPOCvs4cdrn4sQU+EXsy7+T7FIXPeWiC2jhFd6j8WIHAv6/rRPsiwV1dobzZOrCxTOnrqPB+756t7ANxuktsVlAZaM= sameh@sameh-inspiron-3576",
      DOMAIN_NAME = data.grid_gateway_domain.domain.fqdn,
    }
    planetary = true
  }
}

data "grid_gateway_domain" "domain" {
  node = 7
  name = "grid3taiga"
}
resource "grid_name_proxy" "p1" {
  node            = 7
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
