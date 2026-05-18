"""HTTP-level smoke tests for the ACE-Step service.

These tests do NOT exercise the actual ACE-Step model. They verify:
  - The FastAPI app boots in mock mode.
  - /health reports loaded=true after lifespan.
  - /render accepts the wire shape and returns WAV bytes.
  - WAV bytes have a valid RIFF header.
  - tm_to_prompt.build_caption composes prompt + tags + motif sensibly.
  - tm_to_prompt.time_signature_to_int handles the documented forms.

Run with:
    python -m pytest test_server.py -v
"""

from __future__ import annotations

import io
import wave

import pytest
from fastapi.testclient import TestClient

import server
import tm_to_prompt


@pytest.fixture()
def client():
    # Force mock mode so the lifespan doesn't try to load the real model.
    server.STATE.mock_mode = True
    server.STATE.loaded = False
    server.STATE.error = None
    server.STATE.backend = "unknown"
    with TestClient(server.app) as c:
        yield c


def test_health_reports_mock_mode(client):
    r = client.get("/health")
    assert r.status_code == 200
    body = r.json()
    assert body["loaded"] is True
    assert body["mock_mode"] is True
    assert body["backend"] == "mock"


def test_render_returns_wav(client):
    payload = {
        "prompt": "warm lo-fi rhodes in a quiet bookstore on a rainy night",
        "tags": ["lofi", "rhodes", "downtempo"],
        "key": "Cmin",
        "tempo": 78,
        "duration_seconds": 30,
        "scale": "minor",
        "time_signature": "4/4",
        "seed": 71003,
    }
    r = client.post("/render", json=payload)
    assert r.status_code == 200, r.text
    assert r.headers["content-type"] == "audio/wav"
    assert "X-Generation-Time-Seconds" in r.headers
    assert r.headers["X-Backend"] == "mock"
    assert r.headers["X-Seed"] == "71003"

    # Must be a valid RIFF/WAVE.
    body = r.content
    assert body[:4] == b"RIFF"
    assert body[8:12] == b"WAVE"

    with wave.open(io.BytesIO(body), "rb") as wf:
        assert wf.getframerate() == 48000
        assert wf.getnchannels() == 1
        assert wf.getsampwidth() == 2
        assert wf.getnframes() > 0


def test_render_rejects_unloaded(client):
    # Flip loaded=false manually to simulate a model load failure.
    server.STATE.loaded = False
    server.STATE.error = "intentional test failure"
    try:
        r = client.post("/render", json={"prompt": "x", "tags": []})
        assert r.status_code == 503
        assert "not loaded" in r.json()["detail"]
    finally:
        server.STATE.loaded = True
        server.STATE.error = None


def test_build_caption_composes_all_fields():
    spec = tm_to_prompt.RenderSpec(
        prompt="warm rhodes",
        tags=["lofi", "downtempo"],
        motif="5 7 5 3 stepwise minor",
        harmony_chain="Am7 Fmaj7",
        section_descriptions=["intro", "head"],
    )
    caption = tm_to_prompt.build_caption(spec)
    assert "warm rhodes" in caption
    assert "lofi" in caption
    assert "5 7 5 3" in caption
    assert "Am7" in caption
    assert "intro" in caption


def test_build_caption_truncates():
    spec = tm_to_prompt.RenderSpec(
        prompt="x" * 600,
        tags=[],
    )
    caption = tm_to_prompt.build_caption(spec)
    assert len(caption) <= 510


def test_time_signature_to_int():
    assert tm_to_prompt.time_signature_to_int("4/4") == 4
    assert tm_to_prompt.time_signature_to_int("3/4") == 3
    assert tm_to_prompt.time_signature_to_int("6/8") == 6
    assert tm_to_prompt.time_signature_to_int("") is None
    assert tm_to_prompt.time_signature_to_int("garbage") is None
