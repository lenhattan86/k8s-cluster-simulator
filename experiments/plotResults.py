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
    NumSatifiesPods = []
    NumPods = []
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
                        cpuUsage = formatCpuQuatity(usageDict[rsName])
                        totalCpuUsage = totalCpuUsage+ cpuUsage
                        if cpuUsage > maxCpuUsage:
                            maxCpuUsage = cpuUsage
                    elif(rsName==memStr):
                        memUsage = formatMemQuatity(usageDict[rsName])
                        totalMemUsage = totalMemUsage+ memUsage
                        if memUsage > maxMemUsage:
                            maxMemUsage = memUsage

                allocatableDict = node['Allocatable']    
                for rsName in allocatableDict:
                    if(rsName==cpuStr):
                        cpuAllocatable = formatCpuQuatity(allocatableDict[rsName])
                        totalCpuCapacity = totalCpuCapacity + cpuAllocatable
                    elif(rsName==memStr):
                        memAllocatable = formatMemQuatity(allocatableDict[rsName])
                        totalMemCapacity = totalMemCapacity + memAllocatable
                
                requestDict = node['TotalResourceRequest']    
                for rsName in requestDict:
                    if(rsName==cpuStr):
                        cpuRequest = formatCpuQuatity(requestDict[rsName])
                        totalCpuRequest = totalCpuRequest + cpuRequest
                    elif(rsName==memStr):
                        memRequest = formatMemQuatity(requestDict[rsName])
                        totalMemRequest = totalMemRequest + memRequest
                
                allocationDict = node['TotalResourceAllocation']    
                for rsName in allocationDict:
                    if(rsName==cpuStr):
                        cpuAllocation = formatCpuQuatity(allocationDict[rsName])
                        totalCpuAllocation = totalCpuAllocation + cpuAllocation
                    elif(rsName==memStr):
                        memAllocation = formatMemQuatity(allocationDict[rsName])
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
                NumSatifiesPods.append(float(queue['NumSatifisedPods']))
                NumPods.append(float(queue['NumPods']))
            if (loads[5]):
                PredPenalty.append(float(queue['PredictionPenalty']))

            i=i+1            
            if line_num > 0 and i >= line_num:
                break
            line = fp.readline()

    fp.close()

    return busyNodes, overloadNodes, overBookNodes, cpuUsages, memUsages, cpuRequests, \
        memRequests, totalCpuAllocations, totalMemAllocations, maxCpuUsages, cpuAllocatables, memAllocatables, \
        QoS, NumSatifiesPods, NumPods, PredPenalty 

def formatCpuQuatity(str):
    strArray = re.split('(\d+)', str)
    val = float(strArray[1])
    scaleStr = strArray[2]
    if(scaleStr == "m"):
        val = val/1000        
    elif (scaleStr == "Mi"):
        val = val/1024
    elif (scaleStr == ""):
        val = val
    else:
        print("error @ formatMemQuatity "+str)

    return val

def formatMemQuatity(str):
    strArray = re.split('(\d+)', str)
    val = float(strArray[1])
    scaleStr = strArray[2]
    if(scaleStr == "m"):
        val = val/(1000*1000)
    elif (scaleStr == "Mi"):
        val = val/(1024*1024)
    elif (scaleStr == "Gi"):
        va = val
    elif (scaleStr == ""): # byte
        val = val/(1024*1024*1024)
    else:
        print("error @ formatMemQuatity "+str)
    return val


methods = ["worstfit","oversub", "proposed"]
colors = [COLOR_WORST_FIT, COLOR_OVER_SUB, COLOR_PROPOSED]
proposed_idx = 2
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
NumSatifiesPods = []
NumPods = []
PredPenalties = []

for m in methods:
    b, ol, ob, u_cpu, u_mem, ur_cpu, ur_mem, a_cpu, a_mem, mu, c_cpu, c_mem, q, nsp, nps, p = loadLog(path+"/kubesim_"+m+".log")
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
    NumSatifiesPods.append(nsp)
    NumPods.append(nps)
    PredPenalties.append(p)

