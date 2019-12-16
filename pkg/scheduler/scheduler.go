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

package scheduler

import (
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

var GlobalMetrics metrics.Metrics
var NodeMetricsArray []*NodeMetrics

var penaltyMap map[string]float32
var penaltyTiming map[string]int
var predictionPenalty float32
var penaltyTimeout int
var prevPredictions []*NodeMetrics

// Scheduler defines the lowest-level scheduler interface.
type Scheduler interface {
	// Schedule makes scheduling decisions for (subset of) pending pods and running pods.
	// The return value is a list of scheduling events.
	// This method must never block.
	Schedule(
		clock clock.Clock,
		podQueue queue.PodQueue,
		nodeLister algorithm.NodeLister,
		nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error)
}

// Event defines the interface of a scheduling event.
// Submit can returns any type in a list that implements this interface.
type Event interface {
	IsSchedulerEvent() bool
}

// BindEvent represents an event of deciding the binding of a pod to a node.
type BindEvent struct {
	Pod            *v1.Pod
	ScheduleResult core.ScheduleResult
}

// DeleteEvent represents an event of the deleting a bound pod on a node.
type DeleteEvent struct {
	PodNamespace string
	PodName      string
	NodeName     string
}

func (b *BindEvent) IsSchedulerEvent() bool   { return true }
func (d *DeleteEvent) IsSchedulerEvent() bool { return true }

// NodeMetrics contains node's name & metrics
type NodeMetrics struct {
	Name        string
	Usage       v1.ResourceList
	Allocatable v1.ResourceList
}

// Monitor monitors metrics.
func Monitor(nodeInfoMap map[string]*nodeinfo.NodeInfo) map[string]*NodeMetrics {
	nodeMetricsMap := make(map[string]*NodeMetrics)
	for nodeName := range nodeInfoMap {
		nodeMetrics := &NodeMetrics{
			Name:        nodeName,
			Usage:       GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].TotalResourceUsage,
			Allocatable: GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].Allocatable,
		}
		nodeMetricsMap[nodeName] = nodeMetrics
	}
	return nodeMetricsMap
}

// Estimate predict resource usage
func Estimate(nodeInfoMap map[string]*nodeinfo.NodeInfo) []*NodeMetrics {
	// monitorMap := monitor(nodeInfoMap)
	nodeMetricsArray := make([]*NodeMetrics, 0, len(nodeInfoMap))
	// predict.
	for nodeName := range nodeInfoMap {
		nodeMetrics := &NodeMetrics{
			Name:        nodeName,
			Usage:       GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].TotalResourceUsage,
			Allocatable: GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].Allocatable,
		}
		nodeMetricsArray = append(nodeMetricsArray, nodeMetrics)
		if _, ok := penaltyMap[nodeName]; !ok {
			penaltyMap[nodeName] = predictionPenalty
		}
	}

	// react to prediction errors.
	if prevPredictions != nil {
		for i, p := range prevPredictions {
			m := nodeMetricsArray[i]
			if !util.ResourceListGE(p.Usage, m.Usage) {
				penaltyMap[m.Name] = penaltyMap[m.Name] * predictionPenalty
				penaltyTiming[m.Name] = 0
			} else if util.ResourceListGE(p.Usage, m.Usage) {
				penaltyTiming[m.Name]++
				if penaltyTiming[m.Name] >= penaltyTimeout {
					penaltyMap[m.Name] = predictionPenalty
				}
			} else {
				penaltyTiming[m.Name] = 0
			}
			nodeMetricsArray[i].Usage = util.ResourceListMultiply(m.Usage, penaltyMap[m.Name])
		}
	}

	prevPredictions = nodeMetricsArray

	return nodeMetricsArray
}
