// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package supervisorconfig

import (
	"encoding/json"
	"fmt"

	"gopkg.in/square/go-jose.v2"
	"k8s.io/apimachinery/pkg/labels"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/klog/v2"

	"go.pinniped.dev/generated/1.19/client/supervisor/informers/externalversions/config/v1alpha1"
	pinnipedcontroller "go.pinniped.dev/internal/controller"
	"go.pinniped.dev/internal/controllerlib"
)

type jwksObserverController struct {
	issuerToJWKSSetter   IssuerToJWKSMapSetter
	oidcProviderInformer v1alpha1.OIDCProviderInformer
	secretInformer       corev1informers.SecretInformer
}

type IssuerToJWKSMapSetter interface {
	SetIssuerToJWKSMap(issuerToJWKSMap map[string]*jose.JSONWebKeySet)
}

// Returns a controller which watches all of the OIDCProviders and their corresponding Secrets
// and fills an in-memory cache of the JWKS info for each currently configured issuer.
// This controller assumes that the informers passed to it are already scoped down to the
// appropriate namespace. It also assumes that the IssuerToJWKSMapSetter passed to it has an
// underlying implementation which is thread-safe.
func NewJWKSObserverController(
	issuerToJWKSSetter IssuerToJWKSMapSetter,
	secretInformer corev1informers.SecretInformer,
	oidcProviderInformer v1alpha1.OIDCProviderInformer,
	withInformer pinnipedcontroller.WithInformerOptionFunc,
) controllerlib.Controller {
	return controllerlib.New(
		controllerlib.Config{
			Name: "jwks-observer-controller",
			Syncer: &jwksObserverController{
				issuerToJWKSSetter:   issuerToJWKSSetter,
				oidcProviderInformer: oidcProviderInformer,
				secretInformer:       secretInformer,
			},
		},
		withInformer(
			secretInformer,
			pinnipedcontroller.MatchAnythingFilter(nil),
			controllerlib.InformerOption{},
		),
		withInformer(
			oidcProviderInformer,
			pinnipedcontroller.MatchAnythingFilter(nil),
			controllerlib.InformerOption{},
		),
	)
}

func (c *jwksObserverController) Sync(ctx controllerlib.Context) error {
	ns := ctx.Key.Namespace
	allProviders, err := c.oidcProviderInformer.Lister().OIDCProviders(ns).List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list OIDCProviders: %w", err)
	}

	// Rebuild the whole map on any change to any Secret or OIDCProvider, because either can have changes that
	// can cause the map to need to be updated.
	issuerToJWKSMap := map[string]*jose.JSONWebKeySet{}

	for _, provider := range allProviders {
		secretRef := provider.Status.JWKSSecret
		jwksSecret, err := c.secretInformer.Lister().Secrets(ns).Get(secretRef.Name)
		if err != nil {
			klog.InfoS("jwksObserverController Sync could not find JWKS secret", "namespace", ns, "secretName", secretRef.Name)
			continue
		}
		jwkFromSecret := jose.JSONWebKeySet{}
		err = json.Unmarshal(jwksSecret.Data[jwksKey], &jwkFromSecret)
		if err != nil {
			klog.InfoS("jwksObserverController Sync found a JWKS secret with Data in an unexpected format", "namespace", ns, "secretName", secretRef.Name)
			continue
		}
		issuerToJWKSMap[provider.Spec.Issuer] = &jwkFromSecret
	}

	klog.InfoS("jwksObserverController Sync updated the JWKS cache", "issuerCount", len(issuerToJWKSMap))
	c.issuerToJWKSSetter.SetIssuerToJWKSMap(issuerToJWKSMap)

	return nil
}
