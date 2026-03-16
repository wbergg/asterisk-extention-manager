# Asterisk Extention Manager (AEM)

A self-service web portal for managing Asterisk PBX SIP extensions. Admins delegate numeric extension ranges to users, who can then register and manage SIP extensions within their permitted range. The Go backend is the source of truth, managing a SQLite database and pushing PJSIP configurations to Asterisk.

Single-binary deployment — the Vue frontend is embedded into the Go binary at compile time.

## Features

- **Range-based delegation** — admins assign extension ranges (e.g. 2000-2050) to users
- **Self-service registration** — users register extensions within their range with auto-generated SIP credentials
- **Extension editing** — users can update caller ID and SIP password with validation (8+ chars, 1 uppercase, 1 number)
- **Asterisk integration** — generates `pjsip_custom.conf` using template inheritance and triggers `pjsip reload` on every change
- **HTTPS** — bring your own certificate or automatic Let's Encrypt provisioning
- **Admin dashboard** — full CRUD for users and extensions, force sync, impersonation
- **Extension blocking** — admins can block specific numbers within user ranges
- **Call log** — view Asterisk CDR records with charts (calls per day, status breakdown, top 5 busiest days), infinite scroll pagination
- **Directory** — searchable list of all extensions and names, plus Cisco XML phonebook endpoint
- **User settings** — self-service password change with validation (10+ chars, 1 uppercase, 1 number)
- **Admin impersonation** — view the portal as any user for troubleshooting
- **Per-user call log access** — admin toggle to enable/disable call log for individual users
- **Concurrency safe** — SQLite UNIQUE constraint ensures only one user can claim an extension

## Requirements

- Go 1.21+ (with CGO enabled for SQLite)
- Node.js 18+ (for building the frontend)
- Asterisk PBX (optional — can run in dev mode without it)

## Quick Start

```bash
# Install frontend dependencies and build
cd frontend && npm install && npm run build && cd ..

# Copy and edit the config file
cp config/config.example.json config/config.json

# Run in dev mode (set asterisk_cmd to "echo" in config.json)
go run .
```

Open http://localhost:8080 and log in with `admin` / `admin`.

## Build

```bash
cd frontend && npm install && npm run build && cd ..
CGO_ENABLED=1 go build -o aem .
```

This produces a single `aem` binary with the frontend embedded.

## Configuration

Configuration is via a JSON file at `config/config.json` (override with `-config` flag):

```bash
./aem                              # uses config/config.json
./aem -config /etc/aem/config.json # custom path
```

Copy `config/config.example.json` to get started:

```json
{
    "listen_addr": ":8080",
    "db_path": "./data/extensions.db",
    "jwt_secret": "change-me",
    "pjsip_conf_path": "/etc/asterisk/pjsip_custom.conf",
    "asterisk_cmd": "asterisk",
    "admin_user": "admin",
    "admin_pass": "admin",
    "cdr_log_path": "/var/log/asterisk/cdr-csv/Master.csv",
    "tls_domain": "",
    "tls_cert_dir": "./certs",
    "tls_cert_file": "",
    "tls_key_file": ""
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `listen_addr` | `:8080` | HTTP listen address (used when no TLS is configured) |
| `db_path` | `./data/extensions.db` | SQLite database path |
| `jwt_secret` | `change-me` | Secret for signing JWT tokens |
| `pjsip_conf_path` | `/etc/asterisk/pjsip_custom.conf` | Output path for generated PJSIP config |
| `asterisk_cmd` | `asterisk` | Path to Asterisk binary (set to `echo` for dev) |
| `admin_user` | `admin` | Default admin username (seeded on first run) |
| `admin_pass` | `admin` | Default admin password (seeded on first run) |
| `cdr_log_path` | `/var/log/asterisk/cdr-csv/Master.csv` | Path to Asterisk CDR CSV log file |
| `tls_domain` | _(empty)_ | Domain for Let's Encrypt auto-cert |
| `tls_cert_dir` | `./certs` | Directory for storing Let's Encrypt certificates |
| `tls_cert_file` | _(empty)_ | Path to existing TLS certificate (e.g. fullchain.pem) |
| `tls_key_file` | _(empty)_ | Path to existing TLS private key (e.g. privkey.pem) |

The admin user is only seeded when the users table is empty (first run).

## HTTPS

Three modes depending on which TLS fields are set:

**Existing certificate** — point to your cert and key files:
```json
{
    "tls_cert_file": "/etc/letsencrypt/live/example.com/fullchain.pem",
    "tls_key_file": "/etc/letsencrypt/live/example.com/privkey.pem"
}
```

**Let's Encrypt auto-cert** — set the domain (cert/key fields must be empty):
```json
{
    "tls_domain": "pbx.example.com",
    "tls_cert_dir": "./certs"
}
```

**Plain HTTP** — leave all TLS fields empty, uses `listen_addr`.

Both HTTPS modes bind port 443 (HTTPS) and port 80. On port 80, `/directory.xml` and `/logo.bmp` are served directly over plain HTTP (for Cisco IP phones that don't support HTTPS), while all other requests are redirected to HTTPS. The process needs permission to bind these ports:

```bash
# Allow unprivileged port binding (persistent)
sudo sysctl -w net.ipv4.ip_unprivileged_port_start=80
echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee /etc/sysctl.d/99-unprivileged-ports.conf
```

## API Routes

### Public

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/login` | Authenticate, returns JWT |
| GET | `/directory.xml` | Cisco IP Phone XML directory |

