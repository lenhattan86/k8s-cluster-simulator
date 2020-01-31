# https://matplotlib.org/3.1.0/gallery/index.html
import matplotlib
import matplotlib.pyplot as plt
import numpy as np

def autolabel(rects, ax, xpos='center'):
    """
    Attach a text label above each bar in *rects*, displaying its height.

    *xpos* indicates which side to place the text w.r.t. the center of
    the bar. It can be one of the following {'center', 'right', 'left'}.
    """

    ha = {'center': 'center', 'right': 'left', 'left': 'right'}
    offset = {'center': 0, 'right': 1, 'left': -1}

    for rect in rects:
        height = rect.get_height()
        if height%1==0:
            height=int(height)
            
        ax.annotate('{}'.format(height),
                    xy=(rect.get_x() + rect.get_width() / 2, height),
                    xytext=(offset[xpos]*3, 3),  # use 3 points offset
                    textcoords="offset points",  # in both directions
                    ha=ha[xpos], va='bottom')

def plot_group_bar(fig_size, width, xticklabels, ylabel, legends, data, title, colors):
    x = np.arange(len(xticklabels))  # the label locations
    fig, ax = plt.subplots(figsize=fig_size)   
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
            rects = ax.bar(x - n/2*width + width/2 + i*width, data[i], width, label='mem')
        else:
            rects = ax.bar(x - n/2*width + width/2 + i*width, data[i], width, label='mem')
        Y_MAX = np.maximum(np.amax(data[i]),Y_MAX)
        # Y_MIN = np.mininum(np.amin(data[i]),Y_MIN)
        autolabel(rects, ax)

    # fig.tight_layout()
    plt.ylim(Y_MIN,Y_MAX*1.1)
    # plt.show()
    return fig
