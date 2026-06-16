"""
KNOWN LIMITATION: the trained KMeans pipeline has a documented silhouette
score of ~0.14 (heavy cluster overlap, see (Kasaran)_Capstone_K5_B1.ipynb),
and the production schema has no signal for 7 of its 17 features (those
default to fixed values below). Predictions here are demo-quality, not
reliable enough for real segmentation decisions until the model is retrained
on production-shaped features.
"""

from __future__ import annotations

import os
from dataclasses import dataclass
from functools import lru_cache
from typing import Any

import joblib
import numpy as np

MODEL_DIR = os.path.join(os.path.dirname(__file__), "..", "machine-learning")

# Order must match the `features` list used to fit scaler/pca/kmeans in
# (Kasaran)_Capstone_K5_B1.ipynb.
FEATURE_ORDER = [
    "transaction_frequency_monthly",
    "avg_balance",
    "transfer_ratio",
    "payment_ratio",
    "qris_ratio",
    "topup_ratio",
    "investment_activity",
    "loan_usage",
    "credit_card_usage",
    "cardless_usage",
    "poin_xtra_active",
    "login_frequency_weekly",
    "avg_session_duration_min",
    "recent_transaction_days",
    "dominant_activity_enc",
    "frequent_product_enc",
    "customer_type_enc",
]

# Fields the production schema has no signal for at all (no login/session
# tracking, no loan/credit-card/cardless/loyalty flags, no customer tier).
# Defaults are the expected value of the distribution used to generate the
# training dataset, so an unknown user shifts the clustering as little as
# possible instead of being pulled toward an arbitrary class.
DEFAULT_LOAN_USAGE = 0.40
DEFAULT_CREDIT_CARD_USAGE = 0.40
DEFAULT_CARDLESS_USAGE = 0.25
DEFAULT_POIN_XTRA_ACTIVE = 0.40
DEFAULT_LOGIN_FREQUENCY_WEEKLY = 26.5
DEFAULT_AVG_SESSION_DURATION_MIN = 8.0
DEFAULT_AVG_BALANCE = 50_250_000.0
DEFAULT_RECENT_TRANSACTION_DAYS = 15.0

# LabelEncoder assigns codes alphabetically over the categories seen during
# training. frequent_product/customer_type have no equivalent in the
# production schema, so unknown users default to the most common category.
DOMINANT_ACTIVITY_CODES = {
    "bills_topup": 0,
    "low_activity": 1,
    "qris_payment": 2,
    "transfer": 3,
    "wealth_investment": 4,
}
DEFAULT_FREQUENT_PRODUCT_ENC = 4  # savings_deposito (most common, p=0.30)
DEFAULT_CUSTOMER_TYPE_ENC = 2  # nasabah reguler (most common, p=0.70)

SEGMENT_DESCRIPTIONS = {
    "investor": "Nasabah dengan sinyal ketertarikan investasi dan saldo rata-rata tinggi.",
    "low_activity": "Nasabah dengan frekuensi transaksi dan penggunaan fitur yang masih rendah.",
    "digital_spender": "Nasabah yang aktif melakukan top-up dan transaksi digital/e-wallet.",
    "bill_payer": "Nasabah yang rutin melakukan pembayaran tagihan dan cocok untuk auto-debit.",
}


@dataclass
class SegmentPrediction:
    customer_id: str
    segment_name: str
    description: str
    confidence: float

    def to_api_payload(self) -> dict[str, Any]:
        return {
            "customer_id": self.customer_id,
            "segment_name": self.segment_name,
            "description": self.description,
            "confidence": self.confidence,
        }


@lru_cache(maxsize=1)
def _load_pipeline():
    return (
        joblib.load(os.path.join(MODEL_DIR, "scaler.pkl")),
        joblib.load(os.path.join(MODEL_DIR, "pca_model.pkl")),
        joblib.load(os.path.join(MODEL_DIR, "kmeans_model.pkl")),
        joblib.load(os.path.join(MODEL_DIR, "segment_map.pkl")),
    )


def predict_customer_segment(user: dict[str, Any], activity: dict[str, Any]) -> SegmentPrediction:
    customer_id = str(user.get("customer_id") or "")
    data = activity.get("data") or {}
    transactions = data.get("transactions") or []
    frequently_used = data.get("frequently_used_features") or []

    features = _extract_features(transactions, frequently_used)
    segment_name, confidence = _predict_from_model(features)

    return SegmentPrediction(
        customer_id=customer_id,
        segment_name=segment_name,
        description=_build_description(segment_name, features),
        confidence=confidence,
    )


def _extract_features(
    transactions: list[dict[str, Any]],
    frequently_used: list[dict[str, Any]],
) -> dict[str, Any]:
    feature_counts: dict[str, int] = {}
    for item in frequently_used:
        feature = _normalize_feature(item.get("feature"))
        if feature:
            feature_counts[feature] = feature_counts.get(feature, 0) + _as_int(item.get("usage_count"))

    transaction_type_counts: dict[str, int] = {}
    for transaction in transactions:
        transaction_type = _normalize_feature(transaction.get("type"))
        if transaction_type:
            transaction_type_counts[transaction_type] = transaction_type_counts.get(transaction_type, 0) + 1

    combined_counts = dict(feature_counts)
    for key, value in transaction_type_counts.items():
        combined_counts[key] = combined_counts.get(key, 0) + value

    total_activity = sum(combined_counts.values())
    transfer_count = _count_matching(combined_counts, "transfer")
    payment_count = _count_matching(combined_counts, "bill", "payment", "tagihan", "auto_debit")
    qris_count = _count_matching(combined_counts, "qris")
    topup_count = _count_matching(combined_counts, "topup", "e_wallet", "wallet", "octo_pay")
    investment_count = _count_matching(combined_counts, "investment", "reksa", "deposito")

    dominant_feature = ""
    if combined_counts:
        dominant_feature = sorted(combined_counts.items(), key=lambda item: (-item[1], item[0]))[0][0]

    return {
        "transaction_frequency_monthly": len(transactions),
        "avg_balance": DEFAULT_AVG_BALANCE,
        "transfer_ratio": _ratio(transfer_count, total_activity),
        "payment_ratio": _ratio(payment_count, total_activity),
        "qris_ratio": _ratio(qris_count, total_activity),
        "topup_ratio": _ratio(topup_count, total_activity),
        "investment_activity": 1 if investment_count > 0 else 0,
        "loan_usage": DEFAULT_LOAN_USAGE,
        "credit_card_usage": DEFAULT_CREDIT_CARD_USAGE,
        "cardless_usage": DEFAULT_CARDLESS_USAGE,
        "poin_xtra_active": DEFAULT_POIN_XTRA_ACTIVE,
        "login_frequency_weekly": DEFAULT_LOGIN_FREQUENCY_WEEKLY,
        "avg_session_duration_min": DEFAULT_AVG_SESSION_DURATION_MIN,
        "recent_transaction_days": _recent_transaction_days(transactions),
        "dominant_activity_enc": _dominant_activity_code(dominant_feature),
        "frequent_product_enc": DEFAULT_FREQUENT_PRODUCT_ENC,
        "customer_type_enc": DEFAULT_CUSTOMER_TYPE_ENC,
        "dominant_feature": dominant_feature,
        "total_activity": total_activity,
    }


def _predict_from_model(features: dict[str, Any]) -> tuple[str, float]:
    scaler, pca, kmeans, segment_map = _load_pipeline()

    vector = np.array([[features[name] for name in FEATURE_ORDER]])
    pca_vector = pca.transform(scaler.transform(vector))

    cluster = int(kmeans.predict(pca_vector)[0])
    segment_name = segment_map[cluster]

    distances = kmeans.transform(pca_vector)[0]
    probabilities = np.exp(-distances) / np.exp(-distances).sum()
    confidence = float(probabilities[cluster])

    return segment_name, round(confidence, 2)


def _build_description(segment_name: str, features: dict[str, Any]) -> str:
    dominant_feature = features.get("dominant_feature") or "aktivitas mobile banking"
    total_activity = int(features["total_activity"])
    base_description = SEGMENT_DESCRIPTIONS.get(segment_name, "Segmen nasabah berdasarkan model clustering.")
    return f"{base_description} Fitur dominan: {dominant_feature}. Total aktivitas terdeteksi: {total_activity}."


def _recent_transaction_days(transactions: list[dict[str, Any]]) -> float:
    if not transactions:
        return DEFAULT_RECENT_TRANSACTION_DAYS

    from datetime import datetime, timezone

    latest = transactions[0].get("created_at")
    if not latest:
        return DEFAULT_RECENT_TRANSACTION_DAYS

    try:
        created_at = datetime.fromisoformat(str(latest))
        if created_at.tzinfo is None:
            created_at = created_at.replace(tzinfo=timezone.utc)
        days = (datetime.now(timezone.utc) - created_at).days
        return float(max(days, 0))
    except ValueError:
        return DEFAULT_RECENT_TRANSACTION_DAYS


def _dominant_activity_code(dominant_feature: str) -> int:
    if not dominant_feature:
        return DOMINANT_ACTIVITY_CODES["low_activity"]

    if any(needle in dominant_feature for needle in ("investment", "reksa", "deposito")):
        return DOMINANT_ACTIVITY_CODES["wealth_investment"]
    if "qris" in dominant_feature:
        return DOMINANT_ACTIVITY_CODES["qris_payment"]
    if any(needle in dominant_feature for needle in ("bill", "payment", "tagihan", "auto_debit", "topup", "e_wallet", "wallet", "octo_pay")):
        return DOMINANT_ACTIVITY_CODES["bills_topup"]
    if "transfer" in dominant_feature:
        return DOMINANT_ACTIVITY_CODES["transfer"]

    return DOMINANT_ACTIVITY_CODES["low_activity"]


def _ratio(matching: int, total: int) -> float:
    if total <= 0:
        return 0.0
    return round(matching / total, 4)


def _count_matching(counts: dict[str, int], *needles: str) -> int:
    total = 0
    for key, value in counts.items():
        if any(needle in key for needle in needles):
            total += value
    return total


def _normalize_feature(value: Any) -> str:
    return str(value or "").strip().lower().replace("-", "_").replace(" ", "_")


def _as_int(value: Any) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return 0
