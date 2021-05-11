package waiter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ProductReadyTimeout  = 3 * time.Minute
	ProductDeleteTimeout = 3 * time.Minute

	ProvisioningArtifactReadyTimeout   = 3 * time.Minute
	ProvisioningArtifactDeletedTimeout = 3 * time.Minute

	StatusNotFound    = "NOT_FOUND"
	StatusUnavailable = "UNAVAILABLE"

	// AWS documentation is wrong, says that status will be "AVAILABLE" but it is actually "CREATED"
	StatusCreated = "CREATED"
)

func ProductReady(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable, StatusCreated},
		Refresh: ProductStatus(conn, acceptLanguage, productID),
		Timeout: ProductReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProductAsAdminOutput); ok {
		return output, err
	}

	return nil, err
}

func ProductDeleted(conn *servicecatalog.ServiceCatalog, acceptLanguage, productID string) (*servicecatalog.DescribeProductAsAdminOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: ProductStatus(conn, acceptLanguage, productID),
		Timeout: ProductReadyTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, nil
	}

	return nil, err
}

func ProvisioningArtifactReady(conn *servicecatalog.ServiceCatalog, id, productID string) (*servicecatalog.DescribeProvisioningArtifactOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, StatusNotFound, StatusUnavailable},
		Target:  []string{servicecatalog.StatusAvailable, StatusCreated},
		Refresh: ProvisioningArtifactStatus(conn, id, productID),
		Timeout: ProvisioningArtifactReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicecatalog.DescribeProvisioningArtifactOutput); ok {
		return output, err
	}

	return nil, err
}

func ProvisioningArtifactDeleted(conn *servicecatalog.ServiceCatalog, id, productID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating, servicecatalog.StatusAvailable, StatusCreated, StatusUnavailable},
		Target:  []string{StatusNotFound},
		Refresh: ProvisioningArtifactStatus(conn, id, productID),
		Timeout: ProvisioningArtifactDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for state of provisioning artifact (%s): %w", id, err)
	}

	return nil
}
