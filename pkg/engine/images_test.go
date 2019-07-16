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
