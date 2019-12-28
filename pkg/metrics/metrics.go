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
	"fmt"
	"sync"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/pod"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
	v1 "k8s.io/api/core/v1"
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

func minFloat32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func allocate(clock clock.Clock, runingPodKeys []string, capacity, demand, request *nodeinfo.Resource, podsMetrics map[string]pod.Metrics, podsMetricsMutex *sync.RWMutex) (float32, float32) {
	cpuFairSharePolicy := whichSharePolicy(demand.MilliCPU, request.MilliCPU, capacity.MilliCPU)
	memFairSharePolicy := whichSharePolicy(demand.Memory, request.Memory, capacity.Memory)

	numRunningPods := float32(len(runingPodKeys))
	if numRunningPods == 0 {
		return 0, 0
	}
	qos := float32(0)
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
			for _, key := range runingPodKeys {
				podsMetricsMutex.Lock()
				currMetrics := podsMetrics[key]
				podsMetricsMutex.Unlock()

				pRequest := nodeinfo.NewResource(currMetrics.ResourceRequest)
				pUsage := nodeinfo.NewResource(currMetrics.ResourceUsage)
				guarantee := int64(0)

				if R > 0 {
					guarantee = min(C*pRequest.MilliCPU/R, pRequest.MilliCPU)
					guarantee = min(pUsage.MilliCPU, guarantee)
				}
				c -= guarantee
				d -= guarantee
				pAllocation := nodeinfo.NewResource(currMetrics.ResourceAllocation)
				pAllocation.MilliCPU = int64(guarantee)
				currMetrics.ResourceAllocation = pAllocation.ResourceList()
				podsMetricsMutex.Lock()
				podsMetrics[key] = currMetrics
				podsMetricsMutex.Unlock()
			}
		}

		// fairShare = float32(c) / float32(d)
		if d > 0 && c > 0 {
			C := c
			D := d
			for _, key := range runingPodKeys {
				podsMetricsMutex.Lock()
				currMetrics := podsMetrics[key]
				podsMetricsMutex.Unlock()

				pUsage := nodeinfo.NewResource(currMetrics.ResourceUsage)
				pAllocation := nodeinfo.NewResource(currMetrics.ResourceAllocation)
				extra := max(pUsage.MilliCPU-pAllocation.MilliCPU, 0)
				pAllocation.MilliCPU += min(int64(C*extra/D), extra)
				currMetrics.ResourceAllocation = pAllocation.ResourceList()

				podsMetricsMutex.Lock()
				podsMetrics[key] = currMetrics
				podsMetricsMutex.Unlock()
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
			for _, key := range runingPodKeys {
				podsMetricsMutex.Lock()
				currMetrics := podsMetrics[key]
				podsMetricsMutex.Unlock()

				pRequest := nodeinfo.NewResource(currMetrics.ResourceRequest)
				pUsage := nodeinfo.NewResource(currMetrics.ResourceUsage)
				guarantee := int64(0)
				if R > 0 {
					guarantee = min(C*pRequest.Memory/R, pRequest.Memory)
					guarantee = min(pUsage.Memory, guarantee)
				}
				c -= guarantee
				d -= guarantee
				pAllocation := nodeinfo.NewResource(currMetrics.ResourceAllocation)
				pAllocation.Memory = int64(guarantee)
				currMetrics.ResourceAllocation = pAllocation.ResourceList()

				podsMetricsMutex.Lock()
				podsMetrics[key] = currMetrics
				podsMetricsMutex.Unlock()
			}
		}

		// fairShare = float32(c) / float32(d)
		if d > 0 && c > 0 {
			C := c
			D := d
			for _, key := range runingPodKeys {
				podsMetricsMutex.Lock()
				currMetrics := podsMetrics[key]
				podsMetricsMutex.Unlock()
				pUsage := nodeinfo.NewResource(currMetrics.ResourceUsage)
				pAllocation := nodeinfo.NewResource(currMetrics.ResourceAllocation)
				extra := max(pUsage.Memory-pAllocation.Memory, 0)
				extra = min(int64(float64(C)/float64(D)*float64(extra)), extra)
				fmt.Printf("extra: %v, pUsage.Memory: %v pAllocation.Memory: %v, C: %v, D: %v \n", extra, pUsage.Memory, pAllocation.Memory, C, D)
				pAllocation.Memory += extra
				currMetrics.ResourceAllocation = pAllocation.ResourceList()

				podsMetricsMutex.Lock()
				podsMetrics[key] = currMetrics
				podsMetricsMutex.Unlock()
			}
		}
	}

	for _, key := range runingPodKeys {
		podsMetricsMutex.Lock()
		currMetrics := podsMetrics[key]
		podsMetricsMutex.Unlock()

		pUsage := nodeinfo.NewResource(currMetrics.ResourceUsage)
		pAllocation := nodeinfo.NewResource(currMetrics.ResourceAllocation)
		pRequest := nodeinfo.NewResource(currMetrics.ResourceRequest)

		// if (pUsage.MilliCPU <= pAllocation.MilliCPU) &&
		// 	(pUsage.Memory <= pAllocation.Memory) {
		// 	// guaranteed
		// 	qos += 1
		// } else if pRequest.MilliCPU < pUsage.MilliCPU || pRequest.Memory < pUsage.Memory {
		// 	// best effort
		// 	c := float32(1)
		// 	m := float32(1)
		// 	if pUsage.MilliCPU != 0 {
		// 		c = float32(pAllocation.MilliCPU) / float32(pUsage.MilliCPU)
		// 	}
		// 	if pUsage.Memory != 0 {
		// 		m = float32(pAllocation.Memory) / float32(pUsage.Memory)
		// 	}
		// 	qos += minFloat32(c, m)
		// }

		if (pUsage.MilliCPU <= pAllocation.MilliCPU || pAllocation.MilliCPU >= pRequest.MilliCPU) &&
			(pUsage.Memory <= pAllocation.Memory || pAllocation.Memory >= pRequest.Memory) {
			// guaranteed
			qos += 1
		}
	}

	return qos, numRunningPods
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
	numPods := float32(0)
	podQoses := float32(0)

	var nodesMetricsMutex = sync.RWMutex{}
	var podsMetricsMutex = sync.RWMutex{}
	var qosMutex = sync.RWMutex{}
	var podNumMutex = sync.RWMutex{}
	nodeNames := make([]string, 0, len(nodes))
	for k := range nodes {
		nodeNames = append(nodeNames, k)
	}
	ctx, _ := context.WithCancel(context.Background())
	workqueue.ParallelizeUntil(ctx, workerNum, len(nodes), func(i int) {
		name := nodeNames[i]
		node := nodes[name]
		nodeMetrics := node.Metrics(clock)
		resourceAllocation := v1.ResourceList{}
		if !isTinyMetrics {
			capacity := nodeinfo.NewResource(nodeMetrics.Allocatable)
			demand := nodeinfo.NewResource(nodeMetrics.TotalResourceUsage)
			request := nodeinfo.NewResource(nodeMetrics.TotalResourceRequest)
			pods := node.PodList()
			runingPodKeys := make([]string, 0, len(pods))
			for _, pod := range pods {
				if !pod.IsTerminated(clock) {
					key, err := util.PodKey(pod.ToV1())
					if err != nil {
						return
					}
					podMetrics := pod.Metrics(clock)
					podsMetricsMutex.Lock()
					podsMetrics[key] = podMetrics
					podsMetricsMutex.Unlock()
					runingPodKeys = append(runingPodKeys, key)
				}
			}
			qos, podNum := allocate(clock, runingPodKeys, capacity, demand, request, podsMetrics, &podsMetricsMutex)
			// compute the real resource allocated (usage)
			for _, key := range runingPodKeys {
				podsMetricsMutex.Lock()
				resourceAllocation = util.ResourceListSum(resourceAllocation, podsMetrics[key].ResourceAllocation)
				podsMetricsMutex.Unlock()
			}
			qosMutex.Lock()
			podQoses += qos
			qosMutex.Unlock()
			podNumMutex.Lock()
			numPods += podNum
			podNumMutex.Unlock()
		}
		nodeMetrics.TotalResourceAllocation = resourceAllocation
		nodesMetricsMutex.Lock()
		nodesMetrics[name] = nodeMetrics
		nodesMetricsMutex.Unlock()
	})
	if numPods == 0 {
		QualityOfService = 1.0
	} else {
		QualityOfService = float32(podQoses) / float32(numPods)
	}

	metrics[NodesMetricsKey] = nodesMetrics
	metrics[PodsMetricsKey] = make(map[string]pod.Metrics) //podsMetrics
	metrics[QueueMetricsKey] = queue.Metrics(QualityOfService, predictionPenalty, podQoses, numPods)

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
