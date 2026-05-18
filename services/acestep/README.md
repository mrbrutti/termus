# termus ACE-Step service

Local FastAPI daemon that exposes ACE-Step 1.5 (a text-to-music diffusion
model) to the Go side of termus. The Go side (`internal/acestep` + `cmd/termus-stream`) talks to this service over HTTP and gets back WAV bytes.

## What this is and isn't

- Runs **locally**. No cloud, no API key.
- Wraps the upstream Python package; this directory is just a thin HTTP shim.
- Uses the **2B turbo** model by default - the smallest and fastest variant.
- On Apple Silicon, uses **MPS** for the DiT and **MLX-LM** for the language model when those are available.

## Honest status

Everything in this directory was written from the published source of two repos:

- https://github.com/ace-step/ACE-Step-1.5
- https://github.com/clockworksquirrel/ace-step-apple-silicon (Apple Silicon fork; primary reference)

The HTTP surface (`/health`, `/render`) and the wire shape (`RenderRequest` matching the Go `RenderSpec`) are **covered by `test_server.py`** with the ACE-Step model mocked out. The actual inference path has **not** been exercised in this PR because the model weights are several GB. The user must run `install.sh` and verify end-to-end after merge.

The exact `acestep.inference` import path is taken verbatim from the upstream `inference.py` (`generate_music`, `GenerationParams`, `GenerationConfig`). The handler bootstrap (`initialize_service`, `initialize_lm_service`) has shifted argument names across ACE-Step versions; `server.py` tries the kwargs form first and falls back to the no-arg form (which lets the handler use its own defaults).

## Install

```bash
cd services/acestep
./install.sh
```

This will:

1. Create a Python 3.11 venv in `services/acestep/venv`.
2. Install FastAPI, uvicorn, pydantic.
3. Clone `ace-step-apple-silicon` into `services/acestep/vendor/`.
4. `pip install -e` that source tree.
5. Pre-download the 2B turbo model (5-10 GB on first run).

Override the source repo if you want the upstream CUDA fork:

```bash
ACESTEP_REPO_URL=https://github.com/ace-step/ACE-Step-1.5.git ./install.sh
```

Override the model:

```bash
ACESTEP_MODEL=acestep-v15-base ./install.sh
```

## Run

```bash
cd services/acestep
source venv/bin/activate
python server.py                  # full model, ~30 s warm-up
python server.py --dry-run        # mock mode for HTTP testing
python server.py --port 7790      # default
```

## Endpoints

### `GET /health`

```json
{
  "loaded": true,
  "backend": "mlx",
  "model_name": "acestep-v15-turbo",
  "lm_model_name": "acestep-5Hz-lm-1.7B",
  "mock_mode": false,
  "error": null,
  "load_time_seconds": 28.4
}
```

### `POST /render`

```json
{
  "prompt": "warm lo-fi rhodes in a quiet bookstore on a rainy night",
  "tags": ["lofi", "rhodes", "downtempo", "ambient"],
  "key": "Cmin",
  "tempo": 78,
  "duration_seconds": 90,
  "scale": "minor",
  "time_signature": "4/4",
  "seed": 71003,
  "section_descriptions": [
    "soft intro with brushed kick",
    "rhodes states the motif over Am7 → Fmaj7",
    "fade-out with vinyl crackle"
  ],
  "harmony_chain": "Am7 Fmaj7 Dm7 G7sus",
  "motif": "wandering minor melody, scale degrees 5 7 5 3",
  "inference_steps": 8
}
```

Returns `audio/wav` bytes. Headers include `X-Generation-Time-Seconds`, `X-Backend`, `X-Seed`.

## Test

```bash
source venv/bin/activate
python -m pytest test_server.py -v
```

The test exercises the HTTP layer with the model mocked. It does **not** verify actual music generation.
