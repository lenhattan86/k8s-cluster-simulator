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
data_range=[10, 300]
target_qos = 0.99

cpuStr = 'cpu'
memStr = 'memory'
show=False
plotObj = True
plotOverload = False
plotOverbook = False
plotQoS=True
plotPredictionPenalty=True
plotUtilization=True
loads = [plotUtilization, False, plotOverload, plotOverbook, plotQoS, plotPredictionPenalty]

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
    totalCpuAllocations = []
    totalMemAllocations = []
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
            totalCpuAllocation = 0
            totalMemAllocation = 0
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
                
                allocationDict = node['TotalResourceAllocation']    
                for rsName in allocationDict:
                    if(rsName==cpuStr):
                        cpuAllocation = formatQuatity(allocationDict[rsName])
                        totalCpuAllocation = totalCpuAllocation + cpuAllocation
                    elif(rsName==memStr):
                        memAllocation = formatQuatity(allocationDict[rsName])
                        totalMemAllocation = totalMemAllocation + memAllocation

                if(cpuUsage > cpuAllocatable or memUsage > memAllocatable):
                    overloadNode = overloadNode+1
           
                if(cpuRequest > cpuAllocatable or memRequest > memAllocatable):
                    overBookNode = overBookNode +1
           
                if(runningPodsNum > 0):
                    busyNode = busyNode + 1

            if (loads[0]):
                cpuUsages.append(totalCpuUsage)
                memUsages.append(totalMemUsage)
                cpuAllocatables.append(totalCpuCapacity)
                memAllocatables.append(totalMemCapacity)
                cpuRequests.append(totalCpuRequest)
                memRequests.append(totalMemRequest)
                maxCpuUsages.append(maxCpuUsage)
                totalCpuAllocations.append(totalCpuAllocation)
                totalMemAllocations.append(totalMemAllocation)
            if (loads[1]):
                busyNodes.append(busyNode)
            if (loads[2]):
                overloadNodes.append(overloadNode) 
            if (loads[3]):
                overBookNodes.append(overBookNode)

            # Queue":{"PendingPodsNum":1,"QualityOfService":1,"PredictionPenalty":2.97}
            queue = data['Queue']
            if (loads[4]):
                QoS.append(float(queue['QualityOfService']))
            if (loads[5]):
                PredPenalty.append(float(queue['PredictionPenalty']))

            i=i+1            
            if line_num > 0 and i >= line_num:
                break
            line = fp.readline()

    fp.close()

    return busyNodes, overloadNodes, overBookNodes, cpuUsages, memUsages, cpuRequests, memRequests, totalCpuAllocations, totalMemAllocations, maxCpuUsages, cpuAllocatables, memAllocatables, QoS, PredPenalty

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
colors = [COLOR_PROPOSED, COLOR_WORST_FIT, COLOR_OVER_SUB]
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
cpuAllocations = []
memAllocations = []
cpuRequests = []
memRequests = []
QoSs = []
PredPenalties = []

for m in methods:
    b, ol, ob, u_cpu, u_mem, ur_cpu, ur_mem, a_cpu, a_mem, mu, c_cpu,c_mem, q, p = loadLog(path+"/kubesim_"+m+".log")
    busyNodes.append(b)
    overloadNodes.append(ol)
    overbookNodes.append(ob)
    cpuUsages.append(u_cpu)
    memUsages.append(u_mem)
    maxCpuUsages.append(mu)
    cpuAllocatables.append(c_cpu)
    memAllocatables.append(c_mem)
    cpuAllocations.append(a_cpu)
    memAllocations.append(a_mem)
    cpuRequests.append(ur_cpu)
    memRequests.append(ur_mem)
    QoSs.append(q)
    PredPenalties.append(p)

for i in range(methodsNum):
    if (len(cpuRequests[i]) > data_range[1]):
        data_range[1] = len(cpuRequests[i])
    if (len(cpuRequests[i]) < data_range[0]):
        data_range[0] = len(cpuRequests[i])

