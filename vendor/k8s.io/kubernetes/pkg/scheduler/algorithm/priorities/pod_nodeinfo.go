/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package priorities

import (
	v1 "k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// PodNodeInfoPriority compute the priority for pod based on nodeinfo
type PodNodeInfoPriority struct {
	Name   string
	scorer func(*v1.Pod, *schedulernodeinfo.NodeInfo) int64
}

func (p *PodNodeInfoPriority) PriorityMap(
	pod *v1.Pod,
	meta interface{},
	nodeInfo *schedulernodeinfo.NodeInfo) (schedulerapi.HostPriority, error) {
	node := nodeInfo.Node()
	score := p.scorer(pod, nodeInfo)
	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: int(score),
	}, nil
}
