from train import detect_cycle
import yaml, os, time
from yaml.loader import SafeLoader
import logging

log_level = os.getenv("LOG_LEVEL", "ERROR")
logging.basicConfig(format='%(levelname)s %(message)s', level=log_level)
# Open the file and load the file
config = None
with open('config.yaml') as f:
    config = yaml.load(f, Loader=SafeLoader)

logging.info(f"Got config {config}")
interval_mins = config['interval_mins']
interval = interval_mins * 60 * 60
while True:
    detect_cycle(config)
    time.sleep(interval)