### Authenticated (JWT required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/me` | Current user profile |
| PUT | `/api/me/password` | Change own password |
| GET | `/api/extensions` | List own extensions |
| POST | `/api/extensions` | Register extension `{extension, callerid}` |
| GET | `/api/extensions/{ext}` | Get extension details + SIP creds |
| PUT | `/api/extensions/{ext}` | Update extension (caller ID, SIP password) |
| DELETE | `/api/extensions/{ext}` | Delete own extension |
| GET | `/api/directory` | All extensions as JSON (name + number) |

### Call Log (JWT + call_log_access)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/cdr` | Paginated CDR records (`?offset=0&limit=100`) |
| GET | `/api/cdr/stats` | Aggregate stats (total, answered, avg duration, per-day) |

### Admin (JWT + role=admin)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/users` | List all users |
| POST | `/api/admin/users` | Create user |
| PUT | `/api/admin/users/{id}` | Update user |
| DELETE | `/api/admin/users/{id}` | Delete user (cascades extensions) |
| GET | `/api/admin/extensions` | List all extensions |
| PUT | `/api/admin/extensions/{ext}` | Update any extension |
| DELETE | `/api/admin/extensions/{ext}` | Delete any extension |
| POST | `/api/admin/sync` | Force Asterisk config sync + reload |
| POST | `/api/admin/impersonate/{id}` | Get JWT token for another user |
| GET | `/api/admin/blocked` | List blocked extension numbers |
| POST | `/api/admin/blocked` | Block an extension number |
| DELETE | `/api/admin/blocked/{ext}` | Unblock an extension number |

## Project Structure

```
├── main.go                          # Router setup, server startup
├── frontend_embed.go                # go:embed for frontend/dist
├── config/
│   ├── config.json                  # Your config (gitignored)
│   └── config.example.json          # Example config
├── internal/
│   ├── config/config.go             # JSON config loader
│   ├── database/
│   │   ├── database.go              # SQLite connection (WAL mode)
│   │   ├── migrations.go            # Embedded SQL migration runner
│   │   └── migrations/
│   │       ├── 001_schema.sql       # Base schema (users, extensions, audit_log)
│   │       ├── 002_indexes.sql      # Performance indexes
│   │       ├── 003_call_log_access.sql  # Per-user call log toggle
│   │       └── 004_blocked_extensions.sql  # Blocked extension numbers
│   ├── models/
│   │   ├── user.go                  # User CRUD
│   │   ├── extension.go             # Extension CRUD + SIP credential generation
│   │   └── blocked.go               # Blocked extension CRUD
│   ├── auth/
│   │   ├── jwt.go                   # JWT token generation/parsing
│   │   └── middleware.go            # Auth, admin, and call log access middleware
│   ├── handlers/
│   │   ├── auth_handler.go          # Login, profile, password change, impersonation
│   │   ├── user_handler.go          # Admin user management
│   │   ├── extension_handler.go     # Extension management, directory, blocking
│   │   └── cdr_handler.go           # Call log from Asterisk CDR CSV
│   ├── asterisk/
│   │   ├── pjsip.go                 # PJSIP config generation (template inheritance)
│   │   └── reload.go                # Asterisk reload via CLI
│   └── middleware/
│       └── cors.go                  # CORS headers
├── frontend/                        # Vue 3 + Vite + Tailwind CSS 4
│   └── src/
│       ├── assets/                  # Static assets (logo)
│       ├── views/                   # Login, Dashboard, Admin, Call Log, Settings, Directory
│       ├── components/              # Extension/User tables and forms
│       ├── stores/                  # Pinia stores (auth, extensions)
│       ├── api/client.ts            # Axios + JWT interceptor
│       └── router/index.ts          # Vue Router with auth guards
└── data/                            # SQLite database location
```

## Asterisk Integration

On every extension create, update, or delete:

1. All extensions are queried from the database
2. A complete `pjsip_custom.conf` is generated using Asterisk template inheritance (`endpoint-basic`, `auth-userpass`, `aor-single`)
3. The file is written atomically (write to `.tmp`, then rename)
4. `asterisk -rx "pjsip reload"` is executed

If the reload fails, the database change is kept (DB is source of truth). Use the admin "Force Sync" button to retry.

Make sure Asterisk includes the generated config file, e.g. in `/etc/asterisk/pjsip.conf`:

```ini
#include pjsip_custom.conf
```

The process running AEM needs permission to execute `asterisk -rx`. Add the user to the `asterisk` group and ensure the Asterisk socket is group-writable:

```bash
sudo usermod -aG asterisk <your-user>
sudo chmod g+w /var/run/asterisk/asterisk.ctl
```

## Call Log (CDR)

The call log reads Asterisk's CDR (Call Detail Records) CSV file. To enable CDR CSV logging, ensure `/etc/asterisk/cdr.conf` contains:

```ini
[csv]
usegmtime = no
loguniqueid = yes
```

The call log page includes:
- Stats cards (total calls, answered, average duration, answer rate)
- Bar chart of calls per day
- Doughnut chart of call status breakdown
- Top 5 busiest days
- Infinite scroll table with color-coded status badges

Call log access is enabled for all users by default. Admins can disable it per-user from the admin panel.

## Cisco IP Phone Directory

AEM serves a Cisco XML directory at `/directory.xml` and a phone logo at `/logo.bmp` for use with Cisco IP phones (7900 series, etc.). These endpoints are always available over plain HTTP on port 80, even when HTTPS is enabled. Configure the phone URLs:

```
Directory: http://your-server/directory.xml
Logo:      http://your-server/logo.bmp
```

Place your `logo.bmp` in `frontend/public/` before building (Cisco 7900 series uses 80x53 pixel 1-bit BMP).

## License

MIT
