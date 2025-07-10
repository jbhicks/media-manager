# 1. Remove Python 3.13 and install Python 3.11
sudo pacman -R python python-pip
sudo pacman -S python3.11 python3.11-pip

# 2. Install system dependencies for lxml
sudo pacman -S libxml2 libxslt

# 3. Verify Python 3.11
python3.11 --version        # Should output 3.11.x
python3.11 -m pip --version # Should output pip 24.2 or similar

# 4. Ensure Ollama is installed and running
curl -fsSL https://ollama.com/install.sh | sh
sudo systemctl enable ollama
sudo systemctl start ollama

# 5. Pull the llama3.1:8b-instruct-q4_K_M model
ollama pull llama3.1:8b-instruct-q4_K_M

# 6. Create and activate a virtual environment with Python 3.11
python3.11 -m venv langroid-venv
source langroid-venv/bin/activate

# 7. Install Rust (required for tiktoken)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
rustc --version # Should output rustc 1.x.x

# 8. Install Langroid and dependencies
pip install langroid langroid[ollama] openai tiktoken lxml --no-cache-dir

# 9. Optimize Ollama for GPU (Flash Attention)
sudo systemctl edit ollama
# Add in the editor:
# [Service]
# Environment="OLLAMA_FLASH_ATTENTION=1"
sudo systemctl daemon-reload
sudo systemctl restart ollama
