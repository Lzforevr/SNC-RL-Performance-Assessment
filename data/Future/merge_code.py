"""
merge_code.py

Description:
    This script merges multiple CSV files containing CPU, memory, and network usage 
    data for Pods in a Kubernetes cluster. The merged data will include CPU usage total,
    CPU usage request, memory usage, memory usage request, network speed in, and network 
    speed out for each Pod.

Author:
    Fangzhou Xie

Notes:
    1. Ensure all input CSV files contain 'timestamp' and 'pod' columns for merging.
    2. The output file will be saved as 'PodMerged.csv'.
"""

import pandas as pd

# Define paths to CSV files to be merged
file_info = {
    "2024-12-14/PodCPUCoreTotal.csv": "cpu_usage",
    "2024-12-14/PodCPUUsageRequest.csv": "cpu_usage_request",
    "2024-12-14/PodMemoryUsageBytes.csv": "memory_usage",
    "2024-12-14/PodMemoryUsageRequest.csv": "memory_usage_request",
    "2024-12-14/PodNetworkIn.csv": "network_speed_in",
    "2024-12-14/PodNetworkOut.csv": "network_speed_out"
}

# Read all CSV files and rename the 'value' column appropriately
dfs = []
for file, new_col in file_info.items():
    df = pd.read_csv(file)
    df = df.rename(columns={"value": new_col})
    dfs.append(df)

# Merge the dataframes on 'timestamp' and 'pod' columns using an outer join
df_merged = dfs[0]
for df in dfs[1:]:
    df_merged = pd.merge(df_merged, df, on=["timestamp", "pod"], how="outer")

# Sort the merged dataframe by 'timestamp' in ascending order
df_merged = df_merged.sort_values(by="timestamp")

# Add container orchestrator and runtime information
df_merged["container orchestrator"] = "Kubernetes"
df_merged["container runtime"] = "containerd"

# Save the merged data to a CSV file
output_path = "PodMerged.csv"
df_merged.to_csv(output_path, index=False)

print(f"Merged CSV file has been saved to {output_path}")