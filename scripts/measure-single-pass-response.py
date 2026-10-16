#!/usr/bin/env python3
"""Measure a single answer-only Codex assignment through a local authenticated daemon/control-plane and SSE."""
import argparse
import hashlib
import json
import os
import pathlib
import re
import secrets
import signal
import socket
import subprocess
import tempfile
import threading
import time
import urllib.error
import urllib.request

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--daemon', required=True, type=pathlib.Path)
parser.add_argument('--server', required=True, type=pathlib.Path)
parser.add_argument('--port', type=int, default=0)
parser.add_argument('--model', default='')
args = parser.parse_args()
for binary in [args.daemon, args.server]:
    if not binary.is_absolute() or not os.access(binary, os.X_OK):
        parser.error('Binary paths must be absolute executables')
with socket.socket() as probe:
    probe.bind(('127.0.0.1', args.port))
    args.port = probe.getsockname()[1]
root = pathlib.Path(tempfile.mkdtemp(prefix='riido-local-validation-')).resolve()
server_dir = root / 'server'
server_dir.mkdir(mode=0o700)
workspace, task = 'local-validation', 'answer-only'
context = {workspace: {task: {'component': {'id': task, 'title': 'Summarize the provided note'}, 'document': {'content': 'Summarize this note in exactly two short Korean sentences. Answer directly; do not use tools or create files. Note: Riido agents can answer questions and create documents without a repository. Repository access is necessary only when the task requires existing project source.'}}}}
(server_dir / 'task-contexts.json').write_text(json.dumps(context))
bindings = []
(root / 'bindings.json').write_text(json.dumps(bindings))
policy = {'schema_version': 'riido-policy-bundle.v1', 'version': 'policy-bundle.local-validation.v1', 'effective_since': '2026-10-12T00:00:00Z', 'trust_tier_policies': {'Host': {'allowed_surfaces': {'native_config_hooks': ['claude:command-hooks:audit'], 'tool_use': ['tool:protected-path-write']}}}}
(root / 'policy.json').write_text(json.dumps(policy))
token = secrets.token_urlsafe(32)
auth = json.dumps([{'principal_id': 'local-validation', 'token_sha256': hashlib.sha256(token.encode()).hexdigest(), 'scopes': ['riido:*']}])
base = f'http://127.0.0.1:{args.port}'
client = f'/v2/client/workspaces/{workspace}/ai-agent'
processes, logs, streams = [], [], []
finished = threading.Event()
stream_ready = threading.Event()
events = []
started = time.monotonic()

def api(path, body=None):
    request = urllib.request.Request(base + path, data=None if body is None else json.dumps(body).encode(), headers={'X-Riido-AI-Agent-Token': token, 'Content-Type': 'application/json'})
    with urllib.request.urlopen(request, timeout=8) as response:
        return json.load(response)

