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

func getVmStorageProfileImageReference(profile *api.AgentPoolProfile) *compute.ImageReference {
	var computeImageRef compute.ImageReference

	image := profile.Image
	if !profile.HasImage() || (profile.HasImage() && image.HasMarketplaceImage()) {
		computeImageRef = compute.ImageReference{
			Offer:     to.StringPtr(fmt.Sprintf("[variables('%sosImageOffer')]", profile.Name)),
			Publisher: to.StringPtr(fmt.Sprintf("[variables('%sosImagePublisher')]", profile.Name)),
			Sku:       to.StringPtr(fmt.Sprintf("[variables('%sosImageSKU')]", profile.Name)),
			Version:   to.StringPtr(fmt.Sprintf("[variables('%sosImageVersion')]", profile.Name)),
		}
	} else if profile.HasImage() { 
		if image.HasGalleryImage() {
			v := fmt.Sprintf("[concat('/subscriptions/', '%s', '/resourceGroups/', '%s', '/providers/Microsoft.Compute/galleries/', '%s', '/images/', '%s', '/versions/', '%s')]", image.ImageRef.SubscriptionID, image.ImageRef.ResourceGroup, image.ImageRef.Gallery, image.ImageRef.Name, image.ImageRef.Version)
			computeImageRef = compute.ImageReference{
				ID: to.StringPtr(v),
			}
		} else if image.HasImageReference() {
			v := fmt.Sprintf("[resourceId('%s', 'Microsoft.Compute/images', '%s')]", image.ImageRef.ResourceGroup, image.ImageRef.Name)
			computeImageRef = compute.ImageReference{
				ID: to.StringPtr(v),
			}
		} else {
			v := fmt.Sprintf("[resourceId('Microsoft.Compute/images', '%sosCustomImage')]", profile.Name)
			computeImageRef = compute.ImageReference{
				ID: to.StringPtr(v),
			}
		}
	}

	return &computeImageRef
}
