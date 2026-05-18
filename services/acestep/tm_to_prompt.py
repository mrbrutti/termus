"""
Convert a termus RenderSpec (the wire shape from internal/acestep/spec.go) into
ACE-Step's GenerationParams.

The Go side does the high-level v3 .tm -> RenderSpec compilation (see
internal/acestep/prompt_compiler.go). This module is the much smaller
Python-side translation: RenderSpec -> GenerationParams. The split exists so
the prompt-building logic stays testable from Go.
"""

from __future__ import annotations

import base64
import os
import tempfile
from dataclasses import dataclass, field
from typing import Any, List, Optional


@dataclass
class RenderSpec:
    """In-process mirror of the Go RenderSpec.

    Field names match the Pydantic RenderRequest fields in server.py so that
    `RenderSpec(**request.dict())` works.
    """

    prompt: str = ""
    tags: List[str] = field(default_factory=list)
    key: str = ""
    tempo: int = 0
    duration_seconds: float = 0.0
    scale: str = ""
    time_signature: str = ""
    seed: int = -1
    reference_audio_b64: str = ""
    section_descriptions: List[str] = field(default_factory=list)
    harmony_chain: str = ""
    motif: str = ""
    inference_steps: int = 8


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def build_caption(spec: RenderSpec) -> str:
    """Build the single 'caption' string that ACE-Step takes as its main prompt.

    Per ACE-Step's documentation the caption is < 512 characters and should
    read like a natural-language description of the desired piece. We compose
    the four authored fields (prompt + tags + motif + harmony) into one
    sentence-ish blob and let the model handle the rest.
    """
    parts: List[str] = []
    if spec.prompt:
        parts.append(spec.prompt.strip())
    if spec.tags:
        parts.append(", ".join(t.strip() for t in spec.tags if t.strip()))
    if spec.motif:
        parts.append(f"motif: {spec.motif.strip()}")
    if spec.harmony_chain:
        parts.append(f"chord progression: {spec.harmony_chain.strip()}")
    if spec.section_descriptions:
        parts.append("structure: " + " -> ".join(d.strip() for d in spec.section_descriptions if d.strip()))
    caption = ". ".join(p for p in parts if p)
    # ACE-Step's caption cap is 512 chars. Truncate defensively.
    if len(caption) > 510:
        caption = caption[:510]
    return caption


def time_signature_to_int(ts: str) -> Optional[int]:
    """Map '4/4' -> 4, '3/4' -> 3, '6/8' -> 6. Returns None if unrecognised."""
    if not ts:
        return None
    s = ts.strip()
    if "/" in s:
        s = s.split("/", 1)[0]
    try:
        return int(s)
    except ValueError:
        return None


def _maybe_save_reference(b64: str) -> Optional[str]:
    """Decode base64 reference audio to a temp file and return its path.

    The tempfile is deliberately *not* cleaned up automatically because
    ACE-Step holds onto the path for the duration of the generation. It will
    be reaped when the process exits. For long-running daemons this is fine
    since reference audio is small.
    """
    if not b64:
        return None
    try:
        data = base64.b64decode(b64)
    except Exception:
        return None
    fd, path = tempfile.mkstemp(prefix="acestep_ref_", suffix=".wav")
    with os.fdopen(fd, "wb") as f:
        f.write(data)
    return path


# ---------------------------------------------------------------------------
# Public entry point
# ---------------------------------------------------------------------------

def compile_spec_to_params(spec: RenderSpec) -> Any:
    """Translate a RenderSpec into an acestep.inference.GenerationParams.

    This import is lazy because the test suite (and --dry-run mode) does not
    have the acestep package installed.
    """
    from acestep.inference import GenerationParams  # type: ignore

    caption = build_caption(spec)
    bpm = spec.tempo if spec.tempo and spec.tempo > 0 else None
    duration = spec.duration_seconds if spec.duration_seconds and spec.duration_seconds > 0 else -1.0
    timesig = time_signature_to_int(spec.time_signature)
    timesig_str = str(timesig) if timesig else ""

    return GenerationParams(
        task_type="text2music",
        caption=caption,
        lyrics="[Instrumental]",
        instrumental=True,
        bpm=bpm,
        keyscale=spec.key or "",
        timesignature=timesig_str,
        duration=duration,
        inference_steps=spec.inference_steps if spec.inference_steps > 0 else 8,
        seed=spec.seed if spec.seed >= 0 else -1,
        reference_audio=_maybe_save_reference(spec.reference_audio_b64),
        # Let the LM enhance metadata when the user only gave us a paragraph.
        thinking=True,
        use_cot_caption=True,
        use_cot_metas=(bpm is None or not spec.key),
        use_cot_language=False,  # instrumental
    )