for i in range(methodsNum):
    if (len(cpuRequests[i]) < data_range[1]):
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

    Y_MAX = 2.2

    for i in range(methodsNum):
        cpuR = cpuRequests[i]
        memR = memRequests[i]

        cpuD = cpuUsages[i]
        memD = memUsages[i]

        cpuU = cpuAllocations[i]
        memU = memAllocations[i]

        cpuReqUtil.append(round(np.average(cpuR[data_range[0]:data_range[1]])/cpuCap, 2))  
        memReqUtil.append(round(np.average(memR[data_range[0]:data_range[1]])/memCap, 2))

        cpuDemandUtil.append(round(np.average(cpuD[data_range[0]:data_range[1]])/cpuCap, 2))  
        memDemandUtil.append(round(np.average(memD[data_range[0]:data_range[1]])/memCap, 2))

        cpuUsageUtil.append(round(np.average(cpuU[data_range[0]:data_range[1]])/cpuCap,2))  
        memUsageUtil.append(round(np.average(memU[data_range[0]:data_range[1]])/memCap,2))

    x = np.arange(methodsNum) 
    width = GBAR_WIDTH/2
    ## plot
    # request    
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    rects = ax.bar(x - width, cpuReqUtil,  width, label=STR_CPU, color=COLOR_CPU)
    autolabel(rects, ax)
    rects = ax.bar(x, memReqUtil,  width, label=STR_MEM, color=COLOR_MEM)
    autolabel(rects, ax)
    labels = methods
    ax.set_ylabel('Request')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend( loc='best')    
    plt.ylim(0,Y_MAX*1.1)

    fig.savefig(FIG_PATH+"/request-avg.pdf", bbox_inches='tight')

    # demand
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    rects = ax.bar(x - width, cpuDemandUtil,  width, label=STR_CPU, color=COLOR_CPU)
    autolabel(rects, ax)
    rects = ax.bar(x, memDemandUtil,  width, label=STR_MEM, color=COLOR_MEM)
    autolabel(rects, ax)
    labels = methods
    ax.set_ylabel('Demand')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend( loc='best')    
    plt.ylim(0,Y_MAX*1.1)

    fig.savefig(FIG_PATH+"/demand-avg.pdf", bbox_inches='tight')

    # usage
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    rects = ax.bar(x - width, cpuUsageUtil,  width, label=STR_CPU, color=COLOR_CPU)
    autolabel(rects, ax)
    rects = ax.bar(x, memUsageUtil,  width, label=STR_MEM, color=COLOR_MEM)
    autolabel(rects, ax)
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

    ## plot figures   
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    Y_MAX = 0
    for i in range(methodsNum):
        qos = QoSs[i][data_range[0]:data_range[1]]        
        violation = 0
        for j in range (len(qos)):
            if round(qos[j],2) < target_qos :
                violation = violation + 1
        y = round(violation*100/len(qos),1)
        Y_MAX = max(y,Y_MAX)
        rects = ax.bar(i - BAR_WIDTH/2, y,  BAR_WIDTH, color=colors[i])
        autolabel(rects, ax)

    labels = methods
    ax.set_ylabel(STR_QoS_Violation)
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    plt.ylim([0, Y_MAX*1.2])

    fig.savefig(FIG_PATH+"/qos_violation_time.pdf", bbox_inches='tight')

    ## plot figures   
    fig, ax = plt.subplots(figsize=FIG_ONE_COL)
    Y_MAX = 0
    for i in range(methodsNum):
        nsp = NumSatifiesPods[i][data_range[0]:data_range[1]]        
        nps = NumPods[i][data_range[0]:data_range[1]]        
        violation = 0
        total = 0
        for j in range(len(nsp)):
            total = total + nps[j]
            violation = violation + nps[j] - nsp[j]
        y = round(violation*100/total,1)
        Y_MAX = max(y, Y_MAX)
        rects = ax.bar(i - BAR_WIDTH/2, y,  BAR_WIDTH, color=colors[i])
        autolabel(rects, ax)

    labels = methods
    ax.set_ylabel(STR_QoS_Violation)
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    plt.ylim([0, Y_MAX*1.2])

    fig.savefig(FIG_PATH+"/qos_violation_pod.pdf", bbox_inches='tight')

if plotPredictionPenalty:
    fig = plt.figure(figsize=FIG_ONE_COL)
    plt.plot(range(0,len(PredPenalties[proposed_idx])*tick,tick), PredPenalties[proposed_idx])
    
    plt.xlabel(STR_TIME_MIN)
    plt.ylabel(STR_Pred_Penalty)
    plt.ylim(0,3.1)

    fig.savefig(FIG_PATH+"/pred_penalty.pdf", bbox_inches='tight')

# STR_Pred_Penalty
## show figures
if show:
    plt.show()