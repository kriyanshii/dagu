<div align="center">
  <a href="https://dagu.sh">
    <img src="./assets/images/hero-logo.png" width="720" alt="Dagu: built for teams whose main work is not orchestration">
  </a>
  <p>
    <a href="https://docs.dagu.sh">Docs</a> ·
    <a href="https://docs.dagu.sh/writing-workflows/examples">Examples</a> ·
    <a href="https://dagu-demo-f5e33d0e.dagu.sh">Live demo</a>
    <code>(username/password: demouser)</code> ·
    <a href="https://discord.gg/gpahPUjGRk">Discord</a>
  </p>
</div>

# Dagu

Local-first orchestration that just works. Define the workflow in declarative YAML; one open-source binary runs it with schedules, dependencies, retries, approvals, logs, and a Web UI. State lives in local files. No external database, no message broker, no framework.

Your business logic stays where it is: scripts, containers, and AI agents in any language. The workflow file points at them. No decorators, no SDK, no rewrite. The same engine runs AI: coding agents as steps, LLM calls as steps, and a human approval gate before anything ships.

## The whole idea

Your repo already has the logic:

```text
scripts/extract.py
scripts/build-report.sh
```

Add one file that holds only the structure:

```yaml
schedule: "0 2 * * *"

steps:
  - id: extract
    run: python scripts/extract.py
    retry_policy:
      limit: 3
      interval_sec: 30

  - id: report
    run: ./scripts/build-report.sh
    depends: extract
```

That is the entire integration. No imports, no decorators, nothing rewritten. Delete the YAML and your scripts still run. Keep it, and every run gets a dependency graph, retries, per-step logs, history, and a Web UI.

## Run it

Install on macOS or Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/dagucloud/dagu/main/scripts/installer.sh | bash
```

On Windows, run this in PowerShell:

```powershell
irm https://raw.githubusercontent.com/dagucloud/dagu/main/scripts/installer.ps1 | iex
```

Prefer a pinned version or a different method? See [releases](https://github.com/dagucloud/dagu/releases) and [all installation options](https://docs.dagu.sh/getting-started/installation/) for Docker, Homebrew, npm, Kubernetes, and manual installs.

Start the scheduler and Web UI in the directory with your YAML:

```sh
dagu start-all --dags .
```

Open <http://localhost:8080> to watch runs, read step logs, and browse history. Running Dagu as a persistent or shared service? Review [server configuration](https://docs.dagu.sh/server-admin/configuration) and [authentication](https://docs.dagu.sh/server-admin/authentication/) before exposing it beyond localhost.

## Why not cron, Airflow, or Temporal

- cron runs commands, but there are no dependencies, retries, or history.
- Airflow orchestrates, but you operate a platform (scheduler, metadata database, workers, a Python environment), and your jobs become framework code.
- Temporal gives durable execution, but your business logic moves into its SDK and programming model.

Dagu's position: workflow structure is configuration, not code. It belongs in a file next to your scripts, and the engine that runs it should be one process you can ignore:

```text
Traditional orchestrator              Dagu

Web server                            dagu start-all
Scheduler                             ├── Web UI
Workers                               ├── Scheduler
Metadata database                     ├── Executor
Message broker                        └── Local file-backed state
Python environment + plugins
Upgrades with schema migrations       Upgrade: replace one binary
```

Everything on the right is one process that idles under 100 MB of RAM. If you run Airflow to schedule what is ultimately a list of shell commands and containers, this is the same orchestration without the platform underneath it.

All of the orchestration. None of the platform.

## Give a step to an AI agent

Steps are not limited to shell commands. This workflow has a coding agent review the latest commit every night, and a human approves before the result goes anywhere:

```yaml
schedule: "0 2 * * *"
working_dir: .

secrets:
  - name: OPENROUTER_API_KEY
    provider: env
    key: OPENROUTER_API_KEY

