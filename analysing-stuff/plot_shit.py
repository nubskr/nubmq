import pandas as pd
import matplotlib.pyplot as plt

def plot_durations(csv_file, title, output_file):
    # Read the CSV file
    data = pd.read_csv(csv_file)
    
    # Convert columns to appropriate data types
    data['Rank'] = data['Rank'].astype(int)
    data['Duration_ms'] = data['Duration_ms'].astype(float)
    
    # Create the plot
    plt.figure(figsize=(8, 6))
    plt.plot(data['Rank'], data['Duration_ms'], marker='o', linestyle='-')
    plt.title(title)
    plt.xlabel('Rank')
    plt.ylabel('Duration (ms)')
    # plt.gca().invert_xaxis()  # Show rank 1 on the left
    plt.grid(True)
    plt.tight_layout()
    
    # Save the plot to a file
    plt.savefig(output_file)
    plt.close()

# Plot SET durations
plot_durations('./top_set_durations.csv', 'Top 10 Max SET Response Times', './top_set_durations.png')

# Plot GET durations
plot_durations('./top_get_durations.csv', 'Top 10 Max GET Response Times', './top_get_durations.png')

print("Plots have been saved as 'top_set_durations.png' and 'top_get_durations.png'.")

