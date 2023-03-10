from train import detect_cycle
import yaml, os, time
from yaml.loader import SafeLoader
import logger as logging
from prometheus_client import start_http_server
import prometheus_client
prometheus_client.REGISTRY.unregister(prometheus_client.GC_COLLECTOR)
prometheus_client.REGISTRY.unregister(prometheus_client.PLATFORM_COLLECTOR)
prometheus_client.REGISTRY.unregister(prometheus_client.PROCESS_COLLECTOR)

# log_level = os.getenv("LOG_LEVEL", "ERROR")
# logging.basicConfig(format='[%(asctime)s] %(levelname)s %(message)s', level=log_level)


# Open the file and load the file
config = None
prometheus_port = 9090
with open('config.yaml') as f:
    config = yaml.load(f, Loader=SafeLoader)
logger = logging.load()
logger.info(f"Got config {config}")
logger.info(f"Starting prometheus endpoint at /metrics on port {prometheus_port}...")
start_http_server(prometheus_port)
interval_mins = config['interval_mins']
interval = interval_mins * 60
logger.info(f"Starting detector loop...")
while True:
    detect_cycle(config, logger)
    time.sleep(interval)