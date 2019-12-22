import matplotlib
import matplotlib.pyplot as plt
import numpy as np

def autolabel(rects, ax):
    """Attach a text label above each bar in *rects*, displaying its height."""
    for rect in rects:
        height = rect.get_height()
        ax.annotate('{}'.format(height),
                    xy=(rect.get_x() + rect.get_width() / 2, height),
                    xytext=(0, 3),  # 3 points vertical offset
                    textcoords="offset points",
                    ha='center', va='bottom')

def plot_group_bar(fig_size, width, xticklabels, ylabel, legends, data, title, colors):
    x = np.arange(len(xticklabels))  # the label locations
    fig, ax = plt.subplots(figsize=fig_size)   
    # fig, ax = plt.figure(figsize=fig_size)
    ax.set_ylabel(ylabel)
    if len(title) > 0:
        ax.set_title(title)

    ax.set_xticks(x)
    ax.set_xticklabels(xticklabels)
    ax.legend()

    n = len(legends)    
    Y_MIN = 0
    Y_MAX = 0
    for i in range(n): 
        if len(colors)==0:
            rects = ax.bar(x - n/2*width + width/2 + i*width, data[i], width, label="mem")
        else:
            rects = ax.bar(x - n/2*width + width/2 + i*width, data[i], width, label="mem")
        Y_MAX = np.maximum(np.amax(data[i]),Y_MAX)
        # Y_MIN = np.mininum(np.amin(data[i]),Y_MIN)
        autolabel(rects, ax)

    # fig.tight_layout()
    plt.ylim(0,Y_MAX*1.1)
    plt.show()
    return fig

labels = ['G1', 'G2', 'G3', 'G4', 'G5']
men_means = [20, 34, 30, 35, 27]
women_means = [25, 32, 34, 20, 25]
data=[]
data.append(men_means)
data.append(women_means)
plot_group_bar((4,3), 0.3, labels, "score", ("women","men"), data, "",[])