############# PLOTTING ##############
if not os.path.exists(FIG_PATH):
    os.makedirs(FIG_PATH)

if plotObj:
    # Y_MAX = cap*1.5
    fig = plt.figure(figsize=FIG_ONE_COL)
    max_len = 0
    for i in range(methodsNum):
        plt.plot(range(0,len(maxCpuUsages[i])*tick,tick), maxCpuUsages[i], color=colors[i])
        if max_len < len(maxCpuUsages[i]):
            max_len = len(maxCpuUsages[i])
    
    plt.plot(range(0,max_len*tick,tick), [cap] * max_len, color=COLOR_CAP)
    legends = methods
    legends.append('capacity')
    plt.legend(legends, loc='best')
    plt.xlabel('time (minutes)')
    plt.ylabel(STR_CPU_CORES)
    plt.suptitle("Max Cpu Usage")
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/max_cpu_usage.pdf", bbox_inches='tight')

if plotUtilization:
    cpuReqUtil = []
    memReqUtil = []
    cpuDemandUtil = []
    memDemandUtil = []
    cpuUsageUtil = []
    memUsageUtil = []
    cpuCap = np.average(cpuAllocatables[0])
    memCap = np.average(memAllocatables[0])

    if memCap == 0:
        memCap = 1.0
    if cpuCap == 0:
        cpuCap = 1.0

    Y_MAX = 1.5

    for i in range(methodsNum):
        cpuR = cpuRequests[i]
        memR = memRequests[i]
        cpuD = cpuUsages[i]
        memD = memUsages[i]
        cpuU = cpuAllocations[i]
        memU = memAllocations[i]
        cpuReqUtil.append(np.average(cpuR[data_range[0]:data_range[1]])/cpuCap)  
        memReqUtil.append(np.average(memR[data_range[0]:data_range[1]])/memCap)

        cpuDemandUtil.append(np.average(cpuD[data_range[0]:data_range[1]])/cpuCap)  
        memDemandUtil.append(np.average(memD[data_range[0]:data_range[1]])/memCap)

        cpuUsageUtil.append(np.average(cpuU[data_range[0]:data_range[1]])/cpuCap)  
        memUsageUtil.append(np.average(memU[data_range[0]:data_range[1]])/memCap)

    Y_MAX = np.maximum(np.amax(cpuReqUtil),Y_MAX)
    Y_MAX = np.maximum(np.amax(memReqUtil),Y_MAX)
    Y_MAX = np.maximum(np.amax(cpuDemandUtil),Y_MAX)
    Y_MAX = np.maximum(np.amax(memDemandUtil),Y_MAX)
    Y_MAX = np.maximum(np.amax(cpuUsageUtil),Y_MAX)
    Y_MAX = np.maximum(np.amax(memUsageUtil),Y_MAX)

    x = np.arange(methodsNum) 
    width = GBAR_WIDTH/2
    ## plot
    # request    
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    rects1 = ax.bar(x - width/2, cpuReqUtil,  width, label=STR_CPU, color=COLOR_CPU)
    rects2 = ax.bar(x + width/2, memReqUtil,  width, label=STR_MEM, color=COLOR_MEM)
    labels = methods
    ax.set_ylabel('Request')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend( loc='best')    
    plt.ylim(0,Y_MAX*1.1)

    fig.savefig(FIG_PATH+"/request-avg.pdf", bbox_inches='tight')

    # demand
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    rects1 = ax.bar(x - width/2, cpuDemandUtil,  width, label=STR_CPU, color=COLOR_CPU)
    rects2 = ax.bar(x + width/2, memDemandUtil,  width, label=STR_MEM, color=COLOR_MEM)
    labels = methods
    ax.set_ylabel('Demand')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend( loc='best')    
    plt.ylim(0,Y_MAX*1.1)

    fig.savefig(FIG_PATH+"/demand-avg.pdf", bbox_inches='tight')

    # usage
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    rects1 = ax.bar(x - width/2, cpuUsageUtil,  width, label=STR_CPU, color=COLOR_CPU)
    rects2 = ax.bar(x + width/2, memUsageUtil,  width, label=STR_MEM, color=COLOR_MEM)
    labels = methods
    ax.set_ylabel('Usage')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend( loc='best')    
    plt.ylim(0,Y_MAX*1.1)

    fig.savefig(FIG_PATH+"/usage-avg.pdf", bbox_inches='tight')

