import csv
import numpy as np
import datetime
import matplotlib.pyplot as plt
from collections import defaultdict

def read_durations_with_timestamps(csv_file_path):
    durations = []
    timestamps = []
    
    with open(csv_file_path, mode='r') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            try:
                duration_ms = float(row['Duration_ms'])
                timestamp_ms = int(row['Timestamp_ms'])  # Read timestamp in milliseconds
                timestamp_sec = timestamp_ms / 1000  # Convert to seconds
                durations.append(duration_ms)
                timestamps.append(timestamp_sec)
            except ValueError:
                print(f"Invalid data row: {row}")

    return timestamps, durations

def calculate_throughput_over_time(timestamps):
    if not timestamps:
        print("No valid timestamps found.")
        return {}

    # Group operations by second
    throughput_per_second = defaultdict(int)
    
    for ts in timestamps:
        second = int(ts)  # Convert to whole second
        throughput_per_second[second] += 1  # Count ops per second

    return throughput_per_second

def plot_nicer_throughput_graph(throughput_data):
    if not throughput_data:
        print("No throughput data to plot.")
        return

    # Convert timestamp keys to human-readable format
    timestamps = sorted(throughput_data.keys())
    throughput_values = [throughput_data[ts] for ts in timestamps]
    human_readable_timestamps = [datetime.datetime.fromtimestamp(ts).strftime('%H:%M:%S') for ts in timestamps]

    # Create the plot
    plt.figure(figsize=(14, 7))
    plt.plot(human_readable_timestamps, throughput_values, marker='o', linestyle='-', linewidth=2.5, markersize=8, color='#FF5733', label="Ops/sec")

    # Enhancements
    plt.xlabel("Time (HH:MM:SS)", fontsize=14, fontweight='bold', color='black')
    plt.ylabel("Operations per Second", fontsize=14, fontweight='bold', color='black')
    plt.title("System Throughput Over Time (SET + GET)", fontsize=16, fontweight='bold', color='black')
    plt.xticks(rotation=45, fontsize=12)
    plt.yticks(fontsize=12)
    plt.grid(True, linestyle='--', linewidth=0.6, alpha=0.7)
    plt.legend(fontsize=12, loc='upper right')
    plt.tight_layout()

    # Show the plot
    plt.show()

if __name__ == "__main__":
    # Paths to CSV files
    set_csv_file = 'top_set_durations.csv'
    get_csv_file = 'top_get_durations.csv'
    
    # Read durations and timestamps
    set_timestamps, set_durations = read_durations_with_timestamps(set_csv_file)
    get_timestamps, get_durations = read_durations_with_timestamps(get_csv_file)

    # Calculate per-second throughput
    set_throughput_data = calculate_throughput_over_time(set_timestamps)
    get_throughput_data = calculate_throughput_over_time(get_timestamps)

    # Combine throughput data
    combined_throughput_data = defaultdict(int)
    for sec in set_throughput_data:
        combined_throughput_data[sec] += set_throughput_data[sec]
    for sec in get_throughput_data:
        combined_throughput_data[sec] += get_throughput_data[sec]

    # Plot improved throughput graph
    plot_nicer_throughput_graph(combined_throughput_data)
