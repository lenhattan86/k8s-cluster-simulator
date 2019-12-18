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

package metrics

import (
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// Metrics represents a metrics at one time point, in the following structure.
//   Metrics[ClockKey] = a formatted clock
//   Metrics[NodesMetricsKey] = map from node name to node.Metrics
//   Metrics[PodsMetricsKey] = map from pod name to pod.Metrics
// 	 Metrics[QueueMetricsKey] = queue.Metrics
type Metrics map[string]interface{}

const (
	// ClockKey is the key associated to a clock.Clock.
	ClockKey = "Clock"
	// NodesMetricsKey is the key associated to a map of node.Metrics.
	NodesMetricsKey = "Nodes"
	// PodsMetricsKey is the key associated to a map of pod.Metrics.
	PodsMetricsKey = "Pods"
	// QueueMetricsKey is the key associated to a queue.Metrics.
	QueueMetricsKey = "Queue"
)

func whichSharePolicy(demand, request, capacity int64) int {
	res := 0
	if demand > capacity && request <= capacity {
		res = 1
	} else if demand > capacity && request > capacity {
		res = 2
	}
	return res
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func allocate(clock clock.Clock, pods []*pod.Pod, capacity, demand, request *nodeinfo.Resource) float32 {
	cpuFairSharePolicy := whichSharePolicy(demand.MilliCPU, request.MilliCPU, capacity.MilliCPU)
	memFairSharePolicy := whichSharePolicy(demand.Memory, request.Memory, capacity.Memory)
	runningPods := float32(0.0)
	for _, pod := range pods {
		if !pod.IsTerminated(clock) {
			runningPods++
		}
	}
	if runningPods == 0 {
		return 0.0
	}
	numSatifisedPods := float32(0.0)
	c := float32(capacity.MilliCPU)
	d := float32(demand.MilliCPU)
	r := float32(request.MilliCPU)
	// cpu
	if cpuFairSharePolicy > 0 {
		// guarantee
		fairShare := float32(c) / float32(d)
		for _, pod := range pods {
			if !pod.IsTerminated(clock) {
				pRequest := nodeinfo.NewResource(pod.CurrentMetrics.ResourceRequest)
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				guarantee := min(int64(fairShare*float32(pRequest.Memory)), pUsage.Memory)
				c -= float32(guarantee)
				d -= float32(guarantee)
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				pAllocation.MilliCPU = int64(guarantee)
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}

		fairShare = float32(c) / float32(d)
		for _, pod := range pods {
			if !pod.IsTerminated(clock) {
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				pRequest := nodeinfo.NewResource(pod.CurrentMetrics.ResourceRequest)

				extra := pUsage.MilliCPU - pAllocation.MilliCPU
				if pUsage.MilliCPU <= pAllocation.MilliCPU || pRequest.MilliCPU <= pAllocation.MilliCPU {
					// &&	(pUsage.Memory <= pAllocation.Memory || pRequest.Memory <= pAllocation.Memory)
					numSatifisedPods++
				}
				pAllocation.MilliCPU += int64(fairShare * float32(extra))
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
				// fmt.Printf("ResourceAllocation: %v \n", pod.CurrentMetrics.ResourceAllocation)
				// fmt.Printf("pAllocation: %v \n", pAllocation)
			}
		}
	} else {
		numSatifisedPods = runningPods
	}

	c = float32(capacity.MilliCPU)
	d = float32(demand.MilliCPU)
	r = float32(request.MilliCPU)
	// memory
	if memFairSharePolicy > 0 {
		fairShare := float32(c) / float32(r)
		// guarantee
		for _, pod := range pods {
			if !pod.IsTerminated(clock) {
				pRequest := nodeinfo.NewResource(pod.CurrentMetrics.ResourceRequest)
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				guarantee := min(int64(fairShare*float32(pRequest.Memory)), pUsage.Memory)
				c -= float32(guarantee)
				d -= float32(guarantee)
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				pAllocation.Memory = guarantee
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}
		// share remaining
		fairShare = float32(c) / float32(d)
		for _, pod := range pods {
			if !pod.IsTerminated(clock) {
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				extra := pUsage.Memory - pAllocation.Memory
				pAllocation.Memory += int64(fairShare * float32(extra))
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}
	}
	return numSatifisedPods
}

// BuildMetrics builds a Metrics at the given clock.
func BuildMetrics(clock clock.Clock, nodes map[string]*node.Node, queue queue.PodQueue, predictionPenalty float32) (Metrics, error) {
	isTinyMetrics := false
	metrics := make(map[string]interface{})
	metrics[ClockKey] = clock.ToRFC3339()

	nodesMetrics := make(map[string]node.Metrics)
	podsMetrics := make(map[string]pod.Metrics)
	QualityOfService := float32(1.0)
	numPods := float32(0.0)
	numSatifisedPods := float32(0.0)
	for name, node := range nodes {
		nodesMetrics[name] = node.Metrics(clock)
		if !isTinyMetrics {
			capacity := nodeinfo.NewResource(nodesMetrics[name].Allocatable)
			demand := nodeinfo.NewResource(nodesMetrics[name].TotalResourceUsage)
			request := nodeinfo.NewResource(nodesMetrics[name].TotalResourceRequest)
			for _, pod := range node.PodList() {
				if !pod.IsTerminated(clock) {
					key, err := util.PodKey(pod.ToV1())
					if err != nil {
						return Metrics{}, err
					}
					podsMetrics[key] = pod.Metrics(clock)
				}
			}
			numSatifisedPods += allocate(clock, node.PodList(), capacity, demand, request)

			for _, pod := range node.PodList() {
				if !pod.IsTerminated(clock) {
					key, err := util.PodKey(pod.ToV1())
					if err != nil {
						return Metrics{}, err
					}
					podsMetrics[key] = *pod.CurrentMetrics
					numPods++
				}
			}
		}
	}
	if numPods == 0 {
		QualityOfService = 1.0
	} else {
		QualityOfService = numSatifisedPods / numPods
	}

	metrics[NodesMetricsKey] = nodesMetrics
	metrics[PodsMetricsKey] = make(map[string]pod.Metrics) //podsMetrics
	metrics[QueueMetricsKey] = queue.Metrics(QualityOfService, predictionPenalty)

	return metrics, nil
}

// Formatter defines the interface of metrics formatter.
type Formatter interface {
	// Format formats the given metrics to a string.
	Format(metrics *Metrics) (string, error)
}

// Writer defines the interface of metrics writer.
type Writer interface {
	// Write writes the given metrics to some location(s).
	Write(metrics *Metrics) error
}
