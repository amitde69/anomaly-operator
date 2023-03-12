import json
import pandas as pd
from prophet import Prophet
from prophet.plot import add_changepoints_to_plot
import numpy as np
from preprocess import preprocess
import time, os
import datetime as dt
from datetime import timedelta, datetime, timezone
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
    train_metrics = preprocess(config, logger)
    anoamly_counter = 0
    i = 0
    for item in train_metrics:
        epoch_now = time.time()
        df = item[0]
        extra_data = item[1]
        flexibility = item[2]
        query_name = item[3]
        buffer_pct = item[4]
        detection_window_hours = item[5]
        df["ds"] = df["ds"].apply(to_date)
        # df['y'] = 5
        # print(df)
        # for timeindex in range(54, 59):
        #     df['y'][timeindex] = 70
        m = Prophet(changepoint_prior_scale=flexibility,changepoint_range=0.8,
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
            future['Date'] = pd.to_datetime(future.ds)
            model_history = m.history
            model_history['Date'] = pd.to_datetime(model_history.ds)

            # get the timestamp of x hours ago
            x_hours_ago = datetime.utcnow() - timedelta(hours=detection_window_hours)
            last_hours_data = future[future['Date'] >= x_hours_ago]
            last_hours_data_y = model_history[model_history['Date'] >= x_hours_ago]
            if len(last_hours_data) < 1:
                logger.info(f"last {detection_window_hours} hours data is empty. skipiing.")
                continue
            forecast = m.predict(last_hours_data)
            # print(model_history)
            # print(future)
            # print(forecast)
            if local == 1:
                from matplotlib import pyplot as plt
                fig = None
                ax = None
                figsize=(10, 6)
                fig = plt.figure(facecolor='w', figsize=figsize)
                ax = fig.add_subplot(111)
                ax.set_title(query_name + f"{extra_data}")
                fig = m.plot(forecast,ax=ax)
                fig.savefig(f'{query_name}-{i}-forcast.png')
        except Exception as e:
            logger.error(f"Failing builiding forcast - {e}")
            continue

        # find the dataframes having same indices
        forecast_truncated_index = forecast.index.intersection(df.index)
        forecast_truncated = forecast.loc[forecast_truncated_index]

        # Identify the thresholds with some buffer
        upper_buffer = np.max( forecast_truncated['yhat_upper']) * buffer_pct
        if buffer_pct != 0:
            lower_buffer = np.min( forecast_truncated['yhat_lower']) / buffer_pct
        else:
            logger.error(f"Threshold evaluation failed. skipping {query_name}.")
            continue
        
        forecast_truncated_last = forecast_truncated
        forecast_truncated_last["Date"] = pd.to_datetime(forecast_truncated.ds)
        forecast_truncated_last = forecast_truncated_last[forecast_truncated_last['Date'] >= x_hours_ago]

        expected = forecast_truncated_last['yhat']
        
        # expected = expected.apply(lambda x: round(x, 0))
        expected = expected.apply(np.round) 
        expected = expected.astype(int)
        upper_indices = last_hours_data_y[last_hours_data_y['y'] > upper_buffer].index
        lower_indices = last_hours_data_y[last_hours_data_y['y'] < lower_buffer].index
        # print(last_hours_data_y.iloc[upper_indices])

        # Get those points that have crossed the threshold
        anomalies = pd.DataFrame()
        anomalies = anomalies.append(model_history.iloc[upper_indices]) # ------> This has the thresholded values and more important timestamp
        anomalies = anomalies.append(model_history.iloc[lower_indices]) # ------> This has the thresholded values and more important timestamp
        
        if len(anomalies) != 0:
            logger.warning(f"Found {len(anomalies)} anomalies for {query_name} in {extra_data}")
            for index, row in anomalies.iterrows():
                anoamly_counter+=1
                logger.warning(f"[{query_name}] {extra_data} time: {row['ds']} actual: {row['y']}")
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
