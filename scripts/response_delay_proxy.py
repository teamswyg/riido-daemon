"""Loopback-only test proxy for isolating serial daemon event HTTP latency."""
import collections
import http.server
import json
import threading
import time
import urllib.error
import urllib.request


class ResponseDelayProxy:
    def __init__(self, upstream, delay_ms):
        self.records = []
        self.lock = threading.Lock()
        owner = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def do_GET(self):
                self.forward()

            def do_POST(self):
                self.forward()

            def forward(self):
                data = self.rfile.read(int(self.headers.get('Content-Length', 0)))
                headers = {k: v for k, v in self.headers.items()
                           if k.lower() not in ('host', 'connection', 'content-length')}
                request = urllib.request.Request(upstream + self.path,
                    data=data if self.command == 'POST' else None,
                    headers=headers, method=self.command)
                start = time.monotonic()
                try:
                    response = urllib.request.urlopen(request, timeout=60)
                except urllib.error.HTTPError as error:
                    response = error
                with response:
                    body = response.read()
                    event = self.command == 'POST' and self.path.endswith('/events')
                    if event:
                        payload = json.loads(data)
                        # Retain timing and event type only, never payload/auth.
                        with owner.lock:
                            owner.records.append({'event_type': payload.get('event_type'),
                                'upstream_ms': round((time.monotonic() - start) * 1000, 2)})
                        time.sleep(delay_ms / 1000)
                    self.send_response(response.status)
                    self.send_header('Content-Type', 'application/json')
                    self.send_header('Content-Length', str(len(body)))
                    self.end_headers()
                    try:
                        self.wfile.write(body)
                    except (BrokenPipeError, ConnectionResetError):
                        pass

        self.server = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        self.url = f'http://127.0.0.1:{self.server.server_port}'
        threading.Thread(target=self.server.serve_forever, daemon=True).start()

    def summary(self):
        with self.lock:
            return {'event_http_requests': len(self.records),
                'event_types': dict(collections.Counter(r['event_type'] for r in self.records)),
                'upstream_event_ms': [r['upstream_ms'] for r in self.records]}

    def close(self):
        self.server.shutdown()
        self.server.server_close()
