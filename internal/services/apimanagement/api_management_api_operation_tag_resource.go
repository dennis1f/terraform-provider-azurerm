package apimanagement

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/apimanagement/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceApiManagementApiOperationTag() *pluginsdk.Resource {
	resource := &pluginsdk.Resource{
		Create: resourceApiManagementApiOperationTagCreate,
		Read:   resourceApiManagementApiOperationTagRead,
		Delete: resourceApiManagementApiOperationTagDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.OperationTagID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"api_operation_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApiOperationID,
			},

			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApiManagementChildName,
			},
		},
	}

	if !features.FourPointOhBeta() {
		resource.Schema["display_name"] = &pluginsdk.Schema{
			Type:       pluginsdk.TypeString,
			Optional:   true,
			ForceNew:   true, // Required, because we don't have an update.
			Deprecated: "This property has been deprecated and will be removed in v4.0 of the provider",
		}
	}

	return resource
}

func resourceApiManagementApiOperationTagCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ApiManagement.TagClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	apiOperationId, err := parse.ApiOperationID(d.Get("api_operation_id").(string))
	if err != nil {
		return err
	}

	tagName := d.Get("name").(string)
	tagId := parse.NewTagID(subscriptionId, apiOperationId.ResourceGroup, apiOperationId.ServiceName, tagName)
	if err != nil {
		return err
	}

	id := parse.NewApiOperationTagID(subscriptionId, apiOperationId.ResourceGroup, apiOperationId.ServiceName, apiOperationId.ApiName, apiOperationId.OperationName, tagId.Name)

	tagExists, err := client.Get(ctx, apiOperationId.ResourceGroup, apiOperationId.ServiceName, tagId.ID())
	if err != nil {
		if !utils.ResponseWasNotFound(tagExists.Response) {
			return fmt.Errorf("checking for presence of Tag %q: %s", id, err)
		}
	}

	tagAssignmentExist, err := client.GetByOperation(ctx, apiOperationId.ResourceGroup, apiOperationId.ServiceName, apiOperationId.ApiName, apiOperationId.OperationName, tagId.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(tagAssignmentExist.Response) {
			return fmt.Errorf("checking for presence of Tag Assignment %q: %s", id, err)
		}
	}

	if !utils.ResponseWasNotFound(tagAssignmentExist.Response) {
		return tf.ImportAsExistsError("azurerm_api_management_api_operation_tag", id.ID())
	}

	if _, err := client.AssignToOperation(ctx, apiOperationId.ResourceGroup, apiOperationId.ServiceName, apiOperationId.ApiName, apiOperationId.OperationName, tagId.Name); err != nil {
		return fmt.Errorf("assigning to api operation %q: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceApiManagementApiOperationTagRead(d, meta)
}

func resourceApiManagementApiOperationTagRead(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ApiManagement.TagClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.OperationTagID(d.Id())
	if err != nil {
		return err
	}

	apiOperationId := parse.NewApiOperationID(subscriptionId, id.ResourceGroup, id.ServiceName, id.ApiName, id.OperationName)

	resp, err := client.GetByOperation(ctx, id.ResourceGroup, id.ServiceName, id.ApiName, id.OperationName, id.TagName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] %q was not found - removing from state!", id)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %q: %+v", id, err)
	}

	d.Set("api_operation_id", apiOperationId.ID())
	d.Set("name", id.TagName)

	return nil
}

func resourceApiManagementApiOperationTagDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.TagClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.OperationTagID(d.Id())
	if err != nil {
		return err
	}

	if _, err = client.DetachFromOperation(ctx, id.ResourceGroup, id.ServiceName, id.ApiName, id.OperationName, id.TagName); err != nil {
		return fmt.Errorf("detaching api operation tag %q: %+v", id, err)
	}

	return nil
}