steps:
  - id: review
    action: harness.run
    with:
      provider: pi
      model: openrouter/deepseek/deepseek-v4-flash
      tools: read,bash
      prompt: |
        Review the most recent commit in this repository.
        Reply with: what changed, one risk, a verdict.
    output: REVIEW

  - id: approve
    action: human.task
    with:
      prompt: Publish this review?
    depends: review

  - id: publish
    run: echo "$REVIEW"
    depends: approve
```

To try it, install the [Pi coding agent](https://pi.dev) (`npm install -g --ignore-scripts @earendil-works/pi-coding-agent`), export an `OPENROUTER_API_KEY`, and drop the file into your DAGs directory. When the review is ready, the run pauses; open it in the Web UI and click Approve. The approved text prints in the `publish` step's log, and the whole exchange stays in the run history.

The same separation applies to AI: the agent is a CLI process, wired in by the file, not by an agent-framework SDK. The agent step calls whichever model provider you configure; this example sends the prompt to OpenRouter. Local models work through `provider: local`. Approvals and human tasks are part of the open-source engine; SSO, RBAC, and audit logs are in the [licensed tier](https://dagu.sh/pricing#self-host).

If Dagu is useful, click **Star** at the top of this page. It helps other developers find the project.

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

You can also open the [live demo](https://dagu-demo-f5e33d0e.dagu.sh) and sign in with username `demouser` and password `demouser`.

From the GitHub community:

> I started out down the Temporal path. Temporal is powerful, but if all you want is to dynamically chain agents, scripts, data processing, and ops tasks together, the whole stack can feel a bit heavy. Then I came across Dagu [...] It runs as a single binary, workflows are written in YAML, everything lives in local files, it ships with a web UI, and there's no extra DB or broker to stand up. [...] A nice surprise was harness.run, which lets you plug external coding agent CLIs straight into a workflow.

## What Dagu gives you

- Orchestration that is not a second job: one binary, local file state, upgrades by replacing the binary.
- Declarative YAML, reviewed and diffed in Git like any other file.
- No decorators, no annotations, no framework imports: logic stays in your languages.
- Structure separate from logic: the file declares order, retries, and approvals; your scripts do the work.
- Run AI where you run cron: agent CLI steps (Claude Code, Codex, Copilot, Pi, OpenCode), LLM steps, and human approval gates.
- Inspect live status, logs, and previous runs in the built-in Web UI.
- Start on one machine, then add queues or distributed workers if the workload grows.

## A step can be

| | | |
|---|---|---|
| A shell command | A Docker container | A Kubernetes Job |
| An SSH command | An HTTP request | A SQL query |
| Another DAG | A fan-out over a list | An LLM call (`chat.completion`) |
| A coding agent (`harness.run`) | A human approval (`human.task`) | One of 50+ built-in actions |

Browse the [workflow examples](https://docs.dagu.sh/writing-workflows/examples) or the [YAML reference](https://docs.dagu.sh/writing-workflows/yaml-specification) when you need them.

## Remote machines

Declare `ssh` once at the DAG level and every `run` step executes on that host. Machines that only have `sshd` get schedules, retries, approvals, and logs, with nothing installed on them:

```yaml
ssh:
  user: deploy
  host: web-1.internal
  key: ~/.ssh/deploy_key

steps:
  - id: health
    run: curl -f http://localhost:8080/health
    retry_policy:
      limit: 3
      interval_sec: 10

  - id: restart
    run: systemctl restart myapp
    depends: health
```

Combine it with `parallel` and the same pattern patches a fleet, one sub-DAG per host with separate logs and retries. See [SSH](https://docs.dagu.sh/step-types/ssh).

## LLM-directed workflows

With `type: controller`, steps become a catalog and `tasks` state the goals; an LLM decides which step runs next until the goals are met. The controller selects only from the steps you declare, records every decision, and runs under the same retries, logs, and controls as any other workflow. This example triages the machine it runs on, and works as-is with an OpenRouter API key:

```yaml
type: controller

