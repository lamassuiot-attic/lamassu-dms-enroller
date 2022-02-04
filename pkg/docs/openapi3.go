package docs

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/lamassuiot/dms-enroller/pkg/config"
)

func NewOpenAPI3(config config.Config) openapi3.T {

	arrayOf := func(items *openapi3.SchemaRef) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "array", Items: items}}
	}

	openapiSpec := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       "Lamassu DMS Enroller API",
			Description: "REST API used for interacting with Lamassu DMS Enroller",
			Version:     "0.0.0",
			License: &openapi3.License{
				Name: "MPL v2.0",
				URL:  "https://github.com/lamassuiot/lamassu-compose/blob/main/LICENSE",
			},
			Contact: &openapi3.Contact{
				URL: "https://github.com/lamassuiot",
			},
		},
		Servers: openapi3.Servers{
			&openapi3.Server{
				Description: "Current Server",
				URL:         "/",
			},
		},
	}

	openapiSpec.Components.Schemas = openapi3.Schemas{
		"CSR": openapi3.NewSchemaRef("",
			openapi3.NewObjectSchema().
				WithProperty("id", openapi3.NewIntegerSchema()).
				WithProperty("dms_name", openapi3.NewStringSchema()).
				WithProperty("country", openapi3.NewStringSchema()).
				WithProperty("state", openapi3.NewStringSchema()).
				WithProperty("locality", openapi3.NewStringSchema()).
				WithProperty("organization", openapi3.NewStringSchema()).
				WithProperty("organization_unit", openapi3.NewStringSchema()).
				WithProperty("common_name", openapi3.NewStringSchema()).
				WithProperty("mail", openapi3.NewStringSchema()).
				WithProperty("status", openapi3.NewStringSchema()).
				WithProperty("csr", openapi3.NewStringSchema()).
				WithProperty("url", openapi3.NewStringSchema()),
		),
		"CSRForm": openapi3.NewSchemaRef("",
			openapi3.NewObjectSchema().
				WithProperty("dms_name", openapi3.NewStringSchema()).
				WithProperty("country", openapi3.NewStringSchema()).
				WithProperty("state", openapi3.NewStringSchema()).
				WithProperty("locality", openapi3.NewStringSchema()).
				WithProperty("organization", openapi3.NewStringSchema()).
				WithProperty("organization_unit", openapi3.NewStringSchema()).
				WithProperty("common_name", openapi3.NewStringSchema()).
				WithProperty("mail", openapi3.NewStringSchema()).
				WithProperty("key_type", openapi3.NewStringSchema()).
				WithProperty("key_bits", openapi3.NewIntegerSchema()).
				WithProperty("url", openapi3.NewStringSchema()),
		),
	}

	openapiSpec.Components.RequestBodies = openapi3.RequestBodies{
		"postCSRRequest": &openapi3.RequestBodyRef{
			Value: openapi3.NewRequestBody().
				WithDescription("Request used for creating a new CSR ").
				WithRequired(true).
				WithJSONSchema(openapi3.NewSchema().
					WithProperty("data", openapi3.NewStringSchema()).
					WithProperty("dms_name", openapi3.NewStringSchema()).
					WithProperty("key_type", openapi3.NewStringSchema()),
				),
		},
		"postCSRFormRequest": &openapi3.RequestBodyRef{
			Value: openapi3.NewRequestBody().
				WithDescription("Request used for creating a new CSR Form").
				WithRequired(true).
				WithJSONSchema(openapi3.NewSchema().
					WithPropertyRef("subject", &openapi3.SchemaRef{
						Ref: "#/components/schemas/CSRForm",
					}),
				),
		},
		"getPendingCSRRequest": &openapi3.RequestBodyRef{
			Value: openapi3.NewRequestBody().
				WithDescription("Request used for creating a new CSR Form").
				WithRequired(true).
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema().
					WithProperty("ID", openapi3.NewIntegerSchema()))),
		},
	}

	openapiSpec.Components.Responses = openapi3.Responses{
		"ErrorResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response when errors happen.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema().
					WithProperty("error", openapi3.NewStringSchema()))),
		},
		"HealthResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after healthchecking.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema().
					WithProperty("healthy", openapi3.NewBoolSchema())),
				),
		},
		"PostCSRResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after getting a CA.").
				WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
					Ref: "#/components/schemas/CSR",
				})),
		},
		"PostCSRFormResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after creating a CSR.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema().
					WithPropertyRef("csr", &openapi3.SchemaRef{
						Ref: "#/components/schemas/CSR",
					}).
					WithProperty("priv_key", openapi3.NewStringSchema())),
				),
		},
		"GetPendingCSRsResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after getting pending CSRs.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema().
					WithPropertyRef("CSRs", arrayOf(&openapi3.SchemaRef{
						Ref: "#/components/schemas/CSR",
					}))),
				),
		},

		"GetPendingCSRDBResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after getting pending CSRDB.").
				WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
					Ref: "#/components/schemas/CSR",
				})),
		},
		//TODO
		"GetPendingCSRFileResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after getting CSR File.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema())),
		},
		"PutChangeCSRStatusResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after changing CSR Status.").
				WithContent(openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
					Ref: "#/components/schemas/CSR",
				})),
		},
		"DeleteCSRResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after deleting CSR.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema())),
		},
		"GetCRTResponse": &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Response returned back after deleting CSR.").
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewSchema())),
		},
	}

	openapiSpec.Paths = openapi3.Paths{
		"/v1/health": &openapi3.PathItem{
			Get: &openapi3.Operation{
				OperationID: "Health",
				Description: "Get health status",
				Responses: openapi3.Responses{
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/HealthResponse",
					},
				},
			},
		},
		"/v1/csrs/{name}": &openapi3.PathItem{
			Post: &openapi3.Operation{
				OperationID: "PostCSR",
				Description: "Post CSR",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("name").
							WithSchema(openapi3.NewStringSchema()),
					},
				},
				RequestBody: &openapi3.RequestBodyRef{
					Ref: "#/components/requestBodies/postCSRRequest",
				},
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/PostCSRResponse",
					},
				},
			},
		},
		"/v1/csrs/{name}/form": &openapi3.PathItem{
			Post: &openapi3.Operation{
				OperationID: "PostCSRForm",
				Description: "Post CSR Form",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("name").
							WithSchema(openapi3.NewStringSchema()),
					},
				},
				RequestBody: &openapi3.RequestBodyRef{
					Ref: "#/components/requestBodies/postCSRFormRequest",
				},
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/PostCSRFormResponse",
					},
				},
			},
		},
		"/v1/csrs": &openapi3.PathItem{
			Get: &openapi3.Operation{
				OperationID: "GetPendingCSRs",
				Description: "Get Pending CSRs",
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/GetPendingCSRsResponse",
					},
				},
			},
		},
		"/v1/csrs/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				OperationID: "GetPendingCSRDB",
				Description: "Get Pending CSRDB by id",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("id").
							WithSchema(openapi3.NewIntegerSchema()),
					},
				},
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/GetPendingCSRDBResponse",
					},
				},
			},
			Put: &openapi3.Operation{
				OperationID: "PutChangeCSRStatus",
				Description: "Change CSR Status by id",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("id").
							WithSchema(openapi3.NewIntegerSchema()),
					},
				},
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/PutChangeCSRStatusResponse",
					},
				},
			},
			Delete: &openapi3.Operation{
				OperationID: "DeleteCSR",
				Description: "Delete CSR by id",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("id").
							WithSchema(openapi3.NewIntegerSchema()),
					},
				},
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/DeleteCSRResponse",
					},
				},
			},
		},

		"/v1/csrs/{id}/file": &openapi3.PathItem{
			Get: &openapi3.Operation{
				OperationID: "GetPendingCSRFile",
				Description: "Get Pending CSR Fileby id",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("id").
							WithSchema(openapi3.NewIntegerSchema()),
					},
				},
				/*RequestBody: &openapi3.RequestBodyRef{
					Ref: "#/components/requestBodies/getPendingCSRRequest",
				},*/

				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/GetPendingCSRFileResponse",
					},
				},
			},
		},
		"/v1/csrs/{id}/crt": &openapi3.PathItem{
			Get: &openapi3.Operation{
				OperationID: "GetCRT",
				Description: "Get CRT by id",
				Parameters: []*openapi3.ParameterRef{
					{
						Value: openapi3.NewPathParameter("id").
							WithSchema(openapi3.NewIntegerSchema()),
					},
				},
				Responses: openapi3.Responses{
					"400": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"401": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"403": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"500": &openapi3.ResponseRef{
						Ref: "#/components/responses/ErrorResponse",
					},
					"200": &openapi3.ResponseRef{
						Ref: "#/components/responses/GetCRTResponse",
					},
				},
			},
		},
	}

	return openapiSpec
}
