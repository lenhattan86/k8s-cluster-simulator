import os
import matplotlib as mpl
if os.environ.get('DISPLAY','') == '':
    print('no display found. Using non-interactive Agg backend')
    mpl.use('Agg')

import sys
import json
import re
import matplotlib.pyplot as plt

sys.path.insert(0, './include')
from common import *
from utils import *
from data_utils import *

tick = 1
## plot utilization: number of busy nodes.
cap = 64

cpuStr = 'cpu'
show=False
plotObj = True
plotOverload = True
plotRequest = True
plotOverbook = True
loads = [False, False, False, plotOverload, plotOverbook, plotObj, plotRequest]

path = "./log"
arg_len = len(sys.argv) - 1
if arg_len > 0:
    path=sys.argv[1]

# path = "./"
line_num = 60*24
def loadLog(filepath) :
    cpuUsages = []
    maxCpuUsages = []
    cpuRequests = []
    memUsages = []
    gpuUsages = []
    cpuAllocatables = []
    requests = []
    busyNodes = []
    overloadNodes = []
    overBookNodes = []

    with open(filepath) as fp:
        line = fp.readline()
        # content = fp.readlines()
        i = 0
        while line:
        # for line in content:ot
            busyNode = 0
            overloadNode = 0
            overBookNode = 0
            totalCpuUsage = 0
            totalCapacity = 0
            maxCpuUsage = 0
            totalCpuRequest = 0

            try:
                data = json.loads(line)
            except:
                print("An json.loads(line) exception occurred") 
                continue           

            nodeDict = data['Nodes']
            for nodeName, node in nodeDict.items():
                cpuUsage = 0
                cpuAllocatable = 0
                cpuRequest = 0                
                runningPodsNum = int(node['RunningPodsNum'])

                usageDict = node['TotalResourceUsage']
                for rsName in usageDict:
                    if(rsName==cpuStr):
                        cpuUsage = formatQuatity(usageDict[rsName])
                        totalCpuUsage = totalCpuUsage+ cpuUsage
                        if cpuUsage > maxCpuUsage:
                            maxCpuUsage = cpuUsage

                allocatableDict = node['Allocatable']    
                for rsName in allocatableDict:
                    if(rsName==cpuStr):
                        cpuAllocatable = formatQuatity(allocatableDict[rsName])
                        totalCapacity = totalCapacity + cpuAllocatable
                
                requestDict = node['TotalResourceRequest']    
                for rsName in requestDict:
                    if(rsName==cpuStr):
                        cpuRequest = formatQuatity(requestDict[rsName])
                        totalCpuRequest = totalCpuRequest + cpuRequest

                if(cpuUsage > cpuAllocatable):
                    overloadNode = overloadNode+1
           
                if(cpuRequest > cpuAllocatable):
                    overBookNode = overBookNode +1
           
                if(runningPodsNum > 0):
                    busyNode = busyNode + 1

            if (loads[0]):
                cpuUsages.append(totalCpuUsage)
            if (loads[1]):
                cpuAllocatables.append(totalCapacity)
            if (loads[2]):
                busyNodes.append(busyNode)
            if (loads[3]):
                overloadNodes.append(overloadNode) 
            if (loads[4]):
                overBookNodes.append(overBookNode)
            if (loads[5]):
                maxCpuUsages.append(maxCpuUsage)
            if (loads[6]):
                cpuRequests.append(totalCpuRequest)

            i=i+1            
            if line_num > 0 and i >= line_num:
                break
            line = fp.readline()

    fp.close()

    return busyNodes, overloadNodes, overBookNodes, cpuUsages, cpuRequests, maxCpuUsages, cpuAllocatables

def formatQuatity(str):
    strArray = re.split('(\d+)', str)
    val = float(strArray[1])
    scaleStr = strArray[2]
    if scaleStr != "":
        if(scaleStr == "m"):
            val = val/1000        
        elif (scaleStr == "Mi"):
            val = val/1024

    return val

methods = ["oneshot","worstfit","oversub"]
# methods = ["oneshot","worstfit"]
methodsNum = len(methods)
busyNodes = []
overloadNodes = []
overbookNodes = []
cpuUsages = []
maxCpuUsages = []
cpuAllocatables = []
cpuRequests = []

for m in methods:
    b, ol, ob, u, ur, mu, a = loadLog(path+"/kubesim_"+m+".log")
    busyNodes.append(b)
    overloadNodes.append(ol)
    overbookNodes.append(ob)
    cpuUsages.append(u)
    maxCpuUsages.append(mu)
    cpuAllocatables.append(a)
    cpuRequests.append(ur)

############# PLOTTING ##############
if not os.path.exists(FIG_PATH):
    os.makedirs(FIG_PATH)

if plotObj:
    # Y_MAX = cap*1.5
    fig = plt.figure(figsize=FIG_ONE_COL)
    max_len = 0
    for i in range(methodsNum):
        plt.plot(range(0,len(maxCpuUsages[i])*tick,tick), maxCpuUsages[i])
        if max_len < len(maxCpuUsages[i]):
            max_len = len(maxCpuUsages[i])
    
    plt.plot(range(0,max_len*tick,tick), [cap] * max_len)
    legends = methods
    legends.append('capacity')
    plt.legend(legends, loc='best')
    plt.xlabel('time (minutes)')
    plt.ylabel(STR_CPU_CORES)
    plt.suptitle("Max Cpu Usage")
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/max_cpu_usage.pdf", bbox_inches='tight')

if plotRequest:
    # Y_MAX = np.amax(cpuRequests)
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuRequests[i])*tick,tick), cpuRequests[i])
    
    plt.legend(methods, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_CPU_CORES)
    # plt.ylim(0,Y_MAX)
    plt.suptitle("Total Cpu Request")

    fig.savefig(FIG_PATH+"/request.pdf", bbox_inches='tight')

## plot performance: number of overload nodes.
if plotOverload:
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(overloadNodes[i])*tick,tick), overloadNodes[i])

    plt.legend(methods, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_NODES)
    plt.suptitle("Overload")
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/overload.pdf", bbox_inches='tight')

## plot performance: number of overload nodes.
if plotOverbook:
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(overbookNodes[i])*tick,tick), overbookNodes[i])

    plt.legend(methods, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_NODES)
    plt.suptitle("Overbook")
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/overbook.pdf", bbox_inches='tight')
    
## show figures
if show:
    plt.show()