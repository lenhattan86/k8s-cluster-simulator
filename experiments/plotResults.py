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
from plot_utils import *
from common import *
from utils import *
from data_utils import *

tick = 1
## plot utilization: number of busy nodes.
cap = 64
utilization_range=[0, 1]

cpuStr = 'cpu'
memStr = 'memory'
show=False
plotObj = True
plotOverload = True
plotTotalRequest = True
plotTotalDemand = True
plotOverbook = True
plotQoS=True
plotPredictionPenalty=True
plotUtilization=True
loads = [plotTotalDemand, True, False, plotOverload, plotOverbook, plotObj, plotTotalRequest, plotQoS, plotPredictionPenalty]

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
    memRequests = []
    memUsages = []
    gpuUsages = []
    cpuAllocatables = []
    memAllocatables = []
    requests = []
    busyNodes = []
    overloadNodes = []
    overBookNodes = []
    QoS = []
    PredPenalty = []

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
            totalMemUsage = 0
            totalCpuCapacity = 0
            totalMemCapacity = 0
            maxCpuUsage = 0
            totalCpuRequest = 0
            totalMemRequest = 0
            maxMemUsage = 0

            try:
                data = json.loads(line)
            except:
                print("An json.loads(line) exception occurred") 
                continue           

            nodeDict = data['Nodes']
            for nodeName, node in nodeDict.items():
                cpuUsage = 0
                memUsage = 0
                cpuAllocatable = 0
                memAllocatable = 0
                cpuRequest = 0    
                memRequest = 0            
                runningPodsNum = int(node['RunningPodsNum'])

                usageDict = node['TotalResourceUsage']
                for rsName in usageDict:
                    if(rsName==cpuStr):
                        cpuUsage = formatQuatity(usageDict[rsName])
                        totalCpuUsage = totalCpuUsage+ cpuUsage
                        if cpuUsage > maxCpuUsage:
                            maxCpuUsage = cpuUsage
                    elif(rsName==memStr):
                        memUsage = formatQuatity(usageDict[rsName])
                        totalMemUsage = totalMemUsage+ memUsage
                        if memUsage > maxMemUsage:
                            maxMemUsage = memUsage

                allocatableDict = node['Allocatable']    
                for rsName in allocatableDict:
                    if(rsName==cpuStr):
                        cpuAllocatable = formatQuatity(allocatableDict[rsName])
                        totalCpuCapacity = totalCpuCapacity + cpuAllocatable
                    elif(rsName==memStr):
                        memAllocatable = formatQuatity(allocatableDict[rsName])
                        totalMemCapacity = totalMemCapacity + memAllocatable
                
                requestDict = node['TotalResourceRequest']    
                for rsName in requestDict:
                    if(rsName==cpuStr):
                        cpuRequest = formatQuatity(requestDict[rsName])
                        totalCpuRequest = totalCpuRequest + cpuRequest
                    elif(rsName==memStr):
                        memRequest = formatQuatity(requestDict[rsName])
                        totalMemRequest = totalMemRequest + memRequest

                if(cpuUsage > cpuAllocatable or memUsage > memAllocatable):
                    overloadNode = overloadNode+1
           
                if(cpuRequest > cpuAllocatable or memRequest > memAllocatable):
                    overBookNode = overBookNode +1
           
                if(runningPodsNum > 0):
                    busyNode = busyNode + 1

            if (loads[0]):
                cpuUsages.append(totalCpuUsage)
                memUsages.append(totalMemUsage)
            if (loads[1]):
                cpuAllocatables.append(totalCpuCapacity)
                memAllocatables.append(totalMemCapacity)
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
                memRequests.append(totalMemRequest)

            # Queue":{"PendingPodsNum":1,"QualityOfService":1,"PredictionPenalty":2.97}
            queue = data['Queue']
            if (loads[7]):
                QoS.append(float(queue['QualityOfService']))
            if (loads[8]):
                PredPenalty.append(float(queue['PredictionPenalty']))

            i=i+1            
            if line_num > 0 and i >= line_num:
                break
            line = fp.readline()

    fp.close()

    return busyNodes, overloadNodes, overBookNodes, cpuUsages, memUsages, cpuRequests, memRequests, maxCpuUsages, cpuAllocatables, memAllocatables, QoS, PredPenalty

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

methods = ["proposed","worstfit","oversub"]
# methods = ["oneshot","worstfit"]
methodsNum = len(methods)
busyNodes = []
overloadNodes = []
overbookNodes = []
cpuUsages = []
memUsages = []
maxCpuUsages = []
cpuAllocatables = []
memAllocatables = []
cpuRequests = []
memRequests = []
QoSs = []
PredPenalties = []

