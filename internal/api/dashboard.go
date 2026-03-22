package api

import "net/http"

func (s *server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(adminDashboardHTML))
}

const adminDashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>SDFS Admin Dashboard</title>
  <style>
    :root {
      --bg: #0f172a;
      --panel: #111827;
      --muted: #94a3b8;
      --text: #e5e7eb;
      --accent: #22d3ee;
      --ok: #10b981;
      --warn: #f59e0b;
      --err: #ef4444;
      --border: #1f2937;
    }
    body {
      margin: 0;
      font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Helvetica, Arial;
      background: radial-gradient(circle at 15% 15%, #1e293b, var(--bg) 35%), var(--bg);
      color: var(--text);
    }
    .wrap {
      max-width: 1080px;
      margin: 0 auto;
      padding: 24px;
    }
    h1 { margin-top: 0; }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 16px;
    }
    .card {
      border: 1px solid var(--border);
      border-radius: 12px;
      background: color-mix(in srgb, var(--panel) 92%, black 8%);
      padding: 14px;
    }
    .row { display: flex; gap: 8px; flex-wrap: wrap; }
    input {
      width: 100%;
      padding: 10px;
      margin: 6px 0;
      border-radius: 8px;
      border: 1px solid #334155;
      background: #0b1220;
      color: var(--text);
      box-sizing: border-box;
    }
    button {
      border: 1px solid #1f2937;
      border-radius: 8px;
      background: #0b1220;
      color: var(--text);
      padding: 8px 12px;
      cursor: pointer;
    }
    button:hover { border-color: var(--accent); }
    pre {
      background: #020617;
      border: 1px solid #1e293b;
      border-radius: 8px;
      padding: 10px;
      overflow: auto;
      max-height: 260px;
      color: #cbd5e1;
      font-size: 12px;
    }
    .muted { color: var(--muted); }
    .status-ok { color: var(--ok); }
    .status-err { color: var(--err); }
  </style>
</head>
<body>
  <div class="wrap">
    <h1>SDFS Admin Dashboard</h1>
    <p class="muted">Use your admin token below. This dashboard talks to the same admin APIs used by the CLI.</p>

    <div class="card">
      <label>Admin Token</label>
      <input id="adminToken" placeholder="Paste X-Admin-Token value" />
      <p id="status" class="muted">Ready.</p>
    </div>

    <div class="grid" style="margin-top:16px;">
      <div class="card">
        <h3>Invites and Enrollment</h3>
        <input id="ttlSeconds" placeholder="Invite TTL seconds (default 600)" value="600" />
        <div class="row">
          <button onclick="issueInvite()">Issue Invite</button>
          <button onclick="listPending()">List Pending</button>
        </div>
        <input id="requestId" placeholder="Request ID for approve" />
        <button onclick="approveEnrollment()">Approve Enrollment</button>
      </div>

      <div class="card">
        <h3>Devices</h3>
        <div class="row">
          <button onclick="listDevices()">List Devices</button>
        </div>
        <input id="deviceId" placeholder="Device ID" />
        <div class="row">
          <button onclick="startSession()">Start Session</button>
          <button onclick="revokeDevice()">Revoke Device</button>
        </div>
      </div>

      <div class="card">
        <h3>Sessions</h3>
        <div class="row">
          <button onclick="listSessions()">List Active Sessions</button>
        </div>
        <input id="sessionId" placeholder="Session ID" />
        <button onclick="heartbeatSession()">Heartbeat Session</button>
      </div>
    </div>

    <div class="card" style="margin-top:16px;">
      <h3>Response</h3>
      <pre id="output">{}</pre>
    </div>
  </div>

  <script>
    function token() {
      return document.getElementById('adminToken').value.trim();
    }

    function setStatus(msg, ok) {
      const el = document.getElementById('status');
      el.textContent = msg;
      el.className = ok ? 'status-ok' : 'status-err';
    }

    function setOut(v) {
      document.getElementById('output').textContent = JSON.stringify(v, null, 2);
    }

    async function adminFetch(path, method, body) {
      const tok = token();
      if (!tok) {
        setStatus('Admin token required', false);
        throw new Error('admin token required');
      }
      const headers = { 'X-Admin-Token': tok };
      if (body) headers['Content-Type'] = 'application/json';

      const resp = await fetch(path, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        setStatus('HTTP ' + resp.status, false);
      } else {
        setStatus('HTTP ' + resp.status, true);
      }
      setOut(data);
      return data;
    }

    function ttlValue() {
      const n = Number(document.getElementById('ttlSeconds').value || '600');
      return Number.isFinite(n) && n > 0 ? Math.floor(n) : 600;
    }

    async function issueInvite() {
      await adminFetch('/v1/admin/invites', 'POST', { ttlSeconds: ttlValue() });
    }
    async function listPending() {
      await adminFetch('/v1/admin/enrollments/pending', 'GET');
    }
    async function approveEnrollment() {
      const requestId = document.getElementById('requestId').value.trim();
      await adminFetch('/v1/admin/enrollments/approve', 'POST', { requestId });
    }
    async function listDevices() {
      await adminFetch('/v1/admin/devices', 'GET');
    }
    async function revokeDevice() {
      const deviceId = document.getElementById('deviceId').value.trim();
      await adminFetch('/v1/admin/devices/revoke', 'POST', { deviceId });
    }
    async function startSession() {
      const deviceId = document.getElementById('deviceId').value.trim();
      await adminFetch('/v1/admin/sessions/start', 'POST', { deviceId });
    }
    async function listSessions() {
      await adminFetch('/v1/admin/sessions/active', 'GET');
    }
    async function heartbeatSession() {
      const sessionId = document.getElementById('sessionId').value.trim();
      await adminFetch('/v1/admin/sessions/heartbeat', 'POST', { sessionId });
    }
  </script>
</body>
</html>
`
