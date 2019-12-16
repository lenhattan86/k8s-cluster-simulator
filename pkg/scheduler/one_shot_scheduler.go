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
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	kutil "k8s.io/kubernetes/pkg/scheduler/util"

	"github.com/pfnet-research/k8s-cluster-simulator/pkg/clock"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/queue"
	"github.com/pfnet-research/k8s-cluster-simulator/pkg/util"
)

// OneShotScheduler makes scheduling decision for each given pod in the one-by-one manner and pick the busiest pod first.
type OneShotScheduler struct {
	extenders    []Extender
	predicates   map[string]predicates.FitPredicate
	prioritizers []priorities.PriorityConfig

	lastNodeIndex     uint64
	preemptionEnabled bool
}

// NewOneShotScheduler creates a new OneShotScheduler.
func NewOneShotScheduler(preeptionEnabled bool, penalty float32, timeOut int) OneShotScheduler {
	penaltyMap = make(map[string]float32)
	penaltyTiming = make(map[string]int)
	penaltyTimeout = timeOut
	predictionPenalty = penalty
	return OneShotScheduler{
		predicates:        map[string]predicates.FitPredicate{},
		preemptionEnabled: preeptionEnabled,
	}
}

// AddExtender adds an extender to this OneShotScheduler.
func (sched *OneShotScheduler) AddExtender(extender Extender) {
	sched.extenders = append(sched.extenders, extender)
}

// AddPredicate adds a predicate plugin to this OneShotScheduler.
func (sched *OneShotScheduler) AddPredicate(name string, predicate predicates.FitPredicate) {
	sched.predicates[name] = predicate
}

// AddPrioritizer adds a prioritizer plugin to this OneShotScheduler.
func (sched *OneShotScheduler) AddPrioritizer(prioritizer priorities.PriorityConfig) {
	sched.prioritizers = append(sched.prioritizers, prioritizer)
}

// GetPodList makes a copy of pod array from PodQueue.
func GetPodList(pendingPods queue.PodQueue) []*v1.Pod {
	pods := make([]*v1.Pod, 0, 0)
	for {
		pod, err := pendingPods.Pop()
		if err != nil {
			break
		}
		pods = append(pods, pod)
	}

	for _, pod := range pods {
		pendingPods.Push(pod)
	}

	return pods
}

// Schedule implements Scheduler interface.
// Schedules pods in one-by-one manner by using registered extenders and plugins.
func (sched *OneShotScheduler) Schedule(
	clock clock.Clock,
	pendingPods queue.PodQueue,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) ([]Event, error) {

	// compute pods vs. nodes.
	pods := GetPodList(pendingPods)
	podNum := len(pods)
	if podNum == 0 {
		return []Event{}, nil
	}
	// do the mapping.

	scheduleMap, _ := sched.scheduleAll(pods, nodeLister, nodeInfoMap)
	results := []Event{}
	for i := 0; i < podNum; i++ {
		// For each pod popped from the front of the queue, ...
		pod, err := pendingPods.Pop()
		if err != nil {
			if err == queue.ErrEmptyQueue {
				break
			} else {
				return []Event{}, errors.New("Unexpected error raised by Queue.Pop()")
			}
		}

		log.L.Tracef("Trying to schedule pod %v", pod)

		podKey, err := util.PodKey(pod)
		if err != nil {
			return []Event{}, err
		}
		log.L.Debugf("Trying to schedule pod %s", podKey)

		// ... try to bind the pod to a node.
		result, ok := scheduleMap[pod.Name]

		if !ok {
			updatePodStatusSchedulingFailure(clock, pod, fmt.Errorf("current load is too high"))
			pendingPods.Push(pod)
			continue
		}

		// If found a node that can accommodate the pod, ...
		log.L.Debugf("Selected node %s", result.SuggestedHost)

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

var _ = Scheduler(&OneShotScheduler{})

// LowerResourceAvailableNode returns less available resource
func LowerResourceAvailableNode(nodeMetrics1, nodeMetrics2 interface{}) bool {
	r1 := util.ResourceListSub(nodeMetrics1.(*NodeMetrics).Allocatable, nodeMetrics1.(*NodeMetrics).Usage)
	r2 := util.ResourceListSub(nodeMetrics2.(*NodeMetrics).Allocatable, nodeMetrics2.(*NodeMetrics).Usage)
	return util.ResourceListGE(r2, r1)
}

func (sched *OneShotScheduler) scheduleAll(
	pods []*v1.Pod,
	nodeLister algorithm.NodeLister,
	nodeInfoMap map[string]*nodeinfo.NodeInfo) (map[string]core.ScheduleResult, error) {
	scheduleMap := make(map[string]core.ScheduleResult)
	nodes, err := nodeLister.List()

	if err != nil {
		return scheduleMap, err
	}

	nodeNum := len(nodes)
	if nodeNum == 0 {
		return scheduleMap, core.ErrNoNodesAvailable
	}
	nodeMetricsArray := Estimate(nodeInfoMap)

	// sort pods
	sortablePods := kutil.SortableList{CompFunc: kutil.HigherResourceRequest}
	for _, p := range pods {
		sortablePods.Items = append(sortablePods.Items, p)
	}
	sortablePods.Sort()

	for _, pod := range sortablePods.Items {
		// init min value
		min := kutil.GetResourceRequest(pod.(*v1.Pod))
		min.Add(nodeMetricsArray[0].Usage)
		host := nodeMetricsArray[0].Name
		cap := nodeinfo.NewResource(nodeMetricsArray[0].Allocatable)
		idx := 0

		// search for min
		for i, n := range nodeMetricsArray {
			temp := kutil.GetResourceRequest(pod.(*v1.Pod))
			temp.Add(n.Usage)
			// log.L.Infof("min %v temp %v", min, temp)
			if temp.MilliCPU < min.MilliCPU && temp.Memory < min.Memory {
				min = temp
				host = n.Name
				idx = i
				cap = nodeinfo.NewResource(n.Allocatable)
			}
		}

		if min.MilliCPU <= cap.MilliCPU && min.Memory <= cap.Memory {
			result := core.ScheduleResult{
				SuggestedHost:  host,
				EvaluatedNodes: nodeNum,
				FeasibleNodes:  1,
			}
			scheduleMap[pod.(*v1.Pod).Name] = result
			// update resource usage
			nodeMetricsArray[idx].Usage = util.ResourceListSum(nodeMetricsArray[idx].Usage, util.PodTotalResourceRequests(pod.(*v1.Pod)))
		}
	}

	return scheduleMap, nil
}
