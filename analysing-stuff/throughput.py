import csv
import numpy as np
import datetime
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

def calculate_throughput_over_time(timestamps, operation_type):
    if not timestamps:
        print(f"No valid timestamps found for {operation_type}.")
        return {}, None

    # Group operations by second
    throughput_per_second = defaultdict(int)
    
    for ts in timestamps:
        second = int(ts)  # Convert to whole second
        throughput_per_second[second] += 1  # Count ops per second

    print(f"\n[{operation_type} Throughput Over Time]")
    for second in sorted(throughput_per_second.keys()):
        ts_human = datetime.datetime.fromtimestamp(second).strftime('%Y-%m-%d %H:%M:%S')
        print(f"{ts_human}: {throughput_per_second[second]} ops/sec")

    # Find peak throughput
    peak_second = max(throughput_per_second, key=throughput_per_second.get)
    peak_throughput = throughput_per_second[peak_second]

    return throughput_per_second, (peak_second, peak_throughput)

def calculate_latency_distribution(combined_durations):
    if not combined_durations:
        print("No valid durations found for percentile calculation.")
        return

    percentiles = [50, 75, 90, 95, 97, 98, 99, 99.5, 99.9]
    percentile_values = np.percentile(combined_durations, percentiles)

    print("\n[Latency Percentile Distribution (Combined SET + GET)]")
    for p, value in zip(percentiles, percentile_values):
        print(f"{p}th Percentile: {value:.6f} ms")

if __name__ == "__main__":
    # Paths to CSV files
    set_csv_file = 'top_set_durations.csv'
    get_csv_file = 'top_get_durations.csv'
    
    # Read durations and timestamps
    set_timestamps, set_durations = read_durations_with_timestamps(set_csv_file)
    get_timestamps, get_durations = read_durations_with_timestamps(get_csv_file)

    # Calculate per-second throughput and find peaks
    set_throughput_data, set_peak = calculate_throughput_over_time(set_timestamps, "SET")
    get_throughput_data, get_peak = calculate_throughput_over_time(get_timestamps, "GET")

    # Combine throughput data
    combined_throughput_data = defaultdict(int)
    for sec in set_throughput_data:
        combined_throughput_data[sec] += set_throughput_data[sec]
    for sec in get_throughput_data:
        combined_throughput_data[sec] += get_throughput_data[sec]

    # Find combined peak throughput
    peak_second = max(combined_throughput_data, key=combined_throughput_data.get)
    peak_throughput = combined_throughput_data[peak_second]
    peak_timestamp_human = datetime.datetime.fromtimestamp(peak_second).strftime('%Y-%m-%d %H:%M:%S')

    print(f"\n[Peak Throughput]")
    print(f"Highest throughput occurred at {peak_timestamp_human}")
    print(f"Peak Throughput: {peak_throughput} ops/sec")

    # Combine durations and calculate latency percentiles
    combined_durations = set_durations + get_durations
    calculate_latency_distribution(combined_durations)
