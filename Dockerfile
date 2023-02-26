FROM python:3

WORKDIR /app
COPY req.txt .
RUN apt-get -y install libc-dev
RUN pip install -r req.txt
COPY . .
CMD python main.py