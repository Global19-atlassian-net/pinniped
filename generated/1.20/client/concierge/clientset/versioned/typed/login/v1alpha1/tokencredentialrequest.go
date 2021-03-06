// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "go.pinniped.dev/generated/1.20/apis/concierge/login/v1alpha1"
	scheme "go.pinniped.dev/generated/1.20/client/concierge/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TokenCredentialRequestsGetter has a method to return a TokenCredentialRequestInterface.
// A group's client should implement this interface.
type TokenCredentialRequestsGetter interface {
	TokenCredentialRequests(namespace string) TokenCredentialRequestInterface
}

// TokenCredentialRequestInterface has methods to work with TokenCredentialRequest resources.
type TokenCredentialRequestInterface interface {
	Create(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.CreateOptions) (*v1alpha1.TokenCredentialRequest, error)
	Update(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.UpdateOptions) (*v1alpha1.TokenCredentialRequest, error)
	UpdateStatus(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.UpdateOptions) (*v1alpha1.TokenCredentialRequest, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.TokenCredentialRequest, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TokenCredentialRequestList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TokenCredentialRequest, err error)
	TokenCredentialRequestExpansion
}

// tokenCredentialRequests implements TokenCredentialRequestInterface
type tokenCredentialRequests struct {
	client rest.Interface
	ns     string
}

// newTokenCredentialRequests returns a TokenCredentialRequests
func newTokenCredentialRequests(c *LoginV1alpha1Client, namespace string) *tokenCredentialRequests {
	return &tokenCredentialRequests{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the tokenCredentialRequest, and returns the corresponding tokenCredentialRequest object, and an error if there is any.
func (c *tokenCredentialRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	result = &v1alpha1.TokenCredentialRequest{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TokenCredentialRequests that match those selectors.
func (c *tokenCredentialRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TokenCredentialRequestList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.TokenCredentialRequestList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested tokenCredentialRequests.
func (c *tokenCredentialRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a tokenCredentialRequest and creates it.  Returns the server's representation of the tokenCredentialRequest, and an error, if there is any.
func (c *tokenCredentialRequests) Create(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.CreateOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	result = &v1alpha1.TokenCredentialRequest{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tokenCredentialRequest).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a tokenCredentialRequest and updates it. Returns the server's representation of the tokenCredentialRequest, and an error, if there is any.
func (c *tokenCredentialRequests) Update(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.UpdateOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	result = &v1alpha1.TokenCredentialRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		Name(tokenCredentialRequest.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tokenCredentialRequest).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *tokenCredentialRequests) UpdateStatus(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.UpdateOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	result = &v1alpha1.TokenCredentialRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		Name(tokenCredentialRequest.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(tokenCredentialRequest).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the tokenCredentialRequest and deletes it. Returns an error if one occurs.
func (c *tokenCredentialRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *tokenCredentialRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched tokenCredentialRequest.
func (c *tokenCredentialRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TokenCredentialRequest, err error) {
	result = &v1alpha1.TokenCredentialRequest{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("tokencredentialrequests").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
