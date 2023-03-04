import json
import pandas as pd
from prophet import Prophet
from prophet.plot import add_changepoints_to_plot
import numpy as np
from matplotlib import pyplot as plt
from preprocess import preprocess
import time, os
import datetime as dt
import logging
from prometheus_client import Counter, Gauge

logging.getLogger("prophet").setLevel(logging.ERROR)
logging.getLogger("cmdstanpy").disabled = True
logging.getLogger('matplotlib.font_manager').disabled = True

# from pythonjsonlogger import jsonlogger
# logger = logging.getLogger()

# logHandler = logging.StreamHandler()
# formatter = jsonlogger.JsonFormatter()
# logHandler.setFormatter(formatter)
# logger.addHandler(logHandler)
g = Gauge('anomaly_counter', 'Total number of anomalies found')

def to_date(epoch_now):
    return dt.datetime.utcfromtimestamp(epoch_now).strftime("%Y-%m-%d %H:%M:%S")

def detect_cycle(config, logger):
    local = int(os.getenv("LOCAL", "0"))
    metrics = preprocess(config, logger)
    anoamly_counter = 0
    i = 0
    for item in metrics:
        epoch_now = time.time()
        df = item[0]
        extra_data = item[1]
        sensitivity = item[2]
        query_name = item[3]
        df["ds"] = df["ds"].apply(to_date)
        m = Prophet(changepoint_prior_scale=sensitivity,changepoint_range=0.9,
                    interval_width=0.95,
                    weekly_seasonality=20,
                    daily_seasonality=20,
                    seasonality_mode='multiplicative')
        try:
            if len(df) < 2:
                logger.info("query returned less then 2 results. skipiing.")
                continue
            m.fit(df)
            future = m.make_future_dataframe(periods=0) 
            forecast = m.predict(future)
            if local == 1:
                fig = None
                ax = None
                figsize=(10, 6)
                fig = plt.figure(facecolor='w', figsize=figsize)
                ax = fig.add_subplot(111)
                ax.set_title(query_name + f"{extra_data}")
                fig = m.plot(forecast,ax=ax)
                fig.savefig(f'{query_name}-{i}-forcast.png')
        except Exception as e:
            print()
            logger.error(f"Failing builiding forcast - {e}")
            continue

        # find the dataframes having same indices
        forecast_truncated_index =forecast.index.intersection(df.index)
        forecast_truncated = forecast.loc[forecast_truncated_index]

        # Identify the thresholds with some buffer
        buffer = np.max( forecast_truncated['yhat_upper'] * 1.5)
        
        expected = forecast_truncated['yhat']
        # expected = expected.apply(lambda x: round(x, 0))
        expected = expected.astype(int)
        
        indices =m.history[m.history['y'] > buffer].index

        # Get those points that have crossed the threshold
        anomalies  = m.history.iloc[indices] # ------> This has the thresholded values and more important timestamp

        if len(anomalies) != 0:
            logger.warning(f"Found {len(anomalies)} anomalies for {query_name} in {extra_data}")
            for index, row in anomalies.iterrows():
                anoamly_counter+=1
                logger.warning(f"[{query_name}] {extra_data} time: {row['ds']} expected: {expected[index]} actual: {row['y']}")
            if local == 1:
                fig = None
                ax = None
                figsize=(10, 6)
                fig = plt.figure(facecolor='w', figsize=figsize)
                ax = fig.add_subplot(111)
                ax.set_title(query_name + f"{extra_data}")
                fig = m.plot(forecast,ax=ax)
                ax.plot(anomalies['ds'].dt.to_pydatetime(), anomalies['y'], 'r.',
                        label='Thresholded data points')
                fig.savefig(f'{query_name}-{i}-anomaly.png')
        else:
            logger.debug(f"No anomalies found for {query_name} in {extra_data}")
        i+=1
    g.set(anoamly_counter)
