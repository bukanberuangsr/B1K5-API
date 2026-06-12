from __future__ import annotations

from dataclasses import dataclass
from typing import Any


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


def predict_customer_segment(user: dict[str, Any], activity: dict[str, Any]) -> SegmentPrediction:
    """Predict a customer segment from the API activity shape.

    This module is intentionally dependency-free so it can run alongside the Go API.
    Replace this function with the exported sklearn pipeline once scaler/PCA/KMeans
    artifacts are available.
    """

    customer_id = str(user.get("customer_id") or "")
    data = activity.get("data") or {}
    transactions = data.get("transactions") or []
    frequently_used = data.get("frequently_used_features") or []

    features = _extract_features(transactions, frequently_used)
    segment_name, confidence = _predict_from_features(features)

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
    transaction_count = len(transactions)
    usage_count = sum(_as_int(item.get("usage_count")) for item in frequently_used)
    total_activity = transaction_count + usage_count

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

    dominant_feature = ""
    if combined_counts:
        dominant_feature = sorted(combined_counts.items(), key=lambda item: (-item[1], item[0]))[0][0]

    return {
        "transaction_count": transaction_count,
        "usage_count": usage_count,
        "total_activity": total_activity,
        "dominant_feature": dominant_feature,
        "investment_count": _count_matching(combined_counts, "investment", "reksa", "deposito"),
        "bill_count": _count_matching(combined_counts, "bill", "payment", "tagihan", "auto_debit"),
        "digital_count": _count_matching(combined_counts, "topup", "e_wallet", "wallet", "qris", "octo_pay"),
        "transfer_count": _count_matching(combined_counts, "transfer"),
    }


def _predict_from_features(features: dict[str, Any]) -> tuple[str, float]:
    total_activity = int(features["total_activity"])
    investment_count = int(features["investment_count"])
    bill_count = int(features["bill_count"])
    digital_count = int(features["digital_count"])
    transfer_count = int(features["transfer_count"])

    if total_activity <= 1:
        return "low_activity", 0.82

    if investment_count > 0:
        return "investor", _confidence(total_activity, investment_count)

    if digital_count > 0 and digital_count >= bill_count:
        return "digital_spender", _confidence(total_activity, digital_count)

    if bill_count > 0:
        return "bill_payer", _confidence(total_activity, bill_count)

    if transfer_count >= 3:
        return "digital_spender", _confidence(total_activity, transfer_count)

    return "low_activity", 0.76


def _build_description(segment_name: str, features: dict[str, Any]) -> str:
    dominant_feature = features.get("dominant_feature") or "aktivitas mobile banking"
    total_activity = int(features["total_activity"])

    if segment_name == "investor":
        return (
            f"Profil nasabah menunjukkan sinyal aktivitas investasi melalui {dominant_feature}. "
            f"Total aktivitas terdeteksi: {total_activity}."
        )

    if segment_name == "digital_spender":
        return (
            f"Nasabah aktif pada transaksi digital seperti top-up, e-wallet, QRIS, atau transfer. "
            f"Fitur dominan: {dominant_feature}. Total aktivitas: {total_activity}."
        )

    if segment_name == "bill_payer":
        return (
            f"Nasabah menunjukkan pola pembayaran tagihan atau payment rutin. "
            f"Fitur dominan: {dominant_feature}. Total aktivitas: {total_activity}."
        )

    return (
        f"Aktivitas digital banking nasabah masih rendah atau belum memiliki pola kuat. "
        f"Total aktivitas terdeteksi: {total_activity}."
    )


def _confidence(total_activity: int, matching_activity: int) -> float:
    if total_activity <= 0:
        return 0.7

    signal_ratio = matching_activity / total_activity
    confidence = 0.72 + min(signal_ratio, 1) * 0.2 + min(total_activity, 10) * 0.003
    return round(min(confidence, 0.95), 2)


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
