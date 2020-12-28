package api

import (
	"context"
	"errors"

	ackcompare "github.com/aws/aws-controllers-k8s/pkg/compare"
	"github.com/aws/aws-controllers-k8s/services/apigatewayv2/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// customCreateApi checks if the Api resource should be imported or created.
// If the API resource should be imported, this operation performs the import and returns the updated ko.
// And if the API resource should be created, this operation returns nil, nil and createApi is performed in sdkCreate
// operation.
func (rm *resourceManager) customCreateApi(
	ctx context.Context,
	r *resource,
) (*resource, error) {
	// Based on the fields in desired, find whether we need to reimport or update
	if rm.importFieldsPresent(r.ko) {
		if err := rm.validateImportApiInputFields(r.ko); err != nil {
			return nil, err
		} else {
			// import
			return rm.importApi(ctx, r)
		}
	} else {
		if err := rm.validateCreateApiInputFields(r.ko); err != nil {
			return nil, err
		} else {
			return nil, nil
		}
	}
}

// customUpdateApi is the custom implementation for API resource's update operation
// If the API resource should be reimported, this operation performs the reimportApi and returns the updated ko.
// And if the API resource should be updated, this operation performs the updateApi and returns the updated ko.
func (rm *resourceManager) customUpdateApi(ctx context.Context,
	desired *resource,
	latest *resource,
	diffReporter *ackcompare.Reporter,
) (*resource, error) {
	// Based on the fields in desired, find whether we need to reimport or update
	if rm.importFieldsPresent(desired.ko) {
		if err := rm.validateReimportApiInputFields(desired.ko); err != nil {
			return nil, err
		} else {
			return rm.reimportApi(ctx, desired)
		}
	} else {
		if err := rm.validateUpdateApiInputFields(desired.ko); err != nil {
			return nil, err
		} else {
			return rm.updateApi(ctx, desired)
		}
	}
}

// importFieldsPresent checks for the presence of 'Body', 'Basepath' & 'FailOnWarning' fields
// in the API resource. When the mentioned fields are present, ImportApi operation is desired over CreateApi
func (rm *resourceManager) importFieldsPresent(api *v1alpha1.API) bool {
	if api.Spec.Body != nil || api.Spec.Basepath != nil || api.Spec.FailOnWarnings != nil {
		return true
	}
	return false
}

// validateImportApiInputFields validates if all the fields are present for a successful 'ImportApi' call
func (rm *resourceManager) validateImportApiInputFields(api *v1alpha1.API) error {
	// For import-api, body is a required field
	if api.Spec.Body == nil {
		errorMessage := ""
		if api.Spec.FailOnWarnings != nil {
			errorMessage += "'FailOnWarnings'"
		}

		if api.Spec.Basepath != nil {
			if errorMessage == "" {
				errorMessage += "'Basepath'"
			} else {
				errorMessage += " and 'Basepath'"
			}
		}
		errorMessage += " field(s) can only be used with 'Body' field for import-api operation"

		return errors.New(errorMessage)
	} else {
		// Body field is present.
		// Check that no other fields except 'Basepath' and 'FailOnWarnings' is present.
		specCopy := api.Spec.DeepCopy()
		specCopy.Body = nil
		specCopy.FailOnWarnings = nil
		specCopy.Basepath = nil
		opts := []cmp.Option{cmpopts.EquateEmpty()}
		if cmp.Equal(*specCopy, v1alpha1.APISpec{}, opts...) {
			return nil
		} else {
			return errors.New("only 'FailOnWarnings' and 'Basepath' fields can be used with 'Body' field")
		}
	}
}

// validateReimportApiInputFields validates if all the fields are present for a successful ReimportApi operation
// Currently this validation is similar to ImportApi validation.
func (rm *resourceManager) validateReimportApiInputFields(api *v1alpha1.API) error {
	return rm.validateImportApiInputFields(api)
}

// validateCreateApiInputFields validates if all the fields are present for a successful CreateApi operation
func (rm *resourceManager) validateCreateApiInputFields(api *v1alpha1.API) error {
	if api.Spec.Name == nil || api.Spec.ProtocolType == nil {
		return errors.New("'Name' and 'ProtocolType' are required properties if 'Body' field is not present")
	}
	return nil
}

// validateUpdateApiInputFields validates if all the fields are present for a successful UpdateApi operation
// Currently this validation is similar to CreateApi validation.
func (rm *resourceManager) validateUpdateApiInputFields(api *v1alpha1.API) error {
	return rm.validateCreateApiInputFields(api)
}

// importApi creates the Api resource by performing ImportApi sdk operation
func (rm *resourceManager) importApi(ctx context.Context, desired *resource) (*resource, error) {
	input, err := rm.importApiInput(desired)
	if err != nil {
		return nil, err
	}

	resp, respErr := rm.sdkapi.ImportApiWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "ImportApi", respErr)
	if respErr != nil {
		return nil, respErr
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.ApiEndpoint != nil {
		ko.Status.APIEndpoint = resp.ApiEndpoint
	}
	if resp.ApiGatewayManaged != nil {
		ko.Status.APIGatewayManaged = resp.ApiGatewayManaged
	}
	if resp.ApiId != nil {
		ko.Status.APIID = resp.ApiId
	}
	if resp.CreatedDate != nil {
		ko.Status.CreatedDate = &metav1.Time{*resp.CreatedDate}
	}
	if resp.ImportInfo != nil {
		f9 := []*string{}
		for _, f9iter := range resp.ImportInfo {
			var f9elem string
			f9elem = *f9iter
			f9 = append(f9, &f9elem)
		}
		ko.Status.ImportInfo = f9
	}
	if resp.Warnings != nil {
		f15 := []*string{}
		for _, f15iter := range resp.Warnings {
			var f15elem string
			f15elem = *f15iter
			f15 = append(f15, &f15elem)
		}
		ko.Status.Warnings = f15
	}

	rm.setStatusDefaults(ko)

	return &resource{ko}, nil
}

// reimportApi updates the Api resource's desired state after performing ReimportApi sdk operation
func (rm *resourceManager) reimportApi(ctx context.Context, desired *resource) (*resource, error) {
	input, err := rm.reimportApiInput(desired)
	if err != nil {
		return nil, err
	}

	resp, respErr := rm.sdkapi.ReimportApiWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "ReimportApi", respErr)
	if respErr != nil {
		return nil, respErr
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.ApiEndpoint != nil {
		ko.Status.APIEndpoint = resp.ApiEndpoint
	}
	if resp.ApiGatewayManaged != nil {
		ko.Status.APIGatewayManaged = resp.ApiGatewayManaged
	}
	if resp.ApiId != nil {
		ko.Status.APIID = resp.ApiId
	}
	if resp.CreatedDate != nil {
		ko.Status.CreatedDate = &metav1.Time{*resp.CreatedDate}
	}
	if resp.ImportInfo != nil {
		f9 := []*string{}
		for _, f9iter := range resp.ImportInfo {
			var f9elem string
			f9elem = *f9iter
			f9 = append(f9, &f9elem)
		}
		ko.Status.ImportInfo = f9
	}
	if resp.Warnings != nil {
		f15 := []*string{}
		for _, f15iter := range resp.Warnings {
			var f15elem string
			f15elem = *f15iter
			f15 = append(f15, &f15elem)
		}
		ko.Status.Warnings = f15
	}

	rm.setStatusDefaults(ko)

	return &resource{ko}, nil
}

// updateApi updates the Api resource's desired state after performing UpdateApi sdk operation
func (rm *resourceManager) updateApi(ctx context.Context, desired *resource) (*resource, error) {
	input, err := rm.updateApiInput(desired)
	if err != nil {
		return nil, err
	}

	resp, respErr := rm.sdkapi.UpdateApiWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateApi", respErr)
	if respErr != nil {
		return nil, respErr
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.ApiEndpoint != nil {
		ko.Status.APIEndpoint = resp.ApiEndpoint
	}
	if resp.ApiGatewayManaged != nil {
		ko.Status.APIGatewayManaged = resp.ApiGatewayManaged
	}
	if resp.ApiId != nil {
		ko.Status.APIID = resp.ApiId
	}
	if resp.CreatedDate != nil {
		ko.Status.CreatedDate = &metav1.Time{*resp.CreatedDate}
	}
	if resp.ImportInfo != nil {
		f9 := []*string{}
		for _, f9iter := range resp.ImportInfo {
			var f9elem string
			f9elem = *f9iter
			f9 = append(f9, &f9elem)
		}
		ko.Status.ImportInfo = f9
	}
	if resp.Warnings != nil {
		f15 := []*string{}
		for _, f15iter := range resp.Warnings {
			var f15elem string
			f15elem = *f15iter
			f15 = append(f15, &f15elem)
		}
		ko.Status.Warnings = f15
	}

	rm.setStatusDefaults(ko)

	return &resource{ko}, nil
}

// importApiInput returns an SDK-specific struct for the HTTP request
// payload of the  ImportApi call for the resource
func (rm *resourceManager) importApiInput(r *resource) (*apigatewayv2.ImportApiInput, error) {
	res := &apigatewayv2.ImportApiInput{}
	if r.ko.Spec.Body != nil {
		res.SetBody(*r.ko.Spec.Body)
	}
	if r.ko.Spec.Basepath != nil {
		res.SetBasepath(*r.ko.Spec.Basepath)
	}
	if r.ko.Spec.FailOnWarnings != nil {
		res.SetFailOnWarnings(*r.ko.Spec.FailOnWarnings)
	}
	return res, nil
}

// reimportApiInput returns an SDK-specific struct for the HTTP request
// payload of the  ReimportApi call for the resource
func (rm *resourceManager) reimportApiInput(r *resource) (*apigatewayv2.ReimportApiInput, error) {
	res := &apigatewayv2.ReimportApiInput{}

	if r.ko.Status.APIID != nil {
		res.SetApiId(*r.ko.Status.APIID)
	} else {
		return nil, errors.New("'APIID' is required input parameter for 'ReimportApi' operation")
	}

	if r.ko.Spec.Body != nil {
		res.SetBody(*r.ko.Spec.Body)
	}
	if r.ko.Spec.Basepath != nil {
		res.SetBasepath(*r.ko.Spec.Basepath)
	}
	if r.ko.Spec.FailOnWarnings != nil {
		res.SetFailOnWarnings(*r.ko.Spec.FailOnWarnings)
	}
	return res, nil
}

// updateApiInput returns an SDK-specific struct for the HTTP request
// payload of the UpdateApi call for the resource
func (rm *resourceManager) updateApiInput(
	r *resource,
) (*apigatewayv2.UpdateApiInput, error) {
	res := &apigatewayv2.UpdateApiInput{}

	if r.ko.Status.APIID != nil {
		res.SetApiId(*r.ko.Status.APIID)
	}
	if r.ko.Spec.APIKeySelectionExpression != nil {
		res.SetApiKeySelectionExpression(*r.ko.Spec.APIKeySelectionExpression)
	}
	if r.ko.Spec.CorsConfiguration != nil {
		f2 := &apigatewayv2.Cors{}
		if r.ko.Spec.CorsConfiguration.AllowCredentials != nil {
			f2.SetAllowCredentials(*r.ko.Spec.CorsConfiguration.AllowCredentials)
		}
		if r.ko.Spec.CorsConfiguration.AllowHeaders != nil {
			f2f1 := []*string{}
			for _, f2f1iter := range r.ko.Spec.CorsConfiguration.AllowHeaders {
				var f2f1elem string
				f2f1elem = *f2f1iter
				f2f1 = append(f2f1, &f2f1elem)
			}
			f2.SetAllowHeaders(f2f1)
		}
		if r.ko.Spec.CorsConfiguration.AllowMethods != nil {
			f2f2 := []*string{}
			for _, f2f2iter := range r.ko.Spec.CorsConfiguration.AllowMethods {
				var f2f2elem string
				f2f2elem = *f2f2iter
				f2f2 = append(f2f2, &f2f2elem)
			}
			f2.SetAllowMethods(f2f2)
		}
		if r.ko.Spec.CorsConfiguration.AllowOrigins != nil {
			f2f3 := []*string{}
			for _, f2f3iter := range r.ko.Spec.CorsConfiguration.AllowOrigins {
				var f2f3elem string
				f2f3elem = *f2f3iter
				f2f3 = append(f2f3, &f2f3elem)
			}
			f2.SetAllowOrigins(f2f3)
		}
		if r.ko.Spec.CorsConfiguration.ExposeHeaders != nil {
			f2f4 := []*string{}
			for _, f2f4iter := range r.ko.Spec.CorsConfiguration.ExposeHeaders {
				var f2f4elem string
				f2f4elem = *f2f4iter
				f2f4 = append(f2f4, &f2f4elem)
			}
			f2.SetExposeHeaders(f2f4)
		}
		if r.ko.Spec.CorsConfiguration.MaxAge != nil {
			f2.SetMaxAge(*r.ko.Spec.CorsConfiguration.MaxAge)
		}
		res.SetCorsConfiguration(f2)
	}
	if r.ko.Spec.CredentialsARN != nil {
		res.SetCredentialsArn(*r.ko.Spec.CredentialsARN)
	}
	if r.ko.Spec.Description != nil {
		res.SetDescription(*r.ko.Spec.Description)
	}
	if r.ko.Spec.DisableExecuteAPIEndpoint != nil {
		res.SetDisableExecuteApiEndpoint(*r.ko.Spec.DisableExecuteAPIEndpoint)
	}
	if r.ko.Spec.DisableSchemaValidation != nil {
		res.SetDisableSchemaValidation(*r.ko.Spec.DisableSchemaValidation)
	}
	if r.ko.Spec.Name != nil {
		res.SetName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.RouteKey != nil {
		res.SetRouteKey(*r.ko.Spec.RouteKey)
	}
	if r.ko.Spec.RouteSelectionExpression != nil {
		res.SetRouteSelectionExpression(*r.ko.Spec.RouteSelectionExpression)
	}
	if r.ko.Spec.Target != nil {
		res.SetTarget(*r.ko.Spec.Target)
	}
	if r.ko.Spec.Version != nil {
		res.SetVersion(*r.ko.Spec.Version)
	}

	return res, nil
}
