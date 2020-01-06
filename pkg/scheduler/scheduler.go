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
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

var GlobalMetrics metrics.Metrics
var NodeResource metrics.Metrics
var NodeMetricsCache = make(map[string]*NodeMetrics)
var TimingMap = make(map[string]int64)

var PenaltyMap = make(map[string]float32)
var PenaltyTiming = make(map[string]int)
var PredictionPenalty float32
var MaxPenalty = float32(2)
var MinPenalty = float32(1.0)
var PenaltyUpdate float32
var StopUpdate = false
var TargetQoS float32
var PenaltyTimeout int
var PrevPredictions map[string]*NodeMetrics
var prevQoS = float32(1.0)
var KeepScheduling = true
var KeepSchedulingTimeout = 10000

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

type NodeMetrics struct {
	Usage       nodeinfo.Resource
	Allocatable nodeinfo.Resource
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// Estimate predict resource usage
var updatePenaltyRule = 5 // 0: fix, others is dynamic
var penaltyUpdated = false

func Estimate(nodeNames []string) map[string]*NodeMetrics {
	if updatePenaltyRule == 0 {
		//do nothing
	} else if updatePenaltyRule == 1 {
		// update min so prediction penalty will converge...
		if GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).PendingPodsNum > 0 {
			qos := GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).QualityOfService
			if qos < TargetQoS {
				if penaltyUpdated {
					MinPenalty = PredictionPenalty
					penaltyUpdated = false
				}
				PredictionPenalty = MaxPenalty
			} else if qos > TargetQoS {
				PredictionPenalty = max(PredictionPenalty*PenaltyUpdate, MinPenalty)
				penaltyUpdated = true
			}
		}
	} else if updatePenaltyRule == 2 {
		// go from max to min.
		if GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).PendingPodsNum > 0 {
			qos := GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).QualityOfService
			if qos < TargetQoS {
				PredictionPenalty = MaxPenalty
			} else if qos > TargetQoS {
				PredictionPenalty = max(PredictionPenalty*PenaltyUpdate, MinPenalty)
			}
		}
	} else if updatePenaltyRule == 3 { // okay but it cannot deal with high demand when prediction penalty converge.
		// we can start at 1.1
		// update max & min so Prediction penalty will converge...
		if GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).PendingPodsNum > 0 {
			qos := GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).QualityOfService
			if qos < TargetQoS {
				tmp := MaxPenalty
				if penaltyUpdated {
					MinPenalty = PredictionPenalty
					MaxPenalty = (MaxPenalty + MinPenalty) / 2
					penaltyUpdated = false
				}
				PredictionPenalty = tmp
			} else if qos > TargetQoS {
				PredictionPenalty = max(PredictionPenalty*PenaltyUpdate, MinPenalty)
				penaltyUpdated = true
			}
		}
	} else if updatePenaltyRule == 4 { // good.
		MinPenalty = 1.1
		PenaltyUpdate = 0.99
		// update max & min so Prediction penalty will converge...
		if GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).PendingPodsNum > 0 {
			qos := GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).QualityOfService
			if qos < TargetQoS {
				if penaltyUpdated {
					PredictionPenalty += (PredictionPenalty - 1.0)
					penaltyUpdated = false
				}
			} else if qos > TargetQoS {
				PredictionPenalty = max(PredictionPenalty*PenaltyUpdate, MinPenalty)
				penaltyUpdated = true
			}
		}
	} else if updatePenaltyRule == 5 {
		MinPenalty = 1.1
		PenaltyUpdate = 0.99
		// update max & min so Prediction penalty will converge...
		if GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).PendingPodsNum > 0 {
			qos := GlobalMetrics[metrics.QueueMetricsKey].(queue.Metrics).QualityOfService
			if qos < TargetQoS {
				if penaltyUpdated || qos < (prevQoS*0.99) {
					PredictionPenalty += (PredictionPenalty - 1.0)
					penaltyUpdated = false
				}
			} else if qos > TargetQoS {
				PredictionPenalty = max(PredictionPenalty*PenaltyUpdate, MinPenalty)
				penaltyUpdated = true
			}
			prevQoS = qos
		}
	}

	nodeMetricsMap := make(map[string]*NodeMetrics)
	// predict.
	for _, nodeName := range nodeNames {
		usage := *nodeinfo.NewResource(GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].TotalResourceUsage)
		cap := *nodeinfo.NewResource(GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].Allocatable)
		usage.MilliCPU = usage.MilliCPU * int64(PredictionPenalty*100) / 100
		usage.Memory = usage.Memory * int64(PredictionPenalty*100) / 100
		nodeMetricsMap[nodeName] = &NodeMetrics{
			Usage:       usage,
			Allocatable: cap,
		}
	}

	return nodeMetricsMap
}
