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

############# PLOTTING ##############
if not os.path.exists(FIG_PATH):
    os.makedirs(FIG_PATH)


plotClusterSize = True

methodNames = [STR_WORSTFIT, STR_OVERSUB, STR_FLEX_F, STR_FLEX_L]
methodsNum = len(methodNames)
colors = [COLOR_WORST_FIT, COLOR_OVER_SUB, COLOR_PROPOSED_1, COLOR_PROPOSED_2]


if plotClusterSize:
    clusterSizes = [3000, 3500, 3600, 3700, 3800, 4000]

    ## QoS violation
    data = []
    worstFitVals = [0, 0, 0, 0, 0, 0]
    data.append(worstFitVals)
    oversubVals = [14.9, 14.1, 9.2, 6.2, 9.9, 2.5]
    data.append(oversubVals)
    flexFVals = [1.1, 1.2, 0.9, 1.6, 1.1, 0.6]
    data.append(flexFVals)
    flexLVals = [4.4, 2.6, 2.6, 2.2, 1.3, 0.5]
    data.append(flexLVals)

    Y_MAX = np.amax(oversubVals)*1.1
    fig = plt.figure(figsize=FIG_ONE_COL)
    max_len = 0
    for i in range(methodsNum):
        plt.plot(clusterSizes, data[i], color=colors[i])
    
    legends = methodNames
    plt.legend(legends, loc='best')
    plt.xlabel(STR_Cluster_Size)
    plt.ylabel(STR_QoS_Violation)
    plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/sensitivity_cluster_size_qos.pdf", bbox_inches='tight')

    ## Memory Usage
    data = []
    worstFitVals = [49, 48, 47, 46, 45, 44]
    data.append(worstFitVals)
    oversubVals = [79, 79, 77, 76, 73, 69]
    data.append(oversubVals)
    flexFVals = [81, 77, 74, 74, 74, 69]
    data.append(flexFVals)
    flexLVals = [79, 74, 75, 75, 74, 69]
    data.append(flexLVals)

    Y_MAX = 100
    fig = plt.figure(figsize=FIG_ONE_COL)
    max_len = 0
    for i in range(methodsNum):
        plt.plot(clusterSizes, data[i], color=colors[i])
    
    legends = methodNames
    plt.legend(legends, loc='best')
    plt.xlabel(STR_Cluster_Size)
    plt.ylabel("usage (%)")    
    plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/sensitivity_cluster_size_mem_usage.pdf", bbox_inches='tight')

    ## CPU Requests
    data = []
    worstFitVals = [100, 100, 100, 100, 100, 100]
    data.append(worstFitVals)
    oversubVals = [199, 199, 196, 190, 178, 171]
    data.append(oversubVals)
    flexFVals = [215, 193, 187, 185, 178, 171]
    data.append(flexFVals)
    flexLVals = [170, 181, 184, 184, 178, 171]
    data.append(flexLVals)

    Y_MAX = 220
    fig = plt.figure(figsize=FIG_ONE_COL)
    max_len = 0
    for i in range(methodsNum):
        plt.plot(clusterSizes, data[i], color=colors[i])
    
    legends = methodNames
    plt.legend(legends, loc='best')
    plt.xlabel(STR_Cluster_Size)
    plt.ylabel("request (%)")    
    plt.ylim(0,Y_MAX)

    fig.savefig(FIG_PATH+"/sensitivity_cluster_size_cpu_req.pdf", bbox_inches='tight')