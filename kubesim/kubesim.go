package kubesim

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cpuguy83/strongerrors"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"

	"github.com/ordovicia/kubernetes-simulator/kubesim/clock"
	"github.com/ordovicia/kubernetes-simulator/kubesim/node"
	"github.com/ordovicia/kubernetes-simulator/log"
	"github.com/ordovicia/kubernetes-simulator/scheduler"
)

// KubeSim represents a kubernetes cluster simulator.
type KubeSim struct {
	nodes    []*node.Node
	podQueue chan v1.Pod
	tick     int

	filters []scheduler.Filter
	scorers []scheduler.Scorer
}

// NewKubeSim creates a new KubeSim with the config.
func NewKubeSim(config *Config) (*KubeSim, error) {
	log.G(context.TODO()).Debugf("Config: %+v", *config)
	if err := configure(config); err != nil {
		return nil, errors.Errorf("error configuring: %s", err.Error())
	}

	nodes := []*node.Node{}
	for _, config := range config.Cluster.Nodes {
		log.L.Debugf("NodeConfig: %+v", config)

		nodeV1, err := buildNode(config)
		if err != nil {
			return nil, errors.Errorf("error building node config: %s", err.Error())
		}

		n := node.NewNode(nodeV1)
		nodes = append(nodes, &n)
		log.L.Debugf("Node %q created", nodeV1.Name)
	}

	kubesim := KubeSim{
		nodes:    nodes,
		podQueue: make(chan v1.Pod),
		tick:     config.Tick,
		filters:  []scheduler.Filter{},
		scorers:  []scheduler.Scorer{},
	}

	return &kubesim, nil
}

// NewKubeSimFromConfigPath creates a new KubeSim with config from configPath (excluding file path).
func NewKubeSimFromConfigPath(configPath string) (*KubeSim, error) {
	config, err := readConfig(configPath)
	if err != nil {
		return nil, errors.Errorf("error reading config: %s", err.Error())
	}

	return NewKubeSim(config)
}

// RegisterFilter registers a new filter plugin to this KubeSim.
func (k *KubeSim) RegisterFilter(filter scheduler.Filter) {
	k.filters = append(k.filters, filter)
}

// RegisterScorer registers a new scorer plugin to this KubeSim.
func (k *KubeSim) RegisterScorer(scorer scheduler.Scorer) {
	k.scorers = append(k.scorers, scorer)
}

// PodQueue returns the channel of this KubeSim for submitting pods.
func (k *KubeSim) PodQueue() chan v1.Pod {
	return k.podQueue
}

// Run executes the main loop, which invokes scheduler plugins and schedules queued pods to a
// selected node.
func (k *KubeSim) Run(ctx context.Context) error {
	tick := make(chan clock.Clock)

	go func() {
		clock := clock.NewClock(time.Now())
		for {
			clock = clock.Add(time.Duration(k.tick) * time.Second)
			tick <- clock
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case clock := <-tick:
			// log.L.Debugf("Clock %s", clock.String())
			if err := k.scheduleOne(clock); err != nil {
				return err
			}
		}
	}
}

func (k *KubeSim) scheduleOne(clock clock.Clock) error {
	pod := <-k.podQueue
	_ = pod
	_ = clock
	// for _, filter := range k.filters {
	// }

	// TODO

	return nil
}

// readConfig reads and parses a config from the path (excluding file extension).
func readConfig(path string) (*Config, error) {
	viper.SetConfigName(path)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	log.G(context.TODO()).Debugf("Using config file %s", viper.ConfigFileUsed())

	var config = Config{
		Cluster:     ClusterConfig{Nodes: []NodeConfig{}},
		APIPort:     10250,
		MetricsPort: 10255,
		LogLevel:    "info",
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func configure(config *Config) error {
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		return strongerrors.InvalidArgument(errors.Errorf("%s: log level %q not supported", err.Error(), level))
	}
	logrus.SetLevel(level)

	logger := log.L
	log.L = logger

	return nil
}
