// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const configStorageFileTest = `
resource "bunnynet_storage_zone" "test" {
  name      = "test-acceptance-%s"
  zone_tier = "Standard"
  region    = "DE"
}

resource "bunnynet_storage_file" "test" {
  zone     = bunnynet_storage_zone.test.id
  path     = "/index.html"
  content = "<p>test-1</p>"
}
`

const configStorageFileIssue40Test = `
variable "name" {
  type = string
}

variable "filepath" {
  type = string
}

resource "bunnynet_storage_zone" "test" {
  name      = "test-acceptance-${var.name}"
  zone_tier = "Standard"
  region    = "DE"
}

resource "bunnynet_storage_file" "test" {
  zone     = bunnynet_storage_zone.test.id
  path     = "/index.html"
  source = "${var.filepath}"
}
`

func TestAccStorageFileResource(t *testing.T) {
	resourceName := "bunnynet_storage_file.test"
	testKey := generateRandomString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(configStorageFileTest, testKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "content", "<p>test-1</p>"),
					resource.TestCheckResourceAttr(resourceName, "checksum", "1EC877C32873420A7C2320916964BE83B8EF755BBC7F9BAC1D37B66E013402C2"),
					resource.TestCheckResourceAttr(resourceName, "size", "13"),
				),
			},
			// @TODO test import
		},
	})
}

func TestAccStorageFileIssue40Resource(t *testing.T) {
	resourceName := "bunnynet_storage_file.test"
	testKey := generateRandomString(12)

	dir, err := os.MkdirTemp("", testKey)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "/index.html")

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					err = testAccStorageFileWriteFile(path, "<p>test-1</p>")
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: configStorageFileIssue40Test,
				ConfigVariables: map[string]config.Variable{
					"name":     config.StringVariable(testKey),
					"filepath": config.StringVariable(path),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "checksum", "1EC877C32873420A7C2320916964BE83B8EF755BBC7F9BAC1D37B66E013402C2"),
					resource.TestCheckResourceAttr(resourceName, "size", "13"),
				),
			},
			{
				PreConfig: func() {
					err = testAccStorageFileWriteFile(path, "<p>test-2</p>")
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: configStorageFileIssue40Test,
				ConfigVariables: map[string]config.Variable{
					"name":     config.StringVariable(testKey),
					"filepath": config.StringVariable(path),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "checksum", "2DEEB3F366F09FDBAAAFE07EE905FA63CA436602D9AF4EAE9E03E988C1BDA9BB"),
					resource.TestCheckResourceAttr(resourceName, "size", "13"),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccStorageFileWriteFile(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(file, content)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
