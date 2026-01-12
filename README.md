# LazyAdmin

A config-driven TUI (Terminal User Interface) database admin engine. Generate a database admin dashboard instantly with just a YAML configuration file.

## Features

- **No-Code Admin Pages**: Define views with raw SQL queries in YAML
- **Dynamic Schema**: Handles any table structure without hardcoded column names
- **Split-View Layout**: Sidebar navigation + data table view
- **Dracula Theme**: Clean, dark mode interface
- **Keyboard-Driven**: Full keyboard navigation support

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

Create an `admin.yaml` file:

```yaml
project_name: "My Admin"
database: "./mydata.db"

views:
  - title: "All Users"
    description: "View all registered users"
    query: "SELECT id, email, role, created_at FROM users LIMIT 50"

  - title: "Recent Orders"
    description: "Last 10 orders"
    query: "SELECT id, amount, status FROM orders ORDER BY id DESC LIMIT 10"
```

## Keyboard Controls

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Execute query / Select |
| `Tab` | Switch focus (Sidebar ↔ Table) |
| `q` / `Ctrl+C` | Quit |

## Tech Stack

- **Language**: Go 1.21+
- **TUI Framework**: [Charmbracelet](https://charm.sh/) (Bubble Tea, Bubbles, Lipgloss)
- **Database**: SQLite (via `go-sqlite3`)

## License

MIT
