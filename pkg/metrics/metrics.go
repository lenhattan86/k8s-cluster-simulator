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
	"context"
	"sync"
	"sync/atomic"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
	"k8s.io/client-go/util/workqueue"
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
	res := 0 // allocaton = demand. (demand <= capacity.)
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

func allocate(clock clock.Clock, pods []*pod.Pod, capacity, demand, request *nodeinfo.Resource) (int32, int32) {

	cpuFairSharePolicy := whichSharePolicy(demand.MilliCPU, request.MilliCPU, capacity.MilliCPU)
	memFairSharePolicy := whichSharePolicy(demand.Memory, request.Memory, capacity.Memory)

	runingPods := make([]*pod.Pod, 0, len(pods))
	for _, pod := range pods {
		if !pod.IsTerminated(clock) {
			runingPods = append(runingPods, pod)
		}
	}
	numRunningPods := int32(len(runingPods))
	if numRunningPods == 0 {
		return 0, 0
	}
	numSatifisedPods := int32(0)
	c := capacity.MilliCPU
	d := demand.MilliCPU
	r := request.MilliCPU
	// cpu
	if cpuFairSharePolicy > 0 {
		// guarantee
		// fairShare := float32(c) / float32(d)
		if d > 0 {
			C := c
			R := r
			for _, pod := range runingPods {
				pRequest := nodeinfo.NewResource(pod.CurrentMetrics.ResourceRequest)
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				guarantee := min(C*pRequest.MilliCPU/R, pUsage.MilliCPU)
				c -= guarantee
				d -= guarantee
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				pAllocation.MilliCPU = int64(guarantee)
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}

		// fairShare = float32(c) / float32(d)
		if d > 0 {
			C := c
			D := d
			for _, pod := range runingPods {
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				extra := max(pUsage.MilliCPU-pAllocation.MilliCPU, 0)
				pAllocation.MilliCPU += int64(C * extra / D)
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}
	}

	c = capacity.Memory
	d = demand.Memory
	r = request.Memory
	// memory
	if memFairSharePolicy > 0 {
		// guarantee
		// fairShare := float32(c) / float32(d)
		if d > 0 {
			C := c
			R := r
			for _, pod := range runingPods {
				pRequest := nodeinfo.NewResource(pod.CurrentMetrics.ResourceRequest)
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				guarantee := min(C*pRequest.Memory/R, pUsage.Memory)
				c -= guarantee
				d -= guarantee
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
				pAllocation.Memory = int64(guarantee)
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}

		// fairShare = float32(c) / float32(d)
		if d > 0 {
			C := c
			D := d
			for _, pod := range runingPods {
				pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
				pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)

				extra := max(pUsage.Memory-pAllocation.Memory, 0)
				pAllocation.Memory += int64(C * extra / D)
				pod.CurrentMetrics.ResourceAllocation = pAllocation.ResourceList()
			}
		}
	}

	for _, pod := range runingPods {
		pUsage := nodeinfo.NewResource(pod.CurrentMetrics.ResourceUsage)
		pAllocation := nodeinfo.NewResource(pod.CurrentMetrics.ResourceAllocation)
		pRequest := nodeinfo.NewResource(pod.CurrentMetrics.ResourceRequest)
		if pUsage.MilliCPU <= pAllocation.MilliCPU || pRequest.MilliCPU <= pAllocation.MilliCPU {
			numSatifisedPods++
		}
	}

	return numSatifisedPods, numRunningPods
}

// BuildMetrics builds a Metrics at the given clock.
const parralel = true
const workerNum = 16

func BuildMetrics(clock clock.Clock, nodes map[string]*node.Node, queue queue.PodQueue, predictionPenalty float32) (Metrics, error) {
	isTinyMetrics := false
	metrics := make(map[string]interface{})
	metrics[ClockKey] = clock.ToRFC3339()

	nodesMetrics := make(map[string]node.Metrics)
	podsMetrics := make(map[string]pod.Metrics)
	QualityOfService := float32(1.0)
	numPods := int32(0)
	numSatifisedPods := int32(0)

	if parralel {
		var nodesMetricsMutex = sync.RWMutex{}
		var podsMetricsMutex = sync.RWMutex{}
		nodeNames := make([]string, 0, len(nodes))
		for k := range nodes {
			nodeNames = append(nodeNames, k)
		}
		ctx, _ := context.WithCancel(context.Background())
		workqueue.ParallelizeUntil(ctx, workerNum, len(nodes), func(i int) {
			name := nodeNames[i]
			node := nodes[name]
			nodeMetrics := node.Metrics(clock)
			nodesMetricsMutex.Lock()
			nodesMetrics[name] = nodeMetrics
			nodesMetricsMutex.Unlock()
			if !isTinyMetrics {
				capacity := nodeinfo.NewResource(nodeMetrics.Allocatable)
				demand := nodeinfo.NewResource(nodeMetrics.TotalResourceUsage)
				request := nodeinfo.NewResource(nodeMetrics.TotalResourceRequest)
				for _, pod := range node.PodList() {
					if !pod.IsTerminated(clock) {
						key, err := util.PodKey(pod.ToV1())
						if err != nil {
							return
						}
						podMetrics := pod.Metrics(clock)
						podsMetricsMutex.Lock()
						podsMetrics[key] = podMetrics
						podsMetricsMutex.Unlock()
					}
				}
				delta1, delta2 := allocate(clock, node.PodList(), capacity, demand, request)
				atomic.AddInt32(&numSatifisedPods, delta1)
				atomic.AddInt32(&numPods, delta2)
			}
		})
	} else {
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
				delta1, delta2 := allocate(clock, node.PodList(), capacity, demand, request)
				numSatifisedPods += delta1
				numPods += delta2

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
	}
	if numPods == 0 {
		QualityOfService = 1.0
	} else {
		QualityOfService = float32(numSatifisedPods) / float32(numPods)
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
