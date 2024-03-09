FROM python:3.10-slim
WORKDIR /app

COPY bootstrap.py /app
COPY cities500.tar.gz /app

RUN tar -xvf cities500.tar.gz
RUN pip install meilisearch==0.30.0

ENV MEILI_URL=http://localhost:7700
ENV MEILI_MASTER_KEY=

CMD ["python", "-u", "./bootstrap.py"]