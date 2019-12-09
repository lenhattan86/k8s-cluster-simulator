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
	"errors"
	"fmt"

	"github.com/containerd/containerd/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"
	"k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	kutil "k8s.io/kubernetes/pkg/scheduler/util"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/metrics"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/node"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// OnlineListScheduler
type OnlineListScheduler struct {
	extenders         []Extender
	predicates        map[string]predicates.FitPredicate
	penaltyMap        map[string]float32
	penaltyTiming     map[string]int
	prioritizers      []priorities.PriorityConfig
	prevPredictions   []*NodeMetrics
	predictionPenalty float32
	penaltyTimeout    int

	lastNodeIndex     uint64
	preemptionEnabled bool
}

// NewOnlineListScheduler creates a new OnlineListScheduler.
func NewOnlineListtScheduler(preeptionEnabled bool, penalty float32, timeOut int) OnlineListScheduler {
	return OnlineListScheduler{
		predicates:        map[string]predicates.FitPredicate{},
		preemptionEnabled: preeptionEnabled,
		penaltyMap:        make(map[string]float32),
		penaltyTiming:     make(map[string]int),
		penaltyTimeout:    timeOut,
		predictionPenalty: penalty,
	}
}

// AddExtender adds an extender to this OnlineListScheduler.
func (sched *OnlineListScheduler) AddExtender(extender Extender) {
	sched.extenders = append(sched.extenders, extender)
}

// AddPredicate adds a predicate plugin to this OnlineListScheduler.
func (sched *OnlineListScheduler) AddPredicate(name string, predicate predicates.FitPredicate) {
	sched.predicates[name] = predicate
}

// AddPrioritizer adds a prioritizer plugin to this OnlineListScheduler.
func (sched *OnlineListScheduler) AddPrioritizer(prioritizer priorities.PriorityConfig) {
	sched.prioritizers = append(sched.prioritizers, prioritizer)
}

// Schedule implements Scheduler interface.
// Schedules pods in one-by-one manner by using registered extenders and plugins.
func (sched *OnlineListScheduler) Schedule(
	clock clock.Clock,
	pendingPods queue.PodQueue,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error) {

	results := []Event{}

	for {
		// For each pod popped from the front of the queue, ...
		pod, err := pendingPods.Front() // not pop a pod here; it may fail to any node
		if err != nil {
			if err == queue.ErrEmptyQueue {
				break
			} else {
				return []Event{}, errors.New("Unexpected error raised by Queueu.Pop()")
			}
		}

		log.L.Tracef("Trying to schedule pod %v", pod)

		podKey, err := util.PodKey(pod)
		if err != nil {
			return []Event{}, err
		}
		log.L.Debugf("Trying to schedule pod %s", podKey)
		// ... try to bind the pod to a node.
		nodeMetricsArray := sched.estimate(nodeInfoMap)
		result, err := sched.scheduleOne(pod, nodeLister, nodeInfoMap, nodeMetricsArray)

		if err != nil {
			updatePodStatusSchedulingFailure(clock, pod, err)

			// If failed to select a node that can accommodate the pod, ...
			if _, ok := err.(*core.FitError); ok {
				log.L.Tracef("Pod %v does not fit in any node", pod)
				log.L.Debugf("Pod %s does not fit in any node", podKey)

				// Else, stop the scheduling process at this clock.
				break
			} else {
				return []Event{}, nil
			}
		}

		// If found a node that can accommodate the pod, ...
		log.L.Debugf("Selected node %s", result.SuggestedHost)

		pod, _ = pendingPods.Pop()
		updatePodStatusSchedulingSucceess(clock, pod)
		if err := pendingPods.RemoveNominatedNode(pod); err != nil {
			return []Event{}, err
		}

		nodeInfo, ok := nodeInfoMap[result.SuggestedHost]
		if !ok {
			return []Event{}, fmt.Errorf("No node named %s", result.SuggestedHost)
		}
		nodeInfo.AddPod(pod)

		// ... then bind it to the node.
		results = append(results, &BindEvent{Pod: pod, ScheduleResult: result})
	}

	return results, nil
}

var _ = Scheduler(&OnlineListScheduler{})

func (sched *OnlineListScheduler) estimate(nodeInfoMap map[string]*nodeinfo.NodeInfo) []*NodeMetrics {
	// monitorMap := sched.monitor(nodeInfoMap)
	nodeMetricsArray := make([]*NodeMetrics, 0, len(nodeInfoMap))
	// predict.
	for nodeName := range nodeInfoMap {
		nodeMetrics := &NodeMetrics{
			Name:        nodeName,
			Usage:       GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].TotalResourceUsage,
			Allocatable: GlobalMetrics[metrics.NodesMetricsKey].(map[string]node.Metrics)[nodeName].Allocatable,
		}
		nodeMetricsArray = append(nodeMetricsArray, nodeMetrics)
		if _, ok := sched.penaltyMap[nodeName]; !ok {
			sched.penaltyMap[nodeName] = sched.predictionPenalty
		}
	}

	// react to prediction errors.
	if sched.prevPredictions != nil {
		for i, p := range sched.prevPredictions {
			m := nodeMetricsArray[i]
			if !util.ResourceListGE(p.Usage, m.Usage) {
				sched.penaltyMap[m.Name] = sched.penaltyMap[m.Name] * sched.predictionPenalty
				sched.penaltyTiming[m.Name] = 0
			} else if util.ResourceListGE(p.Usage, m.Usage) {
				sched.penaltyTiming[m.Name]++
				if sched.penaltyTiming[m.Name] >= sched.penaltyTimeout {
					sched.penaltyMap[m.Name] = sched.predictionPenalty
				}
			} else {
				sched.penaltyTiming[m.Name] = 0
			}
			nodeMetricsArray[i].Usage = util.ResourceListMultiply(m.Usage, sched.penaltyMap[m.Name])
		}
	}

	sched.prevPredictions = nodeMetricsArray

	return nodeMetricsArray
}

func (sched *OnlineListScheduler) monitor(nodeInfoMap map[string]*nodeinfo.NodeInfo) map[string]*NodeMetrics {
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

// scheduleOne makes scheduling decision for the given pod and nodes.
// Returns core.ErrNoNodesAvailable if nodeLister lists zero nodes, or core.FitError if the given
// pod does not fit in any nodes.
func (sched *OnlineListScheduler) scheduleOne(
	pod *v1.Pod,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo,
	nodeMetricsArray []*NodeMetrics) (core.ScheduleResult, error) {

	result := core.ScheduleResult{}
	nodes, err := nodeLister.List()

	if err != nil {
		return result, err
	}

	if len(nodes) == 0 {
		return result, core.ErrNoNodesAvailable
	}

	// init min value
	min := kutil.GetResourceRequest(pod)
	min.Add(nodeMetricsArray[0].Usage)
	host := nodeMetricsArray[0].Name
	cap := nodeinfo.NewResource(nodeMetricsArray[0].Allocatable)

	// search for min
	for _, n := range nodeMetricsArray {
		temp := kutil.GetResourceRequest(pod)
		temp.Add(n.Usage)
		// log.L.Infof("min %v temp %v", min, temp)
		if temp.MilliCPU < min.MilliCPU && temp.Memory < min.Memory {
			min = temp
			host = n.Name
			cap = nodeinfo.NewResource(n.Allocatable)
		}
	}

	if min.MilliCPU <= cap.MilliCPU && min.Memory <= cap.Memory {
		result = core.ScheduleResult{
			SuggestedHost:  host,
			EvaluatedNodes: len(nodes),
			FeasibleNodes:  1,
		}
	}

	return result, err
}

// selectHost takes a prioritized list of nodes and then picks one
// in a round-robin manner from the nodes that had the highest score.
func (sched *OnlineListScheduler) selectHost(priorities api.HostPriorityList) (string, error) {
	if len(priorities) == 0 {
		return "", errors.New("Empty priorities")
	}

	maxScores := findMaxScores(priorities)
	// idx := int(sched.lastNodeIndex % uint64(len(maxScores)))
	// sched.lastNodeIndex++

	// return priorities[maxScores[idx]].Host, nil
	// TanLe: Fix the issue for best-fit: do not allow round-robin
	idx := len(maxScores) - 1
	return priorities[maxScores[idx]].Host, nil
}
