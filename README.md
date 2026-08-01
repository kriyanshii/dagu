<div align="center">
  <a href="https://dagu.sh">
    <img src="./assets/images/hero-logo.png" width="720" alt="Dagu: turn scripts into reliable workflows">
  </a>
  <p>
    <a href="https://docs.dagu.sh">Docs</a> ·
    <a href="https://docs.dagu.sh/writing-workflows/examples">Examples</a> ·
    <a href="https://dagu-demo-f5e33d0e.dagu.sh">Live demo</a> ·
    <a href="https://discord.gg/gpahPUjGRk">Discord</a>
  </p>
</div>

# Dagu

Dagu turns scripts and commands into reliable YAML workflows. It adds schedules, dependencies, retries, approvals, logs, and a Web UI in one open-source binary. You do not need an external database or message broker.

## Quick start

Install Dagu on macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/dagucloud/dagu/main/scripts/installer.sh | bash
```

The installer can add Dagu to your `PATH`, set up a background service, and create the first admin account. See the [installation guide](https://docs.dagu.sh/getting-started/installation/) for Windows, Docker, Homebrew, npm, Kubernetes, and manual options.

Save this as `hello.yaml`:

```yaml
steps:
  - id: hello
    run: echo "Hello from Dagu!"
  - id: done
    run: echo "Workflow finished"
    depends: hello
```

Run it:

```sh
dagu start hello.yaml
```

Start the Web UI in the same directory:

```sh
dagu start-all --dags .
```

Open <http://localhost:8080> to see the run, step logs, and history. The [full quickstart](https://docs.dagu.sh/getting-started/quickstart) also covers Docker, validation, expected output, and next steps.

If Dagu is useful, click **Star** at the top of this page. It helps other developers find the project.

## What Dagu gives you

- Keep your current scripts, commands, containers, and tools.
- Store readable workflow definitions in Git.
- Add dependencies, schedules, retries, timeouts, and approvals in YAML.
- Inspect live status, logs, and previous runs in the built-in Web UI.
- Start on one machine, then add queues or distributed workers if the workload grows.

## See it in action

Click the image to watch the short product walkthrough.

<div align="center">
  <a href="./assets/images/dagu-demo.mp4?raw=1">
    <img src="./assets/images/cockpit-demo-poster.jpg" width="720" alt="Dagu Cockpit showing queued, running, completed, and failed workflow runs">
  </a>
</div>

| Run details | Step logs |
|---|---|
| ![Run details in dark mode](./assets/images/readme-run-details-dark.png) | ![Workflow logs in dark mode](./assets/images/readme-logs-dark.png) |

You can also open the [live demo](https://dagu-demo-f5e33d0e.dagu.sh) and sign in with `demouser` / `demouser`.

## Why Dagu

Cron is easy to start but gives you little visibility once jobs depend on each other. Larger orchestrators solve that problem by adding services and a framework. Dagu keeps the operating model small:

```text
Traditional orchestrator              Dagu

Web server                            dagu start-all
Scheduler                             ├── Web UI
Workers                               ├── Scheduler
Database                              ├── Executor
Message broker                        └── Local file-backed state
Language runtime
```

The workflow calls the software you already use. It does not require you to move that code into a Dagu-specific framework.

## A practical workflow

This example runs a nightly report, retries the data step, and keeps the order explicit:

```yaml
schedule: "0 2 * * *"

steps:
  - id: extract
    run: python extract.py
    retry_policy:
      limit: 3
      interval_sec: 30

  - id: report
    run: ./build-report.sh
    depends: extract

  - id: archive
    run: tar -czf report.tgz report/
    depends: report
```

Dagu can also run containers, Kubernetes Jobs, SSH commands, SQL, HTTP requests, sub-workflows, human tasks, and reusable actions. Browse the [workflow examples](https://docs.dagu.sh/writing-workflows/examples) or the [YAML reference](https://docs.dagu.sh/writing-workflows/yaml-specification) when you need them.

## Common uses

- Replace fragile cron chains while keeping the underlying scripts.
- Run ETL, reporting, backup, media, and infrastructure jobs.
- Coordinate Docker, Kubernetes, SSH, SQL, and HTTP work in one graph.
- Give operators a controlled way to run approved internal tasks.
- Keep automation close to data on servers, edge devices, or private networks.
- Distribute heavier workloads to workers selected by labels.

## Ways to run Dagu

| Model | Where Dagu runs | Good fit |
|---|---|---|
| Single server | One `dagu start-all` process | Development, scheduled jobs, and internal automation |
| Self-hosted workers | Server and workers on your infrastructure | Private networks, heavier workloads, and multiple execution hosts |
| Managed | A dedicated managed Dagu instance, with optional private workers | Teams that want Dagu operated for them |

The same YAML works across these models. See [deployment models](https://docs.dagu.sh/overview/deployment-models) for the architecture, security boundaries, and setup details.

## AI tools are optional

Dagu has a built-in MCP endpoint for clients that need to inspect workflows or control runs:

```text
http://localhost:8080/mcp
```

For workflow-authoring help in Claude Code, Codex, Gemini CLI, and other coding tools, install the Dagu skill:

```sh
gh skill install dagucloud/dagu dagu
```

The basic workflow engine does not require an AI provider. Start with the [MCP guide](https://docs.dagu.sh/mcp/quickstart) only if you need that integration.

## Learn more

| Topic | Documentation |
|---|---|
| Install and first run | [Quickstart](https://docs.dagu.sh/getting-started/quickstart) |
| Workflow syntax | [Writing workflows](https://docs.dagu.sh/writing-workflows/) |
| Ready-to-run YAML | [Examples](https://docs.dagu.sh/writing-workflows/examples) |
| Built-in and packaged actions | [Dagu Actions](https://docs.dagu.sh/dagu-actions/) |
| Web UI and API | [Web UI](https://docs.dagu.sh/overview/web-ui) |
| Authentication and secrets | [Server administration](https://docs.dagu.sh/server-admin/) |
| Queues and workers | [Distributed execution](https://docs.dagu.sh/server-admin/distributed/) |
| CLI commands | [CLI reference](https://docs.dagu.sh/getting-started/cli) |

## Development

Prerequisites: [Go 1.26+](https://go.dev/doc/install), [Node.js](https://nodejs.org/en/download/), and [pnpm](https://pnpm.io/installation).

```sh
git clone https://github.com/dagucloud/dagu.git
cd dagu
make build
make test
make lint
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the development workflow and code standards.

## Community

- [Discord](https://discord.gg/gpahPUjGRk) for questions and discussion
- [GitHub Issues](https://github.com/dagucloud/dagu/issues) for bugs and feature requests
- [Bluesky](https://bsky.app/profile/dagu-sh.bsky.social) for project updates
- [GitHub Sponsors](https://github.com/sponsors/dagucloud) to support maintenance

Thanks to [/bin labs](https://slashbinlabs.com/) and everyone who has contributed code, documentation, testing, or feedback.

## License

Dagu is licensed under [GNU GPLv3](./LICENSE). See [LICENSING.md](./LICENSING.md) for embedded API and commercial embedding terms.
