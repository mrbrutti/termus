"""
ACE-Step inference daemon for termus SP21.

This is a thin FastAPI shim in front of the ACE-Step text-to-music model.
The Go side (cmd/termus-stream) talks to this service over HTTP and gets
back WAV bytes.

UNTESTED: actual ACE-Step inference has not been exercised in this PR.
Model weights are several GB and require a separate install step
(see install.sh). The HTTP surface is covered by test_server.py with
the model mocked out.

Endpoints:
  POST /render   - generate one track, returns audio/wav bytes
  GET  /health   - report whether the model is loaded

Default port: 7790
"""

from __future__ import annotations

import argparse
import io
import logging
import os
import sys
import time
import wave
from contextlib import asynccontextmanager
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse, Response
from pydantic import BaseModel, Field

from tm_to_prompt import RenderSpec, compile_spec_to_params  # type: ignore

logger = logging.getLogger("acestep.server")
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s :: %(message)s",
    stream=sys.stderr,
)


# ---------------------------------------------------------------------------
# Model state - the heavy ACE-Step handlers live in module scope so the model
# is loaded once at startup, not per-request.
# ---------------------------------------------------------------------------

@dataclass
class ModelState:
    dit_handler: Any = None   # acestep.handler.AceStepHandler
    llm_handler: Any = None   # acestep.llm_inference.LLMHandler
    loaded: bool = False
    backend: str = "unknown"  # "mps", "mlx", "cuda", "cpu", or "mock"
    model_name: str = "acestep-v15-turbo"
    lm_model_name: str = "acestep-5Hz-lm-1.7B"
    error: Optional[str] = None
    load_time_seconds: float = 0.0
    mock_mode: bool = False


STATE = ModelState()


def _detect_backend() -> str:
    """Pick the best backend on this host. MLX-LM is preferred on Apple Silicon
    if the mlx-lm package is importable; otherwise fall back to torch MPS;
    otherwise CUDA; otherwise CPU."""
    try:
        import torch  # type: ignore
    except ImportError:
        return "cpu"

    # MPS = Apple Silicon GPU
    if hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
        try:
            import mlx_lm  # type: ignore  # noqa: F401
            return "mlx"
        except ImportError:
            return "mps"
    if torch.cuda.is_available():
        return "cuda"
    return "cpu"


def _load_real_model() -> None:
    """Load the actual ACE-Step model. May take 30s+ and several GB of RAM/VRAM.

    Import paths follow the Apple Silicon fork:
      https://github.com/clockworksquirrel/ace-step-apple-silicon
    The upstream repo (ace-step/ACE-Step-1.5) exposes the same module layout.
    """
    # NOTE: these imports are inside the function so the server still starts
    # in mock mode if the acestep package is not installed.
    from acestep.handler import AceStepHandler           # type: ignore
    from acestep.llm_inference import LLMHandler         # type: ignore

    backend = _detect_backend()
    STATE.backend = backend
    logger.info(f"loading ACE-Step DiT model={STATE.model_name} backend={backend}")
    t0 = time.time()

    dit = AceStepHandler()
    # The handler exposes initialize_service() in both forks. Argument names
    # have shifted across versions; we pass model_name + checkpoint dir if
    # supported, else fall back to the no-arg form.
    try:
        dit.initialize_service(
            dit_model_name=STATE.model_name,
            use_mlx=(backend == "mlx"),
        )
    except TypeError:
        dit.initialize_service()

    llm = LLMHandler()
    try:
        llm.initialize_lm_service(
            lm_model_name=STATE.lm_model_name,
            use_mlx=(backend == "mlx"),
        )
    except TypeError:
        llm.initialize_lm_service()

    STATE.dit_handler = dit
    STATE.llm_handler = llm
    STATE.loaded = True
    STATE.load_time_seconds = time.time() - t0
    logger.info(f"model loaded in {STATE.load_time_seconds:.1f}s backend={backend}")


def _load_mock_model() -> None:
    """Pretend to be loaded. Used for --dry-run and for the HTTP test suite."""
    STATE.backend = "mock"
    STATE.loaded = True
    STATE.mock_mode = True
    STATE.load_time_seconds = 0.0
    logger.info("mock mode: skipping real ACE-Step model load")


@asynccontextmanager
async def _lifespan(app: FastAPI):
    if STATE.mock_mode:
        _load_mock_model()
    else:
        try:
            _load_real_model()
        except Exception as exc:  # pragma: no cover  -- real model only
            STATE.error = f"{type(exc).__name__}: {exc}"
            logger.error(f"failed to load model: {STATE.error}")
            # Continue anyway so /health can report the error.
    yield


app = FastAPI(title="termus-acestep", lifespan=_lifespan)


# ---------------------------------------------------------------------------
# Request / response models
# ---------------------------------------------------------------------------

class RenderRequest(BaseModel):
    """Wire shape that matches internal/acestep/spec.go on the Go side.

    Field names are snake_case for both consistency with the Python ecosystem
    and to match the Go json tags.
    """

    prompt: str = Field(..., description="Natural-language style description.")
    tags: List[str] = Field(default_factory=list, description="Rank-ordered descriptors; genre first.")
    key: str = Field("", description="Musical key, e.g. 'Cmin', 'C major'.")
    tempo: int = Field(0, description="BPM. 0 = let the model choose.")
    duration_seconds: float = Field(0.0, description="Target length in seconds. 0 = let the model choose.")
    scale: str = Field("", description="'minor', 'major', 'dorian', etc.")
    time_signature: str = Field("", description="'4/4', '3/4'.")
    seed: int = Field(-1, description="Reproducibility seed. -1 = random.")
    reference_audio_b64: str = Field("", description="Optional base64-encoded WAV/MP3 reference audio.")

    # Per-section conditioning, optional. Concatenated into the prompt.
    section_descriptions: List[str] = Field(default_factory=list)
    harmony_chain: str = Field("", description="Chord sequence across all sections, joined.")
    motif: str = Field("", description="Natural-language motif description.")
    inference_steps: int = Field(8, description="Diffusion steps. 8 is turbo-default.")


class HealthResponse(BaseModel):
    loaded: bool
    backend: str
    model_name: str
    lm_model_name: str
    mock_mode: bool
    error: Optional[str] = None
    load_time_seconds: float = 0.0


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------

@app.get("/health", response_model=HealthResponse)
async def health() -> HealthResponse:
    return HealthResponse(
        loaded=STATE.loaded,
        backend=STATE.backend,
        model_name=STATE.model_name,
        lm_model_name=STATE.lm_model_name,
        mock_mode=STATE.mock_mode,
        error=STATE.error,
        load_time_seconds=STATE.load_time_seconds,
    )


@app.post("/render")
async def render(req: RenderRequest, http_request: Request) -> Response:
    if not STATE.loaded:
        raise HTTPException(status_code=503, detail=f"model not loaded: {STATE.error or 'unknown'}")

    spec = RenderSpec(
        prompt=req.prompt,
        tags=list(req.tags),
        key=req.key,
        tempo=req.tempo,
        duration_seconds=req.duration_seconds,
        scale=req.scale,
        time_signature=req.time_signature,
        seed=req.seed,
        reference_audio_b64=req.reference_audio_b64,
        section_descriptions=list(req.section_descriptions),
        harmony_chain=req.harmony_chain,
        motif=req.motif,
        inference_steps=req.inference_steps,
    )

    t0 = time.time()
    try:
        wav_bytes = _generate(spec)
    except HTTPException:
        raise
    except Exception as exc:
        logger.exception("render failed")
        raise HTTPException(status_code=500, detail=f"{type(exc).__name__}: {exc}")
    elapsed = time.time() - t0
    logger.info(
        f"render done duration={spec.duration_seconds:.1f}s seed={spec.seed} "
        f"bytes={len(wav_bytes)} elapsed={elapsed:.1f}s"
    )

    return Response(
        content=wav_bytes,
        media_type="audio/wav",
        headers={
            "X-Generation-Time-Seconds": f"{elapsed:.3f}",
            "X-Backend": STATE.backend,
            "X-Seed": str(spec.seed),
        },
    )


