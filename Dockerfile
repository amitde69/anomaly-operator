FROM python:3

WORKDIR /app
COPY req.txt .
RUN apt-get -y install libc-dev
RUN pip install -r req.txt
COPY preprocess.py .
COPY train.py .
COPY main.py .
CMD python main.py