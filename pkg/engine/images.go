// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package engine

import (
	"fmt"

	"github.com/Azure/aks-engine/pkg/api"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-10-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
)

// Creates an ARM resource for a vhd located at specified URL
func createImageFromURL(profile *api.AgentPoolProfile) ImageARM {
	osType := compute.Linux
	if profile.IsWindows() {
		osType = compute.Windows
	}

	return ImageARM{
		ARMResource: ARMResource{
			APIVersion: "[variables('apiVersionCompute')]",
		},
		Image: compute.Image{
			Type: to.StringPtr("Microsoft.Compute/images"),
			Name: to.StringPtr(fmt.Sprintf("%sosCustomImage", profile.Name)),
			Location: to.StringPtr("[parameters('location')]"),
			ImageProperties: &compute.ImageProperties{
				StorageProfile: &compute.ImageStorageProfile{
					OsDisk: &compute.ImageOSDisk{
						OsType:             osType,
						OsState:            compute.Generalized,
						BlobURI:            to.StringPtr(fmt.Sprintf("[parameters('%sosImageSourceUrl')]", profile.Name)),
						StorageAccountType: compute.StorageAccountTypesStandardLRS,
					},
				},
			},
		},
	}
}
