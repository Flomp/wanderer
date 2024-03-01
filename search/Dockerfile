FROM python:3.10-slim
WORKDIR /app

COPY init.py /app
COPY cities500.tar.gz /app

RUN tar -xvf cities500.tar.gz
RUN pip install meilisearch==0.30.0

ENV PUBLIC_MEILISEARCH_URL=http://localhost:7700
ENV MEILISEARCH_MASTER_KEY=

CMD ["python", "./init.py"]