for m in methods:
    b, ol, ob, u_cpu, u_mem, ur_cpu, ur_mem, mu, a_cpu,a_mem, q, p = loadLog(path+"/kubesim_"+m+".log")
    busyNodes.append(b)
    overloadNodes.append(ol)
    overbookNodes.append(ob)
    cpuUsages.append(u_cpu)
    memUsages.append(u_mem)
    maxCpuUsages.append(mu)
    cpuAllocatables.append(a_cpu)
    memAllocatables.append(a_mem)
    cpuRequests.append(ur_cpu)
    memRequests.append()
    QoSs.append(q)
    PredPenalties.append(p)

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

if plotUtilization:
    # requests
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)

    cpuUtil = []
    cpuMin = []
    cpuMax = []
    memUtil = []
    memMin = []
    memMax = []
    for i in range(methodsNum):
        cpu = cpuRequests[i]
        mem = memRequests[i]
        if (len(cpu) > utilization_range[1]):
            utilization_range[1] = len(cpu)
        if (len(cpu) < utilization_range[0]):
            utilization_range[0] = len(cpu)
        cpuUtil.append(np.average(cpu[utilization_range[0]:utilization_range[1]]))  
        memUtil.append(np.average(mem[utilization_range[0]:utilization_range[1]]))

    x = np.arange(methodsNum) 
    width = GBAR_WIDTH/2
    rects1 = ax.bar(x - width/2, cpuUtil,  width, label="cpu")
    rects2 = ax.bar(x + width/2, memUtil,  width, label='memory')
    
    fig.savefig(FIG_PATH+"/request-avg.pdf", bbox_inches='tight')

    # demand

    # usage

if plotTotalRequest:
    # Y_MAX = np.amax(cpuRequests)
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuRequests[i])*tick,tick), cpuRequests[i])

    plt.plot(range(0,len(cpuAllocatables[0])*tick,tick), cpuAllocatables[0])
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_CPU_CORES)
    # plt.ylim(0,Y_MAX)
    plt.suptitle("Total Cpu Request")

    fig.savefig(FIG_PATH+"/total-request.pdf", bbox_inches='tight')

if plotTotalDemand:
    # Y_MAX = np.amax(cpuRequests)
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuUsages[i])*tick,tick), cpuUsages[i])
    
    plt.plot(range(0,len(cpuAllocatables[0])*tick,tick), cpuAllocatables[0])
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_CPU_CORES)
    # plt.ylim(0,Y_MAX)
    plt.suptitle("Total Cpu Demand")

    fig.savefig(FIG_PATH+"/total-demand-cpu.pdf", bbox_inches='tight')

    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(memUsages[i])*tick,tick), memUsages[i])
    
    plt.plot(range(0,len(memAllocatables[0])*tick,tick), memAllocatables[0])
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_MEM_GB)
    # plt.ylim(0,Y_MAX)
    plt.suptitle("Total Memory Demand")

    fig.savefig(FIG_PATH+"/total-demand-mem.pdf", bbox_inches='tight')

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
    
    legends = methods   
    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_NODES)
    plt.suptitle("Overbook")
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/overbook.pdf", bbox_inches='tight')

## plot QoS
if plotQoS:

    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(QoSs[i])*tick,tick), QoSs[i])
    
    legends = methods   
    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_QoS)
    plt.ylim(0,1.1)

    fig.savefig(FIG_PATH+"/qos.pdf", bbox_inches='tight')
    ##    
    
    ## plot figures   
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        cdf = compute_cdf(numpy.array(QoSs[i]))
        plt.plot(cdf.x, cdf.p)

    legends = methods   
    plt.legend(legends, loc='best')
    plt.xlabel(STR_QoS)
    plt.ylabel(STR_CDF)
    plt.ylim(0,1)
    plt.xlim(0,1)

    fig.savefig(FIG_PATH+"/qos_cdf.pdf", bbox_inches='tight')

if plotPredictionPenalty:
    fig = plt.figure(figsize=FIG_ONE_COL)
    i=0
    plt.plot(range(0,len(PredPenalties[i])*tick,tick), PredPenalties[i])
    
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_Pred_Penalty)
    plt.ylim(0,3.1)

    fig.savefig(FIG_PATH+"/pred_penalty.pdf", bbox_inches='tight')

# STR_Pred_Penalty
## show figures
if show:
    plt.show()