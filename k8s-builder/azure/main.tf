provider "azurerm" {
  version = "=1.36.0"

  subscription_id = "${var.subscription_id}"
}

resource "azurerm_resource_group" "main" {
  name = "${var.prefix}-resource"
  location = "${var.location}"
}

resource "azurerm_virtual_network" "main" {
  name = "${var.prefix}-network"
  address_space = ["${var.network_address_space}"]
  resource_group_name = "${azurerm_resource_group.main.name}"
  location = "${azurerm_resource_group.main.location}"
}

resource "azurerm_subnet" "main" {
  name = "internal"
  resource_group_name = "${azurerm_resource_group.main.name}"
  virtual_network_name = "${azurerm_virtual_network.main.name}"
  address_prefix = "${var.network_address_prefix}"
}

resource "azurerm_public_ip" "builder" {
  name = "${var.prefix}-builder-public-ip"
  resource_group_name = "${azurerm_resource_group.main.name}"
  location = "${azurerm_resource_group.main.location}"
  allocation_method = "Dynamic"
}

resource "azurerm_network_security_group" "main" {
  name = "${var.prefix}-security-group"
  resource_group_name = "${azurerm_resource_group.main.name}"
  location = "${azurerm_resource_group.main.location}"

  security_rule {
    name = "SSH"
    priority = 1001
    direction = "Inbound"
    access = "Allow"
    protocol = "Tcp"
    source_port_range = "*"
    destination_port_range = "22"
    source_address_prefix = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_network_interface" "builder" {
  name = "${var.prefix}-builder-nic"
  resource_group_name = "${azurerm_resource_group.main.name}"
  location = "${azurerm_resource_group.main.location}"
  network_security_group_id = "${azurerm_network_security_group.main.id}"

  ip_configuration {
    name = "${var.prefix}-builder-nic-config"
    subnet_id = "${azurerm_subnet.main.id}"
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id = "${azurerm_public_ip.builder.id}"
  }
}

resource "azurerm_virtual_machine" "builder" {
  name = "${var.prefix}-vm-builder"
  resource_group_name = "${azurerm_resource_group.main.name}"
  location = "${azurerm_resource_group.main.location}"

  network_interface_ids = ["${azurerm_network_interface.builder.id}"]
  vm_size = "${var.builder_vm_size}"

  delete_os_disk_on_termination = true
  storage_os_disk {
    name = "${var.prefix}-vm-os-disk"
    caching = "ReadWrite"
    create_option = "FromImage"
    os_type = "Linux"
  }
  storage_image_reference {
    publisher = "Canonical"
    offer = "UbuntuServer"
    sku = "18.04-LTS"
    version = "latest"
  }

  os_profile {
    computer_name = "${var.builder_hostname}"
    admin_username = "${var.builder_adminusername}"
    admin_password = "${var.builder_adminpassword}"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}