def launch(binary, extra_env, arguments, name):
    env = os.environ.copy()
    # The test profile never inherits production authentication/persistence.
    for key in list(env):
        if key.startswith('RIIDO_'):
            env.pop(key)
    env.update(extra_env)
    log = open(root / (name + '.log'), 'w')
    logs.append(log)
    process = subprocess.Popen([str(binary), *arguments], env=env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
    processes.append(process)
    return process

def wait_for(predicate, timeout=180):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if any(p.poll() is not None for p in processes):
            raise RuntimeError('A local process exited; inspect the retained logs')
        result = predicate()
        if result:
            return result
        time.sleep(0.1)
    raise TimeoutError('Local validation did not reach the expected state')

def ready():
    try:
        return api('/healthz')
    except (urllib.error.URLError, TimeoutError):
        return None

def read_stream():
    request = urllib.request.Request(base + client + '/events', headers={'X-Riido-AI-Agent-Token': token})
    try:
        with urllib.request.urlopen(request, timeout=180) as response:
            streams.append(response)
            stream_ready.set()
            for line in response:
                if finished.is_set():
                    break
                if line.startswith(b'data:'):
                    events.append({'seconds': time.monotonic() - started, 'payload': json.loads(line[5:])})
    except (OSError, ValueError):
        pass

try:
    launch(args.server, {'RIIDO_AI_SERVER_ADDR': f'127.0.0.1:{args.port}', 'RIIDO_AI_SERVER_LOCAL_STATE_DIR': str(server_dir), 'RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON': auth}, [], 'server')
    wait_for(ready, 20)
    device = api(f'/v2/desktop/workspaces/{workspace}/devices/enroll', {'machine_id': root.name, 'display_name': 'Disposable validation', 'platform': 'darwin'})
    launch(args.daemon, {'RIIDO_DAEMON_ID': device['device_id'], 'RIIDO_DEVICE_ID': device['device_id'], 'RIIDO_DEVICE_SECRET': device['device_secret'], 'RIIDO_SAAS_URL': base, 'RIIDO_DAEMON_PROFILE': 'local', 'RIIDO_WORKDIR_ROOT': str(root / 'workspaces'), 'RIIDO_EXECUTION_BINDINGS_FILE': str(root / 'bindings.json'), 'RIIDO_POLICY_BUNDLE_PATH': str(root / 'policy.json'), 'RIIDO_DAEMON_RUN_HARD_TIMEOUT_SECONDS': '150'}, ['daemon', 'start', '--foreground', '--socket', str(root / 'daemon.sock'), '--pid-file', str(root / 'daemon.pid'), '--lock-file', str(root / 'daemon.lock')], 'daemon')
    def runtime_ready():
        devices = api(client + '/devices')['devices']
        for item in devices:
            if item['device_id'] == device['device_id']:
                return next((runtime for runtime in item.get('runtimes', []) if runtime['kind'] == 'codex'), None)
    runtime = wait_for(runtime_ready, 60)
    agent = api(client + '/agents', {'name': 'Response timing probe', 'visibility': 'private', 'runtime_id': runtime['runtime_id'], **({'model_id': args.model} if args.model else {})})['agent']
    threading.Thread(target=read_stream, daemon=True).start()
    if not stream_ready.wait(10):
        raise RuntimeError("SSE connection was not established")
    request_at = time.monotonic() - started
    assignment = api(client + f'/tasks/{task}/agent-assignments', {'agent_id': agent['agent_id']})
    def terminal():
        threads = api(client + f'/tasks/{task}/threads')['threads']
        return next((thread for thread in threads if thread['assignment_id'] == assignment['assignment_id'] and thread['assignment_state'] in ['completed', 'failed', 'stopped']), None)
    result = wait_for(terminal)
    completed_at = time.monotonic() - started
    (root / 'result.json').write_text(json.dumps(result, ensure_ascii=False, indent=2))
    (root / 'sse-events.json').write_text(json.dumps(events, ensure_ascii=False, indent=2))
    assert result['assignment_state'] == 'completed', result.get('message')
    assert result.get('result_message', '').strip(), 'No final answer delivered'
    partials = [event for event in events if event['seconds'] < completed_at and event['payload'].get('assignment_id') == assignment['assignment_id'] and any(line.get('message_key') == 'assistant.partial' for line in event['payload'].get('lines', []))]
    assert partials, 'No partial response observed before completion'
    measurement = {'request_to_first_sse_ms': round((partials[0]['seconds'] - request_at)*1000), 'request_to_terminal_observed_ms': round((completed_at-request_at)*1000), 'partial_snapshots': len(partials), 'model': args.model or 'runtime-default', 'scenario': 'answer-only', 'poll_interval_ms': 100}
    run_events = [json.loads(line) for path in (root / 'workspaces').rglob('events.jsonl') for line in path.read_text().splitlines()]
    measurement['provider_versions'] = sorted({event['ProviderVersion'] for event in run_events if event.get('ProviderVersion')})
    timing_lines = (root / 'daemon.log').read_text().splitlines()
    measurement['daemon_claim_elapsed_ms'] = {match.group(1): int(match.group(2)) for line in timing_lines if (match := re.search(r'event=agent_response_timing .*phase=\"([^\"]+)\" claim_elapsed_ms=(\d+)', line))}
    measurement['native_sessions'] = sum(event.get('Type') == 'SessionPinned' for event in run_events)
    assert measurement['native_sessions'] == 1, 'Expected one native session'
    (root / 'measurement.json').write_text(json.dumps(measurement, indent=2))
    print(json.dumps(measurement), flush=True)
    print(f'Partial snapshots: {len(partials)}; evidence: {root}', flush=True)
finally:
    finished.set()
    for process in reversed(processes):
        if process.poll() is None:
            os.killpg(process.pid, signal.SIGTERM)
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
    for log in logs:
        log.close()
    print(f'Processes stopped; retained private local state: {root}', flush=True)
