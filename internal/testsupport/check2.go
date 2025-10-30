package testsupport

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// GetAPIResourceFunc fetches a resource 'T' from the Hetzner Cloud API using its
// terraform attributes 'attrs'.
type GetAPIResourceFunc[T any] func(c *hcloud.Client, attrs map[string]string) (*T, error)

// CopyAPIResource is used to wrap [GetAPIResourceFunc], and copy the fetched api resource 'T'
// into the provide 'target' pointer variable before passing down the result.
func CopyAPIResource[T any](target *T, getter GetAPIResourceFunc[T]) GetAPIResourceFunc[T] {
	return func(c *hcloud.Client, attrs map[string]string) (*T, error) {
		result, err := getter(c, attrs)
		if result != nil {
			*target = *result
		}
		return result, err
	}
}

// CheckAPIResourcePresent checks that the terraform resource 'tfID' is present in the API.
func CheckAPIResourcePresent[T any](tfID string, getter GetAPIResourceFunc[T]) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfID]
		if !ok {
			return fmt.Errorf("resource not in state: %s", tfID)
		}

		client, err := CreateClient()
		if err != nil {
			return fmt.Errorf("could not create api client: %w", err)
		}

		result, err := getter(client, res.Primary.Attributes)
		if err != nil {
			return fmt.Errorf("could not get resource from api: %w", err)
		}

		if result == nil {
			return fmt.Errorf("resource is not present in api: %s", tfID)
		}
		return nil
	}
}

// CheckAPIResourceAllAbsent checks that all the terraform resource type 'resType' are absent from the API.
func CheckAPIResourceAllAbsent[T any](resType string, getter GetAPIResourceFunc[T]) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for tfID, res := range s.RootModule().Resources {
			if res.Type != resType {
				continue
			}

			client, err := CreateClient()
			if err != nil {
				return fmt.Errorf("could not create api client: %w", err)
			}

			result, err := getter(client, res.Primary.Attributes)
			if err != nil {
				return fmt.Errorf("could not get resource from api: %w", err)
			}

			if result != nil {
				return fmt.Errorf("resource is not absent from api: %s", tfID)
			}
		}
		return nil
	}
}

func StringExactFromFunc(fn func() string) knownvalue.Check {
	return knownvalue.StringFunc(func(other string) error {
		value := fn()

		if other != value {
			return fmt.Errorf("expected value %s for StringExact check, got: %s", value, other)
		}

		return nil
	})
}
