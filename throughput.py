import csv

def calculate_throughput(csv_file_path, num_connections):
    durations = []

    # Read the CSV and collect all durations
    with open(csv_file_path, mode='r') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            try:
                duration_ms = float(row['Duration_ms'])
                durations.append(duration_ms)
            except ValueError:
                print(f"Invalid duration value: {row['Duration_ms']}")

    if not durations:
        print("No valid durations found in the CSV.")
        return

    # Calculate average duration in milliseconds
    avg_duration_ms = sum(durations) / len(durations)
    avg_duration_sec = avg_duration_ms / 1000  # Convert to seconds

    # Calculate throughput
    throughput = num_connections / avg_duration_sec

    print(f"Total Operations: {len(durations)}")
    print(f"Average Duration per Operation: {avg_duration_ms:.6f} ms ({avg_duration_sec:.6f} s)")
    print(f"Number of Concurrent Connections: {num_connections}")
    print(f"Throughput: {throughput:,.2f} operations/second")

if __name__ == "__main__":
    # Path to your CSV file
    csv_file = 'top_set_durations.csv'
    
    # Number of concurrent connections (as per your configuration)
    concurrent_connections = 50
    
    calculate_throughput(csv_file, concurrent_connections)

