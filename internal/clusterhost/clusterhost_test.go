// Copyright 2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterhost

import (
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	coretesting "k8s.io/client-go/testing"

	v1 "k8s.io/api/core/v1"
)

func TestHasControlPlaneNodes(t *testing.T) {
	tests := []struct {
		name            string
		nodes           []*v1.Node
		listNodesErr    error
		wantErr         error
		wantReturnValue bool
	}{
		{
			name:         "Fetching nodes returns an error",
			listNodesErr: errors.New("couldn't get nodes"),
			wantErr:      errors.New("error fetching nodes: couldn't get nodes"),
		},
		{
			name:    "Fetching nodes returns an empty array",
			nodes:   []*v1.Node{},
			wantErr: errors.New("no nodes found"),
		},
		{
			name: "Nodes found, but not control plane nodes",
			nodes: []*v1.Node{{ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			}},
			},
		},
	}
	for _, tt := range tests {
		kubeClient := kubernetesfake.NewSimpleClientset()
		if tt.listNodesErr != nil {
			listNodesErr := tt.listNodesErr
			kubeClient.PrependReactor(
				"list",
				"nodes",
				func(_ coretesting.Action) (bool, runtime.Object, error) {
					return true, nil, listNodesErr
				},
			)
		}
		for _, node := range tt.nodes {
			err := kubeClient.Tracker().Add(node)
			require.NoError(t, err)
		}
		clusterHost := New(kubeClient)
		hasControlPlaneNodes, err := clusterHost.HasControlPlaneNodes(context.Background())
		require.Equal(t, tt.wantErr, err)
		require.Equal(t, tt.wantReturnValue, hasControlPlaneNodes)
	}
}
