import matplotlib
import matplotlib.pyplot as plt
import numpy as np

labels = ['G1', 'G2', 'G3', 'G4', 'G5']
men_means = [20, 34, 30, 35, 27]
women_means = [25, 32, 34, 20, 25]

def autolabel(rects):
    """Attach a text label above each bar in *rects*, displaying its height."""
    for rect in rects:
        height = rect.get_height()
        ax.annotate('{}'.format(height),
                    xy=(rect.get_x() + rect.get_width() / 2, height),
                    xytext=(0, 3),  # 3 points vertical offset
                    textcoords="offset points",
                    ha='center', va='bottom')

def plot_group_bar(width, xticklabels, ylabel, legends, data, ):
    x = np.arange(len(xticklabels))  # the label locations

    fig, ax = plt.subplots()
    n = len(legends)
    for i in range(n): 
        rects = ax.bar(x - width/2, data[i], width, label='Men')

    # Add some text for xticklabels, title and custom x-axis tick labels, etc.
    ax.set_ylabel(ylabel)
    ax.set_title('Scores by group and gender')
    ax.set_xticks(x)
    ax.set_xticklabels(xticklabels)
    ax.legend()

    # autolabel(rects1)
    # autolabel(rects2)
    fig.tight_layout()

    plt.show()