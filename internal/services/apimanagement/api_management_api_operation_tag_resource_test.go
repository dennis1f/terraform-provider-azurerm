package apimanagement_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ApiManagementApiOperationTagResource struct{}

func TestAccApiManagementApiOperationTag_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_api_operation_tag", "test")
	r := ApiManagementApiOperationTagResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccApiManagementApiOperationTag_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_api_management_api_operation_tag", "test")
	r := ApiManagementApiOperationTagResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (ApiManagementApiOperationTagResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.ApiOperationTagID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ApiManagement.TagClient.GetByOperation(ctx, id.ResourceGroup, id.ServiceName, id.ApiName, id.OperationName, id.TagName)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (r ApiManagementApiOperationTagResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurerm_api_management_tag" "test" {
  api_management_id = azurerm_api_management.test.id
  name              = "acctest-Tag-%d"
}

resource "azurerm_api_management_api_operation_tag" "test" {
  api_operation_id = azurerm_api_management_api_operation.test.id
  name             = "acctest-Tag-%d"
}
`, ApiManagementApiOperationResource{}.basic(data), data.RandomInteger, data.RandomInteger)
}

func (r ApiManagementApiOperationTagResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
resource "azurerm_api_management_api_operation_tag" "import" {
  api_operation_id = azurerm_api_management_api_operation.test.id
  name             = azurerm_api_management_api_operation_tag.test.name
}
`, r.basic(data))
}
