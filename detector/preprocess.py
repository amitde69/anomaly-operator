import json
import pandas as pd
import numpy as np
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
from datetime import timedelta, datetime

def preprocess(config, logger):
    prom_url = config["prom_url"]
    prom = PrometheusConnect(url = prom_url)
    queries = config["queries"]
    train_metrics = []
    predict_metrics = []
    for query in queries:
        ## set configs
        start_time = parse_datetime(query["train_window"])
        end_time = parse_datetime("now")
        prom_expression = query["query"]

        flexibility = 0.05
        if "flexibility" in query:
            flexibility = float(query["flexibility"])

        detection_window_hours = 1
        if "detection_window_hours" in query:
            detection_window_hours = int(query["detection_window_hours"])

        diff = end_time - start_time
        diff_in_hours = diff.total_seconds() / 3600
        step = diff_in_hours
        if "resolution" in query:
            step = int(query["resolution"])
        
        buffer_pct = 1
        if "buffer_pct" in query:
            buffer_pct = int(query["buffer_pct"]) / 100
        
        query_name = query["name"]
        try:
            train_data = prom.custom_query_range(
                prom_expression,  # this is the metric name and label config
                start_time=start_time,
                end_time=end_time,
                step=step,
            )
        except Exception as e:
            logger.error(f"Failed while querying {query_name}: {e}")

        columns = ['ds', 'y']
        if len(train_data) == 0:
            logger.error(f"query {prom_expression} returned 0 results")
        for metric in train_data: 
            lst = metric['values']
            extra_data = metric['metric']
            df = pd.DataFrame(lst, columns=columns)
            train_metrics.append((df, extra_data, flexibility, query_name, buffer_pct, detection_window_hours))

    return train_metrics
