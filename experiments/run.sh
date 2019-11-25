echo "================== RUNNING=================="
date

BEST_FIT="bestfit"
OVER_SUB="oversub"
PROPOSED="proposed"
ONE_SHOT="oneshot"
WORST_FIT="worstfit"
GENERIC="generic"

oversub=1.5
nodeNum=5000
cpuPerNode=64
memPerNode=128

maxTaskLengthSeconds=7200 # seconds.
totalPodNumber=600000
workloadSubsetFactor=1
isDebug=true
workloadSubfolderCap=100000
path="/ssd/projects/google-trace-data"
log_path="/ssd/projects/google-trace-data"
tick=60
metricsTick=60
# path="./gen/"
# log_path="./gen/"
# tick=1
# metricsTick=1
runSim(){
    start="2019-01-01T00:00:00+09:00"
    end="2019-01-01T05:00:00+09:00"
    startTrace="600000000"
    ./gen_config.sh $1 "." $nodeNum $cpuPerNode $memPerNode $tick $metricsTick "$start" $log_path
    go run $(go list ./...) --config="./config/cluster_$1" \
    --workload="$path/workload"  \
    --scheduler="$1" \
    --isgen=$2 \
    --oversub=$oversub \
    --istrace=$3 \
    --trace="$path/tasks" \
    --start="$start" \
    --end="$end" \
    --trace-start="$startTrace" \
    --tick="$tick" \
    --max-task-length="$maxTaskLengthSeconds" \
    --total-pods-num=$totalPodNumber \
    --subset-factor=$workloadSubsetFactor \
    --workload-subfolder-cap=$workloadSubfolderCap \
    &> run_${1}.out
}
#rm -rf *.out
SECONDS=0
runSim $GENERIC true true
echo "Generating workload took $SECONDS seconds"

SECONDS=0 
# runSim $WORST_FIT false false &
# runSim $OVER_SUB false false &
# runSim $ONE_SHOT false false &
wait
echo "others took $SECONDS seconds"

SECONDS=0 
echo "==================Plotting=================="
# python plotResults.py
echo "plotResults.py took $SECONDS seconds"
echo "==================FINISHED=================="
date