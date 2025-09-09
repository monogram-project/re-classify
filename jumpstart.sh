# A rough and ready script that installs most of the stuff you need.

# Install basic tools
sudo apt update
sudo apt install -y curl git

# Ensure ~/.local/bin exists.
echo 'Ensure ~/.local/bin is on your $PATH'
mkdir -p "$HOME/.local/bin"

# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install just
mkdir -p "$HOME/.local/bin"
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to "$HOME/.local/bin"

