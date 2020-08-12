variable "client_id" { type = string }
variable "client_secret" { type = string }
variable "tenant_id" { type = string }
variable "subscription_id" { type = string }
variable "location" { type = string }
variable "vm_size" { type = string }
variable "resource_group_name" { type = string }
variable "storage_account_name" { type = string }
variable "container_runtime" { default = "docker" }

variable "os_version" {type = string }

variable "sku_map" {
    type = map(string)
    default = {
        "2019" = "2019-Datacenter-Core-smalldisk"
    }
}

variable "version_map" {
    type = map(string)
    default = {
        "2019" = "17763.1282.2006061952"
    }
}

source "azure-arm" "vm" {
    client_id = var.client_id
    client_secret = var.client_secret
    tenant_id = var.tenant_id
    subscription_id = var.subscription_id
    location = var.location
    vm_size = var.vm_size
    os_type = "Windows"
    image_publisher = "MicrosoftWindowsServer"
    image_offer = "WindowsServer"
    image_sku = "${lookup(var.sku_map, var.os_version, "")}"
    image_version = "${lookup(var.version_map, var.os_version, "")}"
    resource_group_name = var.resource_group_name
    capture_container_name = "aksengine-vhds-windows-ws2019"
#    capture_name_prefix = "aksengine-{{user `create_time`}}"
    capture_name_prefix = "aksengine-now" 
    storage_account = var.storage_account_name
    communicator = "winrm"
    winrm_use_ssl = true
    winrm_insecure = true
    winrm_timeout = "10m"
    winrm_username = "packer"
#    azure_tags {
#        os = "Windows"
#        now = "{{user `create_time`}}"
#        createdBy = "aks-engine-vhd-pipeline"
#    }
}

build {
    sources = [ "source.azure-arm.vm" ]

    provisioner "powershell" {
        inline = [
            "Write-Host 'Hello from Packer!'"
        ]
    }

    provisioner "powershell" {
        elevated_user = "packer"
        elevated_password = "${build.WinRMPassword}"
        environment_vars = [
            "ProvisioningPhase=1",
            "ContainerRuntime=docker"
        ]
        script = "vhd/packer/configure-windows-vhd.ps1"
    }

    provisioner "windows-restart" {
        restart_timeout = "10m"
    }

    provisioner "windows-restart" {
        restart_timeout = "10m"
    }

    provisioner "powershell" {
        elevated_user = "packer"
        elevated_password = "${build.WinRMPassword}"
        environment_vars = [
            "ProvisioningPhase=2",
            "ContainerRuntime=docker"
        ]
        script = "vhd/packer/configure-windows-vhd.ps1"
    }

    provisioner "windows-restart" {
        restart_timeout = "10m"
    }

    provisioner "file" {
        direction = "upload"
        source = "vhd/notice/notice_windows.txt"
        destination = "c:\\NOTICE.txt"
    }

    provisioner "powershell" {
        elevated_user = "packer"
        elevated_password = "${build.WinRMPassword}"
        environment_vars = [
                "BUILD_BRANCH=build_branch",
                "BUILD_COMMIT=build_commit",
                "BUILD_ID=build_id",
                "BUILD_NUMBER=build_number",
                "BUILD_REPO=build_repo"
        ]
        script = "vhd/packer/write-release-notes-windows.ps1"
    }

    provisioner "file" {
        direction = "download"
        source = "c:\\release-notes.txt"
        destination = "release-notes.txt"
    }


#    provisioner "powershell" {
#        inline = [
#            "& $env:SystemRoot\\System32\\Sysprep\\Sysprep.exe /oobe /generalize /mode:vm /quiet /quit",
#            "while($true) { $imageState = Get-ItemProperty HKLM:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Setup\\State | Select ImageState; if($imageState.ImageState -ne 'IMAGE_STATE_GENERALIZE_RESEAL_TO_OOBE') { Write-Output $imageState.ImageState; Start-Sleep -s 10  } else { break } }",
#        ]
#    }

}