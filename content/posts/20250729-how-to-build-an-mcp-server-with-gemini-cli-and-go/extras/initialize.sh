#!/bin/bash
echo '{ "jsonrpc": "2.0", "id": 1, "method": "initialize", "params": { "protocolVersion": "2025-06-18", "clientInfo": { "name": "Gemini CLI", "version": "1.0.0" } } }' | godoctor
