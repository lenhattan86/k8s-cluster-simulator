echo "================== RUNNING=================="
date

isOfficial=false

BEST_FIT="bestfit"
OVER_SUB="oversub"
PROPOSED="proposed"
ONE_SHOT="oneshot"
WORST_FIT="worstfit"
GENERIC="generic"

oversub=2.2

workloadSubsetFactor=1
isDebug=true
workloadSubfolderCap=100000
startTrace="000000000"
targetNum=0
penaltyTimeout=10
predictionPenalty=1.5
targetQoS=0.99
penaltyUpdate=0.99
isDistributeTasks="false"
isMultipleResource="true"

if $isOfficial
then
    cpuPerNode=64
    memPerNode=128
    nodeNum=3000
    totalPodNumber=25000000
    start="2019-01-01T00:00:00+09:00"
    end="2019-01-02T06:00:00+09:00"
    pathToTrace="/home/cc/google-data/tasks"
    pathToWorkload="/home/cc/google-data/workload"
    log_path="/home/cc/google-data/log"
    tick=60
    metricsTick=60
else
    nodeNum=2
    totalPodNumber=40
    targetNum=40
    cpuPerNode=16
    memPerNode=16
    start="2019-01-01T00:00:00+09:00"
    end="2019-01-01T01:00:00+09:00"
    pathToTrace="/ssd/projects/google-trace-data/task-res"
    pathToWorkload="./tmp/workload"
    log_path="./log"
    tick=1
    metricsTick=1
fi

mkdir $pathToWorkload
mkdir $log_path

runSim(){
    ./gen_config.sh $1 "." $nodeNum $cpuPerNode $memPerNode $tick $metricsTick "$start" $log_path
    go run github.com/pfnet-research/k8s-cluster-simulator/experiments --config="./config/cluster_$1" \
    --workload="$pathToWorkload"  \
    --scheduler="$1" \
    --isgen=$2 \
    --oversub=$oversub \
    --istrace=$3 \
    --trace="$pathToTrace" \
    --is-distribute=$isDistributeTasks \
    --start="$start" \
    --end="$end" \
    --trace-start="$startTrace" \
    --tick="$tick" \
    --total-pods-num=$totalPodNumber \
    --target-pod-num=$targetNum \
    --subset-factor=$workloadSubsetFactor \
    --workload-subfolder-cap=$workloadSubfolderCap \
    --penalty-timeout=$penaltyTimeout \
    --prediction-penalty=$predictionPenalty \
    --target-qos=$targetQoS \
    --penalty-update=$penaltyUpdate \
    --is-multiple-resource=$isMultipleResource \
    &> run_${1}.out
}

if $isOfficial
then
    SECONDS=0
    runSim $GENERIC true true
    echo "Generating workload took $SECONDS seconds"

    SECONDS=0 
    echo "running simulation"
    runSim $PROPOSED false false &
    runSim $WORST_FIT false false &
    runSim $OVER_SUB false false &    
    wait
    echo "simulation took $SECONDS seconds"
else
    SECONDS=0
    # runSim $GENERIC true false
    echo "Generating workload took $SECONDS seconds"

    SECONDS=0 
    echo "running simulation"
    runSim $PROPOSED false false &
    runSim $WORST_FIT false false &
    runSim $OVER_SUB false false  &
    wait
    echo "simulation took $SECONDS seconds"
fi
# sudo echo -200 > /proc/$pid/oom_score_adj

SECONDS=0 
echo "==================Plotting=================="
python3 plotResults.py $log_path
echo "plotResults.py took $SECONDS seconds"
echo "==================FINISHED=================="
date