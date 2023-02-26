FROM python:3

WORKDIR /app
COPY . .
RUN apt-get -y install libc-dev
RUN pip install -r req.txt
CMD python main.py