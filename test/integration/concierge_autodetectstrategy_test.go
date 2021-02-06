// Copyright 2020-2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	loginv1alpha1 "go.pinniped.dev/generated/1.20/apis/concierge/login/v1alpha1"
	"go.pinniped.dev/test/library"
)

// see: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html
const AWSMetadataEndpointDefault = "http://169.254.169.254/latest/dynamic/instance-identity/document"

//
const AzureMetadataEndpointDefault = "http://169.254.169.254/metadata/instance?api-version=2018-10-01"

// see: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#gke_mds
const GCPMetaDataEndpointDefault = "http://metadata.google.internal/computeMetadata/v1/instance/zone"

func TestAutodetectStrategy(t *testing.T) {
	//env := library.IntegrationEnv(t) //.WithCapability(library.ClusterSigningKeyIsAvailable)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	authenticator := library.CreateTestWebhookAuthenticator(ctx, t)
	//_, username, groups := library.IntegrationEnv(t).TestUser.Token, env.TestUser.ExpectedUsername, env.TestUser.ExpectedGroups
	//
	var response *loginv1alpha1.TokenCredentialRequest

	t.Logf("making token credential request")
	response, err := makeRequest(ctx, t, validCredentialRequestSpecWithRealToken(t, authenticator))
	require.NoError(t, err)
	t.Logf("response message: " + *response.Status.Message)

	if "authentication failed" == *response.Status.Message {
		t.Logf("found hosted cluster")
		// not self hosted
		t.Logf("&&&&&&&&&&&&&&&&&&&&&&&&&&&&HOSTED CLUSTER TYPE: " + getHostedClusterType(t))
	} else {
		t.Logf("found self service cluster")
	}

	//// Create a client using the admin kubeconfig.
	//adminClient := library.NewClientset(t)
	//
	//// Create a client using the certificate from the CredentialRequest.
	//clientWithCertFromCredentialRequest := library.NewClientsetWithCertAndKey(
	//	t,
	//	response.Status.Credential.ClientCertificateData,
	//	response.Status.Credential.ClientKeyData,
	//)
	//
	//t.Run(
	//	"access as user",
	//	library.AccessAsUserTest(ctx, adminClient, username, clientWithCertFromCredentialRequest),
	//)
	//for _, group := range groups {
	//	group := group
	//	t.Run(
	//		"access as group "+group,
	//		library.AccessAsGroupTest(ctx, adminClient, group, clientWithCertFromCredentialRequest),
	//	)
	//}
}

//func makeRequest(ctx context.Context, t *testing.T, spec loginv1alpha1.TokenCredentialRequestSpec) (*loginv1alpha1.TokenCredentialRequest, error) {
//	t.Helper()
//	env := library.IntegrationEnv(t)
//
//	client := library.NewAnonymousConciergeClientset(t)
//
//	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
//	defer cancel()
//
//	return client.LoginV1alpha1().TokenCredentialRequests(env.ConciergeNamespace).Create(ctx, &loginv1alpha1.TokenCredentialRequest{
//		TypeMeta:   metav1.TypeMeta{},
//		ObjectMeta: metav1.ObjectMeta{Namespace: env.ConciergeNamespace},
//		Spec:       spec,
//	}, metav1.CreateOptions{})
//}

//func validCredentialRequestSpecWithRealToken(t *testing.T, authenticator corev1.TypedLocalObjectReference) loginv1alpha1.TokenCredentialRequestSpec {
//	return loginv1alpha1.TokenCredentialRequestSpec{
//		Token:         library.IntegrationEnv(t).TestUser.Token,
//		Authenticator: authenticator,
//	}
//}
//
//func stringPtr(s string) *string {
//	return &s
//}

// experimenting with determining the region
func getHostedClusterType(t *testing.T) string {
	//Check if cloud-provider is AWS
	region, err := getAWSRegion(AWSMetadataEndpointDefault)
	if err != nil {
		t.Logf("error determining aws region: %v", err)
	} else if region != "" {
		t.Logf("cloudprovider is aws")
		return "aws"
	}

	//Check if cloud-provider is GCP
	region, err = getGCPRegion(GCPMetaDataEndpointDefault)
	if err != nil {
		t.Logf("error determining gcp region: %v", err)
	} else if region != "" {
		t.Logf("cloudprovider is gcp")
		return "gcp"
	}

	//Check if cloud-provider is AZURE
	region, err = getAzureRegion(AzureMetadataEndpointDefault)
	if err != nil {
		t.Logf("error determining azure region: %v", err)
	} else if region != "" {
		t.Logf("cloudprovider is azure")
		return "azure"
	}

	return "not found"
}

func getAWSRegion(endpoint string) (string, error) {
	body, err := metadataServiceRequest(endpoint)
	if err != nil {
		return "", err
	}

	metadata := struct {
		Region string `json:"region"`
	}{}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return "", err
	}
	return metadata.Region, nil
}

func getAzureRegion(endpoint string) (string, error) {
	body, err := metadataServiceRequest(endpoint)
	if err != nil {
		return "", err
	}

	metadata := struct {
		Compute struct {
			Location string `json:"location"`
		} `json:"compute"`
	}{}
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return "", err
	}
	return metadata.Compute.Location, nil
}

func getGCPRegion(endpoint string) (string, error) {
	body, err := metadataServiceRequest(endpoint)
	if err != nil {
		return "", err
	}

	metadata := strings.Split(string(body), "/")
	region := metadata[len(metadata)-1]
	return strings.TrimRight(region, "-abcdefghij"), nil
}

func metadataServiceRequest(endpoint string) ([]byte, error) {
	httpClient := http.Client{Timeout: time.Duration(5 * time.Second)}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// cloud provider metadata specific headers
	req.Header.Set("Metadata-Flavor", "Google") // Google
	req.Header.Set("Metadata", "true")          // Azure
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//ensure that status code is 200
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to metadata status-service failed with error-code: %d(%s)",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
