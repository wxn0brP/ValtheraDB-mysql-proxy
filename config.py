import json
import os
import questionary

CONFIG_FILE = "config.json"

DEFAULT_CONFIG = {
    "server_url": "http://localhost:14785",
    "db_name": "",
    "auth_token": "",
    "port": 4000
}

def load_config():
    if os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE) as f:
            return json.load(f)
    return DEFAULT_CONFIG.copy()

def save_config(cfg):
    with open(CONFIG_FILE, "w") as f:
        json.dump(cfg, f, indent=4)
    print(f"\nConfig saved to {CONFIG_FILE}")

def edit_config(cfg):
    cfg["server_url"] = questionary.text("Server URL", default=cfg["server_url"]).ask()
    cfg["db_name"] = questionary.text("Database name", default=cfg["db_name"]).ask()
    cfg["auth_token"] = questionary.text("Auth token", default=cfg["auth_token"]).ask()
    port = questionary.text("Port", default=str(cfg["port"])).ask()
    try:
        cfg["port"] = int(port)
    except ValueError:
        print("Invalid port. Using default 4000.")
        cfg["port"] = 4000
    return cfg

if __name__ == "__main__":
    cfg = load_config()
    cfg = edit_config(cfg)
    save_config(cfg)