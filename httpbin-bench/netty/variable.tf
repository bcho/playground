variable "security_group_name" {
  description = "name of the security group"
  type = string
  default = "playground-httpbin-netty"
}

variable "app_instance_name" {
  description = "name of the app instance"
  type = string
  default = "playground-httpbin-netty-app"
}

variable "app_instance_type" {
  description = "instance type of the app instance"
  type = string
  # TODO: ecs.sn1ne.3xlarge
  default = "ecs.t5-lc2m1.nano"
}

variable "app_image_id" {
  description = "instance image of the app instance"
  type = string
  default = "ubuntu_18_04_64_20G_alibase_20190624.vhd"
}

variable "app_hostname" {
  description = "hostname of the app instance"
  type = string
  default = "httpbin-app"
}

variable "tester_instance_name" {
  description = "name of the tester instance"
  type = string
  default = "playground-httpbin-netty-tester"
}

variable "tester_instance_type" {
  description = "instance type of the tester instance"
  type = string
  # TODO: ecs.sn1ne.2xlarge
  default = "ecs.t5-lc2m1.nano"
}

variable "tester_image_id" {
  description = "instance image of the tester instance"
  type = string
  default = "ubuntu_18_04_64_20G_alibase_20190624.vhd"
}

variable "tester_hostname" {
  description = "hostname of the tester instance"
  type = string
  default = "httpbin-tester"
}

variable "vpc_name" {
  description = "name of the vpc"
  type = string
  default = "playground-httpbin-netty"
}

variable "vpc_cidr_block" {
  description = "cidr block to use"
  type = string
  default = "10.0.0.0/24"
}

variable "region" {
  description = "region to use"
  type = string
  default = "cn-hongkong"
}

variable "available_zone" {
  description = "az to use"
  type = string
  default = "cn-hongkong-b"
}

variable "keypair_name" {
  description = "keypair name to use"
  type = string
  default = "httpbin-bench"
}

variable "ansible_user" {
  description = "user to run ansible provision"
  type = string
  default = "root"
}

variable "ansible_force_run" {
  description = "override to force rerun provision"
  type = string
  default = ""
}
