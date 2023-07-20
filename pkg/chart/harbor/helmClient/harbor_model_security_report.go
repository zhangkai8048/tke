/*
 * Harbor API
 *
 * These APIs provide services for manipulating Harbor project.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package helmClient

// The security information of the chart
type HarborSecurityReport struct {
	Signature *HarborDigitalSignature `json:"signature,omitempty"`
}
