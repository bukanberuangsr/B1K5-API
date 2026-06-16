FROM python:3.12-slim
WORKDIR /app

COPY assets/requirements.txt assets/requirements.txt
RUN pip install --no-cache-dir -r assets/requirements.txt

COPY assets/ml_service.py assets/b1k5_capstone_model.py assets/
COPY machine-learning/ machine-learning/

WORKDIR /app/assets
EXPOSE 8000
CMD ["uvicorn", "ml_service:app", "--host", "0.0.0.0", "--port", "8000"]
