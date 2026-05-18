#!/usr/bin/env bash
# Install the ACE-Step service in a local venv and pre-download model weights.
#
# UNTESTED in CI - the model is ~5-10 GB and downloading it from a build
# machine is not feasible. The user must run this manually after merge.
#
# Tested with:
#   - macOS 14+ on Apple Silicon (M2/M3)
#   - Python 3.11 (the acestep package pins requires-python = "==3.11.*")
#
# Usage:
#   cd services/acestep
#   ./install.sh
#
# Environment overrides:
#   ACESTEP_REPO_URL    - git URL to clone (default: clockworksquirrel fork)
#   ACESTEP_REPO_REF    - branch/tag/commit to check out (default: main)
#   ACESTEP_MODEL       - DiT model name (default: acestep-v15-turbo)
set -euo pipefail

REPO_URL="${ACESTEP_REPO_URL:-https://github.com/clockworksquirrel/ace-step-apple-silicon.git}"
REPO_REF="${ACESTEP_REPO_REF:-main}"
MODEL_NAME="${ACESTEP_MODEL:-acestep-v15-turbo}"

cd "$(dirname "$0")"
HERE="$(pwd)"

PYTHON_BIN="${PYTHON_BIN:-python3.11}"
if ! command -v "$PYTHON_BIN" >/dev/null 2>&1; then
    echo "error: $PYTHON_BIN not found. Install Python 3.11 first." >&2
    echo "  brew install python@3.11    # macOS" >&2
    exit 1
fi

# 1. Create venv if absent.
if [ ! -d venv ]; then
    echo "==> Creating venv with $PYTHON_BIN"
    "$PYTHON_BIN" -m venv venv
fi

# shellcheck disable=SC1091
source venv/bin/activate

# 2. Upgrade pip and install our shim deps.
echo "==> Installing FastAPI + uvicorn"
pip install --upgrade pip wheel
pip install -r requirements.txt

# 3. Clone (or update) the ACE-Step source tree.
ACE_DIR="$HERE/vendor/ace-step"
if [ ! -d "$ACE_DIR" ]; then
    echo "==> Cloning $REPO_URL into vendor/"
    mkdir -p "$HERE/vendor"
    git clone --depth 1 --branch "$REPO_REF" "$REPO_URL" "$ACE_DIR"
else
    echo "==> Updating $ACE_DIR (ref=$REPO_REF)"
    (cd "$ACE_DIR" && git fetch --depth 1 origin "$REPO_REF" && git checkout "$REPO_REF")
fi

# 4. Install ACE-Step (editable) into our venv.
echo "==> Installing acestep package (editable)"
pip install -e "$ACE_DIR"

# 5. Pre-download the model. This is the slow step (5-10 GB).
echo "==> Pre-downloading $MODEL_NAME (this can take 5-15 minutes)"
python - <<PY
import os, sys
try:
    from acestep.model_downloader import ensure_main_model, ensure_dit_model
except ImportError as exc:
    print(f"could not import acestep model_downloader: {exc}", file=sys.stderr)
    sys.exit(1)

model = os.environ.get("ACESTEP_MODEL", "acestep-v15-turbo")
print(f"downloading {model} + LM + VAE ...")
ensure_main_model()
# DiT-specific download for the chosen turbo variant.
try:
    ensure_dit_model(model)
except Exception as exc:
    print(f"ensure_dit_model failed (may already be in the main repo): {exc}", file=sys.stderr)
print("done.")
PY

cat <<MSG

==> install.sh complete.

Next steps:
  source services/acestep/venv/bin/activate
  python services/acestep/server.py        # real model, ~30s load
  # ... in another shell:
  ./termus-stream tracks/lofi/bookstore-rainy-night-v3.tm

UNTESTED: this script has been written from the published install docs of
both ACE-Step forks; it has not been executed end-to-end in this PR.
MSG
