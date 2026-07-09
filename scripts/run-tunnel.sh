#!/data/data/com.termux/files/usr/bin/bash
# Run from Termux, after the blog server is already running on :8080.
# Prints a random https://*.trycloudflare.com URL - copy it into the
# Worker's PHONE_URL variable in the Cloudflare dashboard.

cloudflared tunnel --url http://localhost:8080