if plotUtilization:
    # Y_MAX = np.amax(cpuRequests)
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuRequests[i])*tick,tick), cpuRequests[i], color=colors[i])

    plt.plot(range(0,len(cpuAllocatables[0])*tick,tick), cpuAllocatables[0], color=COLOR_CAP)
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_CPU_CORES)
    # plt.ylim(0,Y_MAX)
    fig.savefig(FIG_PATH+"/total-request-cpu.pdf", bbox_inches='tight')
    

    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(memRequests[i])*tick,tick), memRequests[i], color=colors[i])

    plt.plot(range(0,len(memAllocatables[0])*tick,tick), memAllocatables[0], color=COLOR_CAP)
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_MEM_GB)
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/total-request-mem.pdf", bbox_inches='tight')

if plotUtilization:
    # Y_MAX = np.amax(cpuRequests)
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuUsages[i])*tick,tick), cpuUsages[i], color=colors[i])
    
    plt.plot(range(0,len(cpuAllocatables[0])*tick,tick), cpuAllocatables[0], color=COLOR_CAP)
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_CPU_CORES)
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/total-demand-cpu.pdf", bbox_inches='tight')

    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(memUsages[i])*tick,tick), memUsages[i], color=colors[i])
    
    plt.plot(range(0,len(memAllocatables[0])*tick,tick), memAllocatables[0], color=COLOR_CAP)
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_MEM_GB)
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/total-demand-mem.pdf", bbox_inches='tight')

if plotUtilization:
    # Y_MAX = np.amax(cpuRequests)
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(cpuAllocations[i])*tick,tick), cpuAllocations[i], color=colors[i])
    
    plt.plot(range(0,len(cpuAllocatables[0])*tick,tick), cpuAllocatables[0], color=COLOR_CAP)
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_CPU_CORES)
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/total-usage-cpu.pdf", bbox_inches='tight')

    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(memAllocations[i])*tick,tick), memAllocations[i], color=colors[i])
    
    plt.plot(range(0,len(memAllocatables[0])*tick,tick), memAllocatables[0], color=COLOR_CAP)
    legends = methods
    legends.append('capacity')

    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_MEM_GB)
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/total-usage-mem.pdf", bbox_inches='tight')

## plot performance: number of overload nodes.
if plotOverload:
    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        plt.plot(range(0,len(overloadNodes[i])*tick,tick), overloadNodes[i], color=colors[i])

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
        plt.plot(range(0,len(overbookNodes[i])*tick,tick), overbookNodes[i], color=colors[i])
    
    legends = methods   
    plt.legend(legends, loc='best')
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_NODES)
    # plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/overbook.pdf", bbox_inches='tight')

## plot QoS
if plotQoS:

    fig = plt.figure(figsize=FIG_ONE_COL)
    for i in range(methodsNum):
        qos = QoSs[i][data_range[0]:data_range[1]]
        plt.plot(range(0,len(qos)*tick,tick), qos, color=colors[i])
    
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
        qos = QoSs[i][data_range[0]:data_range[1]]
        cdf = compute_cdf(numpy.array(qos))
        plt.plot(cdf.x, cdf.p, color=colors[i])

    legends = methods   
    plt.legend(legends, loc='best')
    plt.xlabel(STR_QoS)
    plt.ylabel(STR_CDF)
    plt.ylim(0,1.02)
    plt.xlim(0,1.02)

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