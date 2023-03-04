import json
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import warnings
import logging
warnings.simplefilter(action='ignore', category=FutureWarning)
# from pythonjsonlogger import jsonlogger
# logger = logging.getLogger()

# logHandler = logging.StreamHandler()
# formatter = jsonlogger.JsonFormatter()
# logHandler.setFormatter(formatter)
# logger.addHandler(logHandler)

from prometheus_api_client import PrometheusConnect
from prometheus_api_client.utils import parse_datetime
from datetime import timedelta

def preprocess(config, logger):
    prom_url = config["prom_url"]
    prom = PrometheusConnect(url = prom_url)
    queries = config["queries"]
    metrics = []
    for query in queries:
        ## set configs
        start_time = parse_datetime(query["detection_window"])
        end_time = parse_datetime("now")
        prom_expression = query["query"]
        sensitivity = float(query["sensitivity"])
        step = int(query["resolution"])
        query_name = query["name"]

        data = prom.custom_query_range(
            prom_expression,  # this is the metric name and label config
            start_time=start_time,
            end_time=end_time,
            step=step,
        )

        columns = ['ds', 'y']
        if len(data) == 0:
            logger.error(f"query {prom_expression} returned 0 results")
        for metric in data: 
            lst = metric['values']
            extra_data = metric['metric']
            df = pd.DataFrame(lst, columns=columns)
            metrics.append((df, extra_data, sensitivity, query_name))

    return metrics