secrets:
  - name: OPENROUTER_API_KEY
    provider: env
    key: OPENROUTER_API_KEY

llm:
  provider: openrouter
  model: deepseek/deepseek-v4-flash

steps:
  - name: disk
    description: Show filesystem usage.
    run: df -h
    output: DISK
  - name: load
    description: Show uptime and load average.
    run: uptime
    output: LOAD
  - name: processes
    description: List processes with CPU and memory usage.
    run: ps aux | head -20
  - name: summarize
    description: Write the health summary. Run last, after the checks.
    action: chat.completion
    with:
      prompt: |
        Summarize this machine's health in three sentences:
        ${DISK}
        ${LOAD}

tasks:
  - name: triage
    description: >
      Finished when the machine has been checked and a health summary has been
      written. Inspect processes only if disk or load looks unhealthy.
```

There is no fixed order: the controller picks probes, digs deeper only when something looks off, and ends by writing the summary. On a `schedule`, every decision it made is there in the run's timeline the next morning.

The `summarize` step uses `chat.completion`: a plain LLM call that works in any workflow, shares the workflow's `llm` config, and supports tool use.

See [Controller Workflows](https://docs.dagu.sh/writing-workflows/controller).

## Nested workflows

A step can run another DAG. Sub-DAGs can live in the same file after `---`, and `parallel` fans one sub-DAG out over a list of items. This example works as-is:

```yaml
steps:
  - id: check
    action: dag.run
    with:
      dag: probe
      params:
        url: ${ITEM}
    parallel:
      items:
        - https://example.com
        - https://example.org
        - https://example.net
      max_concurrent: 2

---
name: probe
params:
  - name: url
    type: string
steps:
  - id: fetch
    run: curl -fsS -o /dev/null -w '%{http_code}\n' "${params.url}"
```

Each URL becomes its own child run with separate logs, status, and retries. The same mechanism composes larger systems: shared DAGs in their own files, called from many parents.

See [Sub-DAGs](https://docs.dagu.sh/writing-workflows/sub-dags).

## Operate Dagu from your AI tools

The direction also reverses: AI tools can run Dagu. The MCP endpoint (`http://localhost:8080/mcp`) lets MCP clients inspect workflows, start and control runs, and read results.

MCP Apps hosts can render run-related `dagu_read` and `dagu_execute` results in an interactive inspector with step status, scheduler and per-step logs, refresh, stop, retry, and a link to the full run page in Dagu. Other MCP clients continue to receive the same text and structured results.

For workflow-authoring help in Claude Code, Codex, Gemini CLI, and other coding tools, install the Dagu skill:

```sh
gh skill install dagucloud/dagu dagu
```

See the [MCP guide](https://docs.dagu.sh/mcp/quickstart).

## Common uses

- Replace fragile cron chains while keeping the underlying scripts.
- Run coding agents on schedules with approval gates and full logs.
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
| Licensed self-hosted | Server and workers on your infrastructure, with a paid server license | Teams that need SSO, RBAC, audit logs, incident routing, additional API keys, and support; see [plans and pricing](https://dagu.sh/pricing#self-host) |

The same YAML works across these models. See [deployment models](https://docs.dagu.sh/overview/deployment-models) for the architecture, security boundaries, and setup details.

## Learn more

| Topic | Documentation |
|---|---|
| Install and first run | [Quickstart](https://docs.dagu.sh/getting-started/quickstart) |
| Your first AI workflow | [AI quickstart](https://docs.dagu.sh/getting-started/quickstart-ai) |
| Install on Windows | [PowerShell, Windows service, and manual install](https://docs.dagu.sh/getting-started/installation/windows) |
| Run with Docker | [Docker, Compose, volumes, and image tags](https://docs.dagu.sh/getting-started/installation/docker) |
| Configure a server | [Configuration files, environment variables, and precedence](https://docs.dagu.sh/server-admin/configuration) |
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
