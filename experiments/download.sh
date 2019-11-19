if [ -z "$1" ]
then
	echo "usage download.sh [server]"
else
	server="$1"
fi

if [ -z "$2" ]
then
	dest="./"
else
	dest="$2"
fi

username="tanle"
log_folder="/proj/yarnrm-PG0/google-trace-data"
sim_folder="~/go/src/github.com/pfnet-research/k8s-cluster-simulator/experiments"

rm -rf $dest
mkdir $dest
scp $user@$server:$log_folder/*.log $dest/
scp $user@$server:$sim_folder/run.sh $dest/
scp $user@$server:$sim_folder/figs/*.pdf $dest/