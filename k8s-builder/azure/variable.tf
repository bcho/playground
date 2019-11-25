variable "subscription_id" {
  type = "string"
}

variable "prefix" {
  description = "resource prefix"
  type = "string"
  default = "k8s_builder"
}

variable "location" {
  description = "region location"
  type = "string"
  default = "East Asia"
}

variable "network_address_space" {
  description = "network address space"
  type = "string"
  default = "10.0.0.0/16"
}

variable "network_address_prefix" {
  description = "network address prefix"
  type = "string"
  default = "10.0.2.0/24"
}

variable "builder_vm_size" {
  description = "vm size for the builder"
  type = "string"
  default = "Standard_F4s"
}

variable "builder_hostname" {
  type = "string"
  default = "k8s-builder"
}

variable "builder_adminusername" {
  type = "string"
  default = "k8s"
}

variable "builder_adminpassword" {
  type = "string"
}

variable "ansible_force_run" {
  description = "override to force rerun provision"
  type = string
  default = ""
}
