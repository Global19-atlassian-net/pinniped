/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/

// Package loginrequest provides REST functionality for the LoginRequest resource.
package loginrequest

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog/v2"

	placeholderapi "github.com/suzerain-io/placeholder-name-api/pkg/apis/placeholder"
)

var (
	_ rest.Creater                 = &REST{}
	_ rest.NamespaceScopedStrategy = &REST{}
	_ rest.Scoper                  = &REST{}
	_ rest.Storage                 = &REST{}
)

func NewREST(webhook authenticator.Token) *REST {
	return &REST{
		webhook: webhook,
	}
}

type REST struct {
	webhook authenticator.Token
}

func (r *REST) New() runtime.Object {
	return &placeholderapi.LoginRequest{}
}

func (r *REST) NamespaceScoped() bool {
	return false
}

func (r *REST) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	loginRequest, ok := obj.(*placeholderapi.LoginRequest)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("not a LoginRequest: %#v", obj))
	}

	// TODO refactor all validation checks into a validation function in another package (e.g. see subjectaccessreqview api in k8s)

	if len(loginRequest.Spec.Type) == 0 {
		errs := field.ErrorList{field.Required(field.NewPath("spec", "type"), "type must be supplied")}
		return nil, apierrors.NewInvalid(placeholderapi.Kind(loginRequest.Kind), loginRequest.Name, errs)
	}

	if loginRequest.Spec.Type != placeholderapi.TokenLoginCredentialType {
		errs := field.ErrorList{field.Invalid(field.NewPath("spec", "type"), loginRequest.Spec.Type, "unrecognized type")}
		return nil, apierrors.NewInvalid(placeholderapi.Kind(loginRequest.Kind), loginRequest.Name, errs)
	}

	token := loginRequest.Spec.Token
	if token == nil || len(token.Value) == 0 {
		errs := field.ErrorList{field.Required(field.NewPath("spec", "token", "value"), "token must be supplied")}
		return nil, apierrors.NewInvalid(placeholderapi.Kind(loginRequest.Kind), loginRequest.Name, errs)
	}

	// let dynamic admission webhooks have a chance to validate (but not mutate) as well
	// TODO Are we okay with admission webhooks being able to see tokens? Maybe strip token out before passing obj to createValidation.
	//  Since we are an aggregated API, we should investigate to see if the kube API server is already invoking admission hooks for us.
	//  Even if it is, its okay to call it again here. If the kube API server is already calling the webhooks and passing
	//  the token, then there is probably no reason for us to avoid passing the token when we call the webhooks here, since
	//  they already got the token.
	if createValidation != nil {
		if err := createValidation(ctx, obj.DeepCopyObject()); err != nil {
			return nil, err
		}
	}

	// just a sanity check, not sure how to honor a dry run on a virtual API
	if options != nil {
		if len(options.DryRun) != 0 {
			errs := field.ErrorList{field.NotSupported(field.NewPath("dryRun"), options.DryRun, nil)}
			return nil, apierrors.NewInvalid(placeholderapi.Kind(loginRequest.Kind), loginRequest.Name, errs)
		}
	}

	// the incoming context could have an audience attached to it technically
	// sine we do not want to handle audiences right now, do not pass it through directly
	// instead we just propagate cancellation of the parent context
	cancelCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-cancelCtx.Done():
		}
	}()

	authResponse, authenticated, err := r.webhook.AuthenticateToken(cancelCtx, token.Value)
	if err != nil {
		klog.Warningf("webhook authentication failure: %v", err)
		return failureResponse(), nil
	}

	var out *placeholderapi.LoginRequest
	if authenticated {
		out = successfulResponse(authResponse)
	} else {
		out = failureResponse()
	}

	return out, nil
}

func successfulResponse(authResponse *authenticator.Response) *placeholderapi.LoginRequest {
	return &placeholderapi.LoginRequest{
		Status: placeholderapi.LoginRequestStatus{
			Credential: &placeholderapi.LoginRequestCredential{
				ExpirationTimestamp:   nil,
				Token:                 "snorlax",
				ClientCertificateData: "",
				ClientKeyData:         "",
			},
			User: &placeholderapi.User{
				Name:   authResponse.User.GetName(),
				Groups: authResponse.User.GetGroups(),
			},
		},
	}
}

func failureResponse() *placeholderapi.LoginRequest {
	return &placeholderapi.LoginRequest{
		Status: placeholderapi.LoginRequestStatus{
			Credential: nil,
			Message:    "authentication failed",
		},
	}
}
