// Copyright 2019 Preferred Networks, Inc.
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

package main

import (
	"fmt"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/scheduler"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

func filterExtender(args api.ExtenderArgs) api.ExtenderFilterResult {
	// Filters out no nodes.
	return api.ExtenderFilterResult{
		Nodes:       &v1.NodeList{},
		NodeNames:   args.NodeNames,
		FailedNodes: api.FailedNodesMap{},
		Error:       "",
	}
}

func prioritizeExtender(args api.ExtenderArgs) api.HostPriorityList {
	// Ranks all nodes equally.
	priorities := make(api.HostPriorityList, 0, len(*args.NodeNames))
	for _, name := range *args.NodeNames {
		priorities = append(priorities, api.HostPriority{Host: name, Score: 1})
	}
	fmt.Printf("priorities %v", priorities)
	return priorities
}

func prioritizeLowUsageNode(args api.ExtenderArgs) api.HostPriorityList {
	nodeMetricsMap := Monitor(*args.NodeNames)
	// Ranks all nodes equally.
	priorities := make(api.HostPriorityList, 0, len(*args.NodeNames))
	for _, name := range *args.NodeNames {
		usage := nodeinfo.NewResource(nodeMetricsMap[name].Usage)
		capacity := nodeinfo.NewResource(nodeMetricsMap[name].Allocatable)
		score := int(api.MaxPriority * (capacity.MilliCPU - usage.MilliCPU) / capacity.MilliCPU)
		priorities = append(priorities, api.HostPriority{Host: name, Score: score})
	}
	fmt.Printf("priorities %v", priorities)
	return priorities
}

func Monitor(nodeNames []string) map[string]*scheduler.NodeMetrics {
	nodeMetricsMap := make(map[string]*scheduler.NodeMetrics)
	for _, nodeName := range nodeNames {
		nodeMetrics := &scheduler.NodeMetrics{
			Name:        nodeName,
			Usage:       scheduler.GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].TotalResourceUsage,
			Allocatable: scheduler.GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].Allocatable,
		}
		nodeMetricsMap[nodeName] = nodeMetrics
	}
	return nodeMetricsMap
}
