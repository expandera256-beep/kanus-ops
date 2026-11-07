FROM python:3.12-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt && useradd --uid 10001 --create-home app
COPY *.py ./
USER 10001
ENV PORT=8080
CMD ["sh","-c","exec gunicorn --bind 0.0.0.0:${PORT} --workers 1 --worker-class gthread --threads 32 --timeout 630 --graceful-timeout 30 --keep-alive 5 --access-logfile /dev/null \"${US_APP_FACTORY:-runtime:gateway_app()}\""]
