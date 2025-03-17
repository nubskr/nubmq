import csv
import datetime
import matplotlib.pyplot as plt
from collections import defaultdict
import numpy as np

def read_timestamps(csv_file_path):
    timestamps = []
    
    with open(csv_file_path, mode='r') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            try:
                timestamp_ms = int(row['Timestamp_ms'])  # Read timestamp in milliseconds
                timestamp_sec = timestamp_ms / 1000  # Convert to seconds
                timestamps.append(timestamp_sec)
            except ValueError:
                print(f"Invalid data row: {row}")

    return timestamps

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

# def plot_throughput_graph(throughput_data):
#     if not throughput_data:
#         print("No throughput data to plot.")
#         return

#     # Convert timestamp keys to human-readable format
#     timestamps = sorted(throughput_data.keys())
#     throughput_values = [throughput_data[ts] for ts in timestamps]
#     human_readable_timestamps = [datetime.datetime.fromtimestamp(ts).strftime('%H:%M:%S') for ts in timestamps]

#     # Create the plot
#     plt.figure(figsize=(14, 7))
#     plt.plot(human_readable_timestamps, throughput_values, marker='o', linestyle='-', linewidth=2.5, markersize=8, color='#FF5733', label="Ops/sec")

#     # Enhancements
#     plt.xlabel("Time (HH:MM:SS)", fontsize=14, fontweight='bold', color='black')
#     plt.ylabel("Operations per Second", fontsize=14, fontweight='bold', color='black')
#     plt.title("System Throughput Over Time", fontsize=16, fontweight='bold', color='black')
#     plt.xticks(rotation=45, fontsize=12)
#     plt.yticks(fontsize=12)
#     plt.grid(True, linestyle='--', linewidth=0.6, alpha=0.7)
#     plt.legend(fontsize=12, loc='upper right')
#     plt.tight_layout()

#     # Show the plot
#     plt.show()

def plot_throughput_graph(throughput_data):
    if not throughput_data:
        print("No throughput data to plot.")
        return

    # Convert timestamps to relative time (starting from 0)
    timestamps = sorted(throughput_data.keys())
    relative_timestamps = [ts - timestamps[0] for ts in timestamps]  # Start from 0
    throughput_values = [throughput_data[ts] for ts in timestamps]

    # Create the plot
    plt.figure(figsize=(14, 7))
    plt.plot(relative_timestamps, throughput_values, marker='o', linestyle='-', linewidth=2.5, markersize=6, color='#2E86C1', label="Ops/sec")

    # Enhancements
    plt.xlabel("Time (seconds)", fontsize=14, fontweight='bold', color='black')
    plt.ylabel("Operations per Second", fontsize=14, fontweight='bold', color='black')
    plt.title("System Throughput Over Time", fontsize=16, fontweight='bold', color='black')
    plt.xticks(fontsize=12)
    plt.yticks(fontsize=12)
    plt.grid(True, linestyle='--', linewidth=0.6, alpha=0.7)
    plt.legend(fontsize=12, loc='upper right')
    plt.tight_layout()

    # Show the plot
    plt.show()

if __name__ == "__main__":
    # Path to CSV file
    requests_csv_file = '/Users/nubskr/Downloads/top_requests.csv'
    
    # Read timestamps
    timestamps = read_timestamps(requests_csv_file)

    # Calculate per-second throughput
    throughput_data = calculate_throughput_over_time(timestamps)

    # Calculate and print average throughput
    total_ops = sum(throughput_data.values())
    total_seconds = len(throughput_data)
    avg_throughput = total_ops / total_seconds if total_seconds > 0 else 0

    # Find peak throughput
    if throughput_data:
        peak_second = max(throughput_data, key=throughput_data.get)
        peak_throughput = throughput_data[peak_second]
        peak_timestamp_human = datetime.datetime.fromtimestamp(peak_second).strftime('%Y-%m-%d %H:%M:%S')

        print(f"\n[Peak Throughput]")
        print(f"Highest throughput occurred at {peak_timestamp_human}")
        print(f"Peak Throughput: {peak_throughput} ops/sec")

    print(f"\n[Average Throughput]")
    print(f"Total Operations: {total_ops}")
    print(f"Total Time Tracked: {total_seconds} seconds")
    print(f"Average Throughput: {avg_throughput:.2f} ops/sec")

    # Plot throughput graph
    plot_throughput_graph(throughput_data)
