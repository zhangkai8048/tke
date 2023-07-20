/*
 * Harbor API
 *
 * These APIs provide services for manipulating Harbor project.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package helmClient
import (
	"os"
)

type HarborBody1 struct {
	// The provance file
	Prov **os.File `json:"prov"`
}
