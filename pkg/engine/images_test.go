// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package engine

import (
	"testing"

	"github.com/Azure/aks-engine/pkg/api"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-10-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
)

func TestCreateImageFromURL(t *testing.T) {
	cases := []struct {
		name            string
		p               api.AgentPoolProfile
		expectedName    string
		expectedOsType  compute.OperatingSystemTypes
		expectedBlobURI string
	}{
		{
			name: "Linux image from URL",
			p: api.AgentPoolProfile{
				Name: "foo",
				Image: &api.Image{
					ImageURL: "linuxImageURL",
				},
			},
			expectedName:    "fooosCustomImage",
			expectedOsType:  compute.Linux,
			expectedBlobURI: "[parameters('fooosImageSourceUrl')]",
		},
		{
			name: "Windows image from URL",
			p: api.AgentPoolProfile{
				Name: "bar",
				Image: &api.Image{
					ImageURL: "windowsImageURL",
				},
				OSType: "Windows",
			},
			expectedName:    "barosCustomImage",
			expectedOsType:  compute.Windows,
			expectedBlobURI: "[parameters('barosImageSourceUrl')]",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := createImageFromURL(&c.p)

			expected := ImageARM{
				ARMResource: ARMResource{
					APIVersion: "[variables('apiVersionCompute')]",
				},
				Image: compute.Image{
					Type: to.StringPtr("Microsoft.Compute/images"),
					Name: to.StringPtr(c.expectedName),
					ImageProperties: &compute.ImageProperties{
						StorageProfile: &compute.ImageStorageProfile{
							OsDisk: &compute.ImageOSDisk{
								OsType:             c.expectedOsType,
								OsState:            compute.Generalized,
								BlobURI:            to.StringPtr(c.expectedBlobURI),
								StorageAccountType: compute.StorageAccountTypesStandardLRS,
							},
						},
					},
				},
			}

			diff := cmp.Diff(actual, expected)

			if diff != "" {
				t.Errorf("Unexpected diff while comparing image ARM: %s", diff)
			}
		})
	}
}

func TestGetVmStorageProfileImageReference(t *testing.T) {
	cases := []struct {
		name                   string
		p                      api.AgentPoolProfile
		expectedImageReference compute.ImageReference
	}{
		{
			name: "ImageURL",
			p: api.AgentPoolProfile{
				Name: "agent1",
				Image: &api.Image{
					ImageURL: "https://image.vhd",
				},
			},
			expectedImageReference: compute.ImageReference{
				ID: to.StringPtr("[resourceId('Microsoft.Compute/images', 'agent1osCustomImage')]"),
			},
		},
		{
			name: "ImageReference",
			p: api.AgentPoolProfile{
				Name: "agent1",
				Image: &api.Image{
					ImageRef: &api.ImageReference{
						Name:          "image_name",
						ResourceGroup: "resource_group",
					},
				},
			},
			expectedImageReference: compute.ImageReference{
				ID: to.StringPtr("[resourceId('resource_group', 'Microsoft.Compute/images', 'image_name')]"),
			},
		},
		{
			name: "GalleryImage",
			p: api.AgentPoolProfile{
				Name: "agent1",
				Image: &api.Image{
					ImageRef: &api.ImageReference{
						Gallery:        "gallery_name",
						Name:           "image_name",
						ResourceGroup:  "resource_group",
						SubscriptionID: "00000000-0000-0000-0000-000000000000",
						Version:        "0.1.0",
					},
				},
			},
			expectedImageReference: compute.ImageReference{
				ID: to.StringPtr("[concat('/subscriptions/', '00000000-0000-0000-0000-000000000000', '/resourceGroups/', 'resource_group', '/providers/Microsoft.Compute/galleries/', 'gallery_name', '/images/', 'image_name', '/versions/', '0.1.0')]"),
			},
		},
		{
			name: "MarketplaceImage",
			p: api.AgentPoolProfile{
				Name: "agent1",
				Image: &api.Image{
					MarketplaceImage: &api.MarketplaceImage{
						Offer:     "Offer",
						Publisher: "Publisher",
						Sku:       "Sku",
						Version:   "Version",
					},
				},
			},
			expectedImageReference: compute.ImageReference{
				Offer:     to.StringPtr("[variables('agent1osImageOffer')]"),
				Publisher: to.StringPtr("[variables('agent1osImagePublisher')]"),
				Sku:       to.StringPtr("[variables('agent1osImageSKU')]"),
				Version:   to.StringPtr("[variables('agent1osImageVersion')]"),
			},
		},
		{
			name: "DefaultImage",
			p: api.AgentPoolProfile{
				Name: "agent1",
			},
			expectedImageReference: compute.ImageReference{
				Offer:     to.StringPtr("[variables('agent1osImageOffer')]"),
				Publisher: to.StringPtr("[variables('agent1osImagePublisher')]"),
				Sku:       to.StringPtr("[variables('agent1osImageSKU')]"),
				Version:   to.StringPtr("[variables('agent1osImageVersion')]"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			//		t.Parallel()

			actual := getVmStorageProfileImageReference(&c.p)
			expected := c.expectedImageReference
			diff := cmp.Diff(actual, &expected)

			if diff != "" {
				t.Errorf("Unexpected diff while comparing compute.ImageReference ARM: %s", diff)
			}
		})
	}
}
