# This script analyzes Trantor benchmark output files given on the command line
# and produces summarizing statistics.
#
# Usage:
#     python3 analyze-bench-output.py experiment-id destination-file input-data-file [input-data-file [...]]

import sys
import json

TIME_FACTOR = 1000000000  # Corresponds to the input data using nanoseconds as units of time.
DATA_FACTOR = (1024 * 1024)  # Corresponds to the input data using bytes and output data reported in MiB

# ====================================================================================================
# Helper functions
# ====================================================================================================


def accumulate(acc: dict, new: dict, f):
    for k, v in new.items():
        if k in acc:
            acc[k] = f(acc[k], v)
        else:
            acc[k] = v


def avg_rate(data, sampling_period):
    if len(data.keys()) == 0:
        return 0
    duration = max(data.keys())
    if duration == 0:
        return 0
    else:
        # We need to add one sampling period to the denominator, as counting starts at 0.
        # The constant time_factor normalizes the time units to seconds.
        return sum(data.values()) * TIME_FACTOR / (duration + sampling_period)


# ====================================================================================================
# Stats
# ====================================================================================================
# This represents the accumulated data from multiple nodes' benchmark output


class Stats:
    def __init__(self):
        self.num_datasets = 0
        self.latency_hist = {}
        self.delivered_txs = {}
        self.sampling_period = 0
        self.upload_rate = {}
        self.download_rate = {}
        self.loss_rate = {}

    def add_dataset(self, data):
        self.num_datasets += 1

        # Set sampling period, or do a sanity check if sampling period already set.
        sampling_period = data["Client"]["SamplingPeriod"]
        if data["Client"]["SamplingPeriod"] != data["Net"]["SamplingPeriod"]:
            raise ValueError(f"Inconsistent data 'Client' and 'Net' sampling period mismatch: {data['Client']['SamplingPeriod']} and {data['Net']['SamplingPeriod']}")
        elif self.sampling_period == 0:
            self.sampling_period = data["Client"]["SamplingPeriod"]
        elif self.sampling_period != data["Client"]["SamplingPeriod"]:
            raise ValueError(f"Trying to add dataset with sampling period {sampling_period} to dataset with sampling period {self.sampling_period}")

        accumulate(self.latency_hist, {int(lat): n for lat, n in data["Client"]["LatencyHistogram"].items()}, (lambda x, y: x+y))
        accumulate(self.delivered_txs, {int(t): n for t, n in data["Client"]["DeliveredTxs"].items()}, (lambda x, y: x+y))
        accumulate(self.upload_rate, {int(t): r for t, r in data["Net"]["TotalSent"].items()}, (lambda x, y: x+y))
        accumulate(self.download_rate, {int(t): r for t, r in data["Net"]["TotalReceived"].items()}, (lambda x, y: x+y))
        accumulate(self.loss_rate, {int(t): r for t, r in data["Net"]["TotalDropped"].items()}, (lambda x, y: x+y))

    def to_json(self):
        return json.dumps({"LatencyHistogram": self.latency_hist, "DeliveredTxs": self.delivered_txs}, indent=2)

    def throughput(self):
        if len(self.delivered_txs.keys()) == 0:
            return 0

        duration = max(self.delivered_txs.keys())
        if duration == 0:
            return 0
        else:
            # We need to add one sampling period to the denominator, as counting starts at 0.
            # The constant time_factor normalizes the time units to seconds.
            return sum(self.delivered_txs.values()) * TIME_FACTOR / (duration + self.sampling_period)

    def avg_upload_rate(self):
        return avg_rate(self.upload_rate, self.sampling_period) / (self.num_datasets * DATA_FACTOR)

    def avg_download_rate(self):
        return avg_rate(self.download_rate, self.sampling_period) / (self.num_datasets * DATA_FACTOR)

    def avg_loss_rate(self):
        return avg_rate(self.loss_rate, self.sampling_period) / (self.num_datasets * DATA_FACTOR)

    def avg_latency(self):
        total_txs = sum(self.latency_hist.values())
        if total_txs == 0:
            return 0
        else:
            return sum([lat * num_txs for lat, num_txs in self.latency_hist.items()]) / (total_txs * TIME_FACTOR)

    def latency_pctile(self, pctile):
        total_txs = sum(self.latency_hist.values())
        if total_txs == 0:
            return 0

        tx_count = 0
        for latency in sorted(self.latency_hist.keys()):
            tx_count += self.latency_hist[latency]
            if tx_count / total_txs > pctile:
                return latency / TIME_FACTOR
        return max(self.latency_hist.keys()) / TIME_FACTOR


# ====================================================================================================
# Main code
# ====================================================================================================

exp_id = sys.argv[1]
config_file = sys.argv[2]
output_file = sys.argv[3]
stats = Stats()

with open(config_file, 'r') as file:
    config = json.load(file)

for input_file in sys.argv[4:]:
    try:
        print(f"processing: {input_file}")
        with open(input_file, 'r') as file:
            stats.add_dataset(json.load(file))
    except FileNotFoundError:
        print(f"File '{input_file}' not found.")
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON file {input_file}: {e}")
        sys.exit(1)

latency_hist = {int(lat) / TIME_FACTOR * 1000: stats.latency_hist[lat] for lat in stats.latency_hist.keys()}
delivered_txs = {t / TIME_FACTOR * 1000: stats.delivered_txs[t] for t in stats.delivered_txs.keys()}

num_nodes = len(config["Trantor"]["Iss"]["InitialMembership"]["Nodes"])

with open(output_file, 'w') as file:
    json.dump({
        "experiment-id": exp_id,
        "params": {
            "num-nodes": num_nodes,
            "clients-node": int(config["TxGen"]["NumClients"]),
            "clients-total": int(config["TxGen"]["NumClients"])*num_nodes,
            "max-batch-tx": config["Trantor"]["Mempool"]["MaxTransactionsInBatch"],
            "max-batch-bytes": int(config["Trantor"]["Mempool"]["MaxPayloadInBatch"]),
            "batch-timeout": int(config["Trantor"]["Mempool"]["BatchTimeout"])/TIME_FACTOR,
            "duration-target": int(config["Duration"])/TIME_FACTOR,
            "segment-length": int(config["Trantor"]["Iss"]["SegmentLength"]),
        },
        "aggregates": {
            "duration-actual": max(stats.delivered_txs.keys())/TIME_FACTOR,
            "throughput": stats.throughput(),
            "latency-avg": stats.avg_latency(),
            "latency-median": stats.latency_pctile(0.5),
            "latency-95p": stats.latency_pctile(0.95),
            "latency-max": stats.latency_pctile(1),
            "net-down-avg": stats.avg_download_rate(),
            "net-up-avg": stats.avg_upload_rate(),
            "net-loss-avg": stats.avg_loss_rate(),
        },
        "plots": {
            "latency-hist": latency_hist,
            "throughput-timeline": delivered_txs,
        },
    }, file, sort_keys=True, indent=2)
