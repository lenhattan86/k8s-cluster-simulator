###
import sys

sys.path.insert(0, './include')
from common import *
from utils import *
from data_utils import *

### paramters
num_files = 10
path = '/Users/tanle/projects/google-trace-analysis/results.bk/machines/'
config_path = './config/'
config_file = "cluster_test.yaml"

scheduler="proposed"
log_file="kubesim_"+scheduler+".log"
filePath=config_path+"/cluster_"+scheduler+".yaml"
node_num=num_files
cpu=64
mem=128
tick=1
metricsTick=1
start="2019-01-01T00:00:00+09:00"
log_path="./log"

## 1. visit every file and read its data.
files = list_files(path, ".csv")
print("list " + str(len(files)) +" files")
if num_files < 0:
    num_files = len(files)
elif num_files > len(files):
    num_files = len(files)

cpuUsages = []
cpuReqs = []
memUsages = []
memReqs = []

cpuMax=64
memMax=128
cpuCaps = []
memCaps = []

print("reading " + str(num_files) +" files -- start")

i_file = 0
     
header="""# Log level defined by sirupsen/logrus.
# Optional (info, debug)
logLevel: info

# Interval duration for scheduling and updating the cluster, in seconds.
# Optional (default: 10)
tick: """+str(tick) +"""

# Start time at which the simulation starts, in RFC3339 format.
# Optional (default: now)
startClock: """+start+"""

# Interval duration for logging metrics of the cluster, in seconds.
# Optional (default: same as tick)
metricsTick: """+str(metricsTick) +"""

# Metrics of simulated kubernetes cluster is written
# to standard out, standard error or files at given paths.
# The metrics is formatted with the given formatter.
# Optional (default: not writing metrics)
metricsLogger:
- dest: """+log_path +"""/"""+log_file+"""
  formatter: JSON
# - dest: """+ str(log_path) +"""/kubesim-hr-""" + str(scheduler)+""".log
#   formatter: humanReadable
# - dest: stdout
#   formatter: table

# Write configuration of each node.
cluster:
"""

servers=""
for f in files:
    if i_file >= num_files:
        break
    i_file = i_file + 1

    res = read_machine_cap_csv(f)
    servers = servers + """
- metadata:
    name: node-"""+str(i_file)+"""
    labels:
      beta.kubernetes.io/os: simulated
    annotations:
      foo: bar
  spec:
    unschedulable: false
    # taints:
    # - key: k
    #   value: v
    #   effect: NoSchedule
  status:
    allocatable:
      cpu: """+str(int(cpuMax*res[0]))+"""
      memory: """+str(int(memMax*res[1]))+"""Gi
      nvidia.com/gpu: 0
      pods: 99999
"""
    servers = servers + "\n"

config=header + servers

if not os.path.exists(config_path):
    os.makedirs(config_path)

f=open(config_path+"/"+config_file,"w+")
f.write(config)
f.close()