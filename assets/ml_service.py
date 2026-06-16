from __future__ import annotations

from typing import Any

from fastapi import FastAPI
from pydantic import BaseModel

from b1k5_capstone_model import predict_customer_segment

app = FastAPI(title="B1K5 Segmentation ML Service")


class PredictRequest(BaseModel):
    user: dict[str, Any]
    activity: dict[str, Any]


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/predict")
def predict(request: PredictRequest) -> dict[str, Any]:
    prediction = predict_customer_segment(request.user, request.activity)
    return prediction.to_api_payload()
