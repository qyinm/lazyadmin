# LazyAdmin

> **Stop building React admin pages for internal tools. Just write SQL.**

A config-driven TUI (Terminal User Interface) database admin engine. Generate a database admin dashboard instantly with just a YAML configuration file.

## Install

```bash
go install github.com/qyinm/lazyadmin@latest
```

## Why?

Every internal tool starts with "I just need to view this table" and ends with a week of fighting React components, state management, and API endpoints. 

**LazyAdmin** lets you skip the frontend work. Write a SQL query in a YAML file, and get a full TUI dashboard with CRUD capabilities instantly.

## Features

- **Multi-Database Management**: Connect to and switch between multiple databases in a single session
- **No-Code Admin Pages**: Define views with raw SQL queries in YAML
- **CRUD Operations**: Create, Read, Update, and Delete records directly from the terminal
- **Table Browser**: Explore database tables automatically without defining views
- **Dynamic Schema**: Handles any table structure without hardcoded column names
- **SSH Tunnel**: Connect to remote databases through SSH
- **3-Pane Layout**: Connections sidebar + Tables sidebar + Data table view
- **Universal Compatibility**: Uses standard ANSI 16-color palette and reverse-video for guaranteed rendering across all terminals (including Warp, iTerm2, etc.)
- **Keyboard-Driven**: Full keyboard navigation with clear focus indicators (thick borders)

## Installation

```bash
# Clone the repository
git clone https://github.com/qyinm/lazyadmin.git
cd lazyadmin

# Build
go build -o lazyadmin .
```

## Usage

```bash
# Run with default config (admin.yaml)
./lazyadmin

# Run with custom config
./lazyadmin path/to/config.yaml
```

## Configuration

### Multi-Database Support

You can now define multiple database connections in your configuration file.

```yaml
project_name: "My Dashboard"

connections:
  - label: "Production"
    driver: postgres
    host: prod-db.example.com
    user: admin
    name: main_db
  
  - label: "Local Dev"
    driver: sqlite
    path: "./local.db"

views:
  - title: "Active Users"
    query: "SELECT * FROM users WHERE active = true"
```

### SQLite

```yaml
project_name: "My Admin"

database:
  driver: sqlite
  path: "./mydata.db"

views:
  - title: "All Users"
    description: "View all registered users"
    query: "SELECT id, email, role FROM users LIMIT 50"
```

### PostgreSQL

```yaml
project_name: "Production Admin"

database:
  driver: postgres
  host: localhost
  port: 5432
  user: admin
  password: secret
  name: myapp
  ssl_mode: disable

views:
  - title: "Users"
    description: "All users"
    query: "SELECT * FROM users LIMIT 100"
```

### MySQL

```yaml
project_name: "MySQL Admin"

database:
  driver: mysql
  host: localhost
  port: 3306
  user: root
  password: secret
  name: myapp

views:
  - title: "Orders"
    description: "Recent orders"
    query: "SELECT * FROM orders ORDER BY id DESC LIMIT 50"
```

### SSH Tunnel (Remote Database)

```yaml
project_name: "Remote Admin"

database:
  driver: postgres
  host: 127.0.0.1
  port: 5432
  user: dbuser
  password: dbpass
  name: production
  ssh:
    host: bastion.example.com
    port: 22
    user: ubuntu
    private_key: ~/.ssh/id_rsa

views:
  - title: "Users"
    description: "Production users"
    query: "SELECT id, email, created_at FROM users LIMIT 50"
```

## Database Configuration Options

| Field | Description | Required |
|-------|-------------|----------|
| `driver` | Database driver: `sqlite`, `postgres`, `mysql` | Yes |
| `path` | Path to SQLite database file | SQLite only |
| `host` | Database host | PostgreSQL/MySQL |
| `port` | Database port (default: 5432/3306) | No |
| `user` | Database username | PostgreSQL/MySQL |
| `password` | Database password | PostgreSQL/MySQL |
| `name` | Database name | PostgreSQL/MySQL |
| `ssl_mode` | SSL mode for PostgreSQL | No |

## SSH Configuration Options

| Field | Description | Required |
|-------|-------------|----------|
| `host` | SSH server hostname | Yes |
| `port` | SSH port (default: 22) | No |
| `user` | SSH username | Yes |
| `password` | SSH password (or key passphrase) | No |
| `private_key` | Path to SSH private key | No |

## Keyboard Controls

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select connection / table |
| `Tab` | Switch focus (Connections ↔ Tables ↔ Data Table) |
| `n` | Add New Connection |
| `t` | Toggle Mode (View / Table Browser) |
| `i` | Insert Record (Table Browser Mode) |
| `e` | Edit Record (Table Browser Mode) |
| `d` | Delete Record (Table Browser Mode) |
| `r` | Refresh Table |
| `q` / `Ctrl+C` | Quit |

## Tech Stack

- **Language**: Go 1.21+
- **TUI Framework**: [Charmbracelet](https://charm.sh/) (Bubble Tea, Bubbles, Lipgloss)
- **Databases**: SQLite, PostgreSQL, MySQL

## License

MIT
