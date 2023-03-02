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
logging.getLogger("prophet").setLevel(logging.ERROR)
logging.getLogger("cmdstanpy").disabled = True
logging.getLogger('matplotlib.font_manager').disabled = True


def to_date(epoch_now):
    return dt.datetime.utcfromtimestamp(epoch_now).strftime("%Y-%m-%d %H:%M:%S")

def detect_cycle(config):
    local = int(os.getenv("LOCAL", "0"))
    metrics = preprocess(config)
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
            m.fit(df)
            future = m.make_future_dataframe(periods=0) 
            forecast = m.predict(future)
            if local == 1:
                fig = None
                ax = None
                figsize=(10, 6)
                fig = plt.figure(facecolor='w', figsize=figsize)
                ax = fig.add_subplot(111)
                ax.set_title(query_name)
                fig = m.plot(forecast,ax=ax)
                fig.savefig(f'{query_name}-{i}-forcast.png')
        except Exception as e:
            print()
            logging.error(f"Failing builiding forcast - {e}")
            continue

        # find the dataframes having same indices
        forecast_truncated_index =forecast.index.intersection(df.index)
        forecast_truncated = forecast.loc[forecast_truncated_index]

        # Identify the thresholds with some buffer
        buffer = np.max( forecast_truncated['yhat_upper'])
        
        expected = forecast_truncated['yhat']
        # expected = expected.apply(lambda x: round(x, 0))
        expected = expected.astype(int)
        
        indices =m.history[m.history['y'] > buffer].index

        # Get those points that have crossed the threshold
        anomalies  = m.history.iloc[indices] # ------> This has the thresholded values and more important timestamp

        if len(anomalies) != 0:
            logging.warning(f"Found {len(anomalies)} anomalies in {query_name}")
            for index, row in anomalies.iterrows():
                logging.warning(f"[{query_name}] time: {row['ds']} expected: {expected[index]} actual: {row['y']}")
            if local == 1:
                fig = None
                ax = None
                figsize=(10, 6)
                fig = plt.figure(facecolor='w', figsize=figsize)
                ax = fig.add_subplot(111)
                ax.set_title(query_name)
                fig = m.plot(forecast,ax=ax)
                ax.plot(anomalies['ds'].dt.to_pydatetime(), anomalies['y'], 'r.',
                        label='Thresholded data points')
                fig.savefig(f'{query_name}-{i}-anomaly.png')
        else:
            logging.debug(f"No anomalies found for {query_name}")
        i+=1
