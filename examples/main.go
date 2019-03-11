package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/priorities"

	"github.com/ordovicia/kubernetes-simulator/kubesim"
	"github.com/ordovicia/kubernetes-simulator/kubesim/queue"
	"github.com/ordovicia/kubernetes-simulator/kubesim/scheduler"
	"github.com/ordovicia/kubernetes-simulator/log"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.L.WithError(err).Fatal("Error executing root command")
	}
}

// configPath is the path of the config file, defaulting to "examples/config_sample".
var configPath string

func init() {
	rootCmd.PersistentFlags().StringVar(
		&configPath, "config", "examples/config_sample", "config file (exclusing file extension)")
}

var rootCmd = &cobra.Command{
	Use:   "kubernetes-simulator",
	Short: "kubernetes-simulator provides a virtual kubernetes cluster interface for evaluating your scheduler.",

	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		// Create a KubeSim with a queue and a scheduler.
		queue := queue.NewPriorityQueue() // queue.NewPriorityQueueWithComparator(lifo)
		sched := buildScheduler()
		kubesim, err := kubesim.NewKubeSimFromConfigPath(configPath, queue, sched)
		if err != nil {
			log.G(context.TODO()).WithError(err).Fatalf("Error creating KubeSim: %s", err.Error())
		}

		// Register a submitter
		kubesim.AddSubmitter(newMySubmitter(8))

		// SIGINT (Ctrl-C) cancels the sumbitter and kubesim.Run()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sig
			cancel()
		}()

		// Run the main loop
		if err := kubesim.Run(ctx); err != nil && errors.Cause(err) != context.Canceled {
			log.L.Fatal(err)
		}
	},
}

func buildScheduler() scheduler.Scheduler {
	sched := scheduler.NewGenericScheduler(true)

	// Add an extender
	sched.AddExtender(
		scheduler.Extender{
			Name:             "MyExtender",
			Filter:           filterExtender,
			Prioritize:       prioritizeExtender,
			Weight:           1,
			NodeCacheCapable: true,
		},
	)

	// Add plugins
	sched.AddPredicate("GeneralPredicates", predicates.GeneralPredicates)
	sched.AddPrioritizer(priorities.PriorityConfig{
		Name:   "BalancedResourceAllocation",
		Map:    priorities.BalancedResourceAllocationMap,
		Reduce: nil,
		Weight: 1,
	})
	sched.AddPrioritizer(priorities.PriorityConfig{
		Name:   "LeastRequested",
		Map:    priorities.LeastRequestedPriorityMap,
		Reduce: nil,
		Weight: 1,
	})

	return &sched
}

// for test
func lifo(pod0, pod1 *v1.Pod) bool {
	return pod1.CreationTimestamp.Before(&pod0.CreationTimestamp)
}