# ---------------------------------------------------------------------------
# Generation
# ---------------------------------------------------------------------------

def _generate(spec: RenderSpec) -> bytes:
    """Run one ACE-Step inference. Returns WAV bytes.

    In mock mode we synthesize a short silence-with-clicks WAV so the wire
    contract still works end-to-end for the HTTP tests and for `--dry-run`.
    """
    if STATE.mock_mode:
        return _mock_wav(spec)

    # Real path. Uses acestep.inference.generate_music as documented in the
    # Apple Silicon fork inference.py.
    from acestep.inference import (   # type: ignore
        GenerationParams,
        GenerationConfig,
        generate_music,
    )
    import tempfile

    params: GenerationParams = compile_spec_to_params(spec)
    config = GenerationConfig(
        batch_size=1,
        allow_lm_batch=False,
        use_random_seed=(spec.seed < 0),
        seeds=[spec.seed] if spec.seed >= 0 else None,
        audio_format="wav",
    )

    with tempfile.TemporaryDirectory() as tmpdir:
        result = generate_music(
            dit_handler=STATE.dit_handler,
            llm_handler=STATE.llm_handler,
            params=params,
            config=config,
            save_dir=tmpdir,
        )
        if not result.success:
            raise HTTPException(status_code=500, detail=f"generation failed: {result.error or 'unknown'}")
        if not result.audios:
            raise HTTPException(status_code=500, detail="generation succeeded but no audios returned")
        # result.audios is List[Dict] - each dict contains at minimum a path.
        # The exact key has shifted between forks; we accept several.
        first: Dict[str, Any] = result.audios[0]
        path = (
            first.get("path")
            or first.get("audio_path")
            or first.get("filepath")
            or first.get("file")
        )
        if not path or not os.path.exists(path):
            raise HTTPException(status_code=500, detail=f"audio path missing in result: {first!r}")
        with open(path, "rb") as f:
            return f.read()


def _mock_wav(spec: RenderSpec) -> bytes:
    """Return a tiny silent WAV (~0.5s, mono, 48 kHz). Used in mock mode.

    Real ACE-Step output is 48 kHz stereo; the mock matches the sample rate
    but trims duration for fast tests.
    """
    sample_rate = 48000
    seconds = 0.5
    n_samples = int(sample_rate * seconds)
    buf = io.BytesIO()
    with wave.open(buf, "wb") as wf:
        wf.setnchannels(1)
        wf.setsampwidth(2)
        wf.setframerate(sample_rate)
        # All zeros - silent. Two bytes per mono sample.
        wf.writeframes(b"\x00\x00" * n_samples)
    return buf.getvalue()


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def _parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser()
    p.add_argument("--host", default="127.0.0.1")
    p.add_argument("--port", type=int, default=7790)
    p.add_argument(
        "--dry-run",
        action="store_true",
        help="Don't actually load the model; serve a mock WAV for testing.",
    )
    p.add_argument(
        "--model",
        default="acestep-v15-turbo",
        help="DiT model name. Default is the 2B turbo model (smallest, fastest).",
    )
    p.add_argument(
        "--lm-model",
        default="acestep-5Hz-lm-1.7B",
        help="LM model name. Default is the 1.7B model that ships with the unified repo.",
    )
    return p.parse_args()


def main() -> None:
    args = _parse_args()
    STATE.mock_mode = args.dry_run
    STATE.model_name = args.model
    STATE.lm_model_name = args.lm_model

    import uvicorn  # type: ignore
    uvicorn.run(app, host=args.host, port=args.port, workers=1, log_level="info")


if __name__ == "__main__":
    main()
