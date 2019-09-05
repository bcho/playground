provider "alicloud" {
  region = "${var.region}"
}

resource "alicloud_vpc" "vpc" {
  name = "${var.vpc_name}"
  cidr_block = "${var.vpc_cidr_block}"
}

resource "alicloud_vswitch" "vswitch" {
  vpc_id = "${alicloud_vpc.vpc.id}"
  cidr_block = "${var.vpc_cidr_block}"
  availability_zone = "${var.available_zone}"
}

resource "alicloud_security_group" "security_group" {
  name = "${var.security_group_name}"
  vpc_id = "${alicloud_vpc.vpc.id}"
}

resource "alicloud_security_group_rule" "security_group_ssh" {
  type = "ingress"
  ip_protocol = "tcp"
  policy = "accept"
  port_range = "22/22"
  priority = 1
  security_group_id = "${alicloud_security_group.security_group.id}"
  cidr_ip = "0.0.0.0/0"
}

resource "alicloud_instance" "app" {
  instance_name = "${var.app_instance_name}"
  instance_type = "${var.app_instance_type}"
  instance_charge_type = "PostPaid"
  security_groups = "${alicloud_security_group.security_group.*.id}"
  image_id = "${var.app_image_id}"
  vswitch_id = "${alicloud_vswitch.vswitch.id}"
  system_disk_size = 20
  host_name = "${var.app_hostname}"
  internet_charge_type = "PayByTraffic"
  internet_max_bandwidth_out = 5
}

resource "alicloud_instance" "tester" {
  instance_name = "${var.tester_instance_name}"
  instance_type = "${var.tester_instance_type}"
  instance_charge_type = "PostPaid"
  security_groups = "${alicloud_security_group.security_group.*.id}"
  image_id = "${var.tester_image_id}"
  vswitch_id = "${alicloud_vswitch.vswitch.id}"
  system_disk_size = 20
  host_name = "${var.tester_hostname}"
  internet_charge_type = "PayByTraffic"
  internet_max_bandwidth_out = 5
}

resource "alicloud_key_pair_attachment" "keypair_attachment" {
  key_name = "${var.keypair_name}"
  instance_ids = [
    "${alicloud_instance.app.id}",
    "${alicloud_instance.tester.id}"
  ]
}

resource "null_resource" "provision_app" {
  triggers = {
    app_private_ip = "${alicloud_instance.app.private_ip}",
    tester_private_ip = "${alicloud_instance.tester.private_ip}"
    force = "${var.ansible_force_run}"
  }

  provisioner "local-exec" {
    command = <<EOF
ansible-playbook \
    -u '${var.ansible_user}' \
    -i '${alicloud_instance.app.public_ip},' \
    -e 'ansible_ssh_extra_args="-o StrictHostKeyChecking=no"' \
    -e 'app_private_ip=${alicloud_instance.app.private_ip}' \
    -e 'tester_private_ip=${alicloud_instance.tester.private_ip}' \
    ansible/app.yaml
EOF
  }
}

resource "null_resource" "provision_tester" {
  triggers = {
    app_private_ip = "${alicloud_instance.app.private_ip}",
    tester_private_ip = "${alicloud_instance.tester.private_ip}"
    force = "${var.ansible_force_run}"
  }

  provisioner "local-exec" {
    command = <<EOF
ansible-playbook \
    -u '${var.ansible_user}' \
    -i '${alicloud_instance.tester.public_ip},' \
    -e 'ansible_ssh_extra_args="-o StrictHostKeyChecking=no"' \
    -e 'app_private_ip=${alicloud_instance.app.private_ip}' \
    -e 'tester_private_ip=${alicloud_instance.tester.private_ip}' \
    ansible/tester.yaml
EOF
  }
}
