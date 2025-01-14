// Copyright 2021 Antrea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clustergroupmember

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	"antrea.io/antrea/pkg/apis/controlplane"
)

type REST struct {
	querier groupMembershipQuerier
}

var (
	_ rest.Storage = &REST{}
	_ rest.Scoper  = &REST{}
	_ rest.Getter  = &REST{}
)

// NewREST returns a REST object that will work against API services.
func NewREST(querier groupMembershipQuerier) *REST {
	return &REST{querier}
}

type groupMembershipQuerier interface {
	GetGroupMembers(name string) (controlplane.GroupMemberSet, []controlplane.IPBlock, error)
}

func (r *REST) New() runtime.Object {
	return &controlplane.ClusterGroupMembers{}
}

func (r *REST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	groupMembers, ipBlocks, err := r.querier.GetGroupMembers(name)
	if err != nil {
		return nil, errors.NewInternalError(err)
	}
	memberList := &controlplane.ClusterGroupMembers{}
	if len(ipBlocks) > 0 {
		effectiveIPBlocks := make([]controlplane.IPNet, 0, len(ipBlocks))
		for _, ipb := range ipBlocks {
			// ClusterGroup ipBlock does not support Except slices, so no need to generate an effective
			// list of IPs by removing Except slices from allowed CIDR.
			effectiveIPBlocks = append(effectiveIPBlocks, ipb.CIDR)
		}
		memberList.EffectiveIPBlocks = effectiveIPBlocks
	} else {
		effectiveMembers := make([]controlplane.GroupMember, 0, len(groupMembers))
		for _, member := range groupMembers {
			effectiveMembers = append(effectiveMembers, *member)
		}
		memberList.EffectiveMembers = effectiveMembers
	}
	memberList.Name = name
	return memberList, nil
}

func (r *REST) NamespaceScoped() bool {
	return false
}
