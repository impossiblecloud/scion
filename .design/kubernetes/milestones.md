# Kubernetes Support Milestones

This document outlines the incremental milestones required to evolve the Scion Kubernetes runtime from its current "experimental/broken" state to a fully functional, production-ready environment.

Each milestone is designed to be independently testable ("QAable") and builds upon the previous one.

## Milestone 1: Basic Runtime Configuration & Connectivity

**Goal:** Ensure the Kubernetes runtime honors the basic `run` configuration provided by the CLI. Currently, the runtime ignores `Env`, `Image`, and other critical parameters.

See detailed design in `m1-design.md`.

**Current Flaws Addressed:**
- **Flaw #2:** Missing Configuration Propagation (`Env` variables ignored).
- **Flaw #3:** Credential Propagation (Partially - via Env vars).

**Tasks:**
1.  **Environment Propagation:** Modify `KubernetesRuntime.Run` and the `SandboxClaim` construction logic to serialize the `runCfg.Env` map (which includes API keys discovered by the harness) into the `SandboxClaim` (or underlying Pod spec).
    *   *Interim Solution:* If `SandboxClaim` CRD doesn't support generic Env, modify the `Client` to patch the created Pod or switch to managing `Pod` resources directly for this milestone.
2.  **Command Propagation:** Ensure the command string from the Harness (e.g., `gemini-cli run ...`) is actually executed by the remote container, overriding the default entrypoint if necessary.
3.  **Basic Auth Injection:** As a temporary measure, inject `GEMINI_API_KEY` / `ANTHROPIC_API_KEY` as plain Environment Variables to ensure the harness *can* run if it had the right files.

**Verification (Manual QA):**
- Run: `scion run --runtime kubernetes --env TEST_VAR=foo --image alpine "echo $TEST_VAR"`
- Success: Agent starts, logs show "foo", and status becomes "completed".

---

## Milestone 2: Identity & Context Projection (The "Snapshot" Fix)

**Goal:** Enable standard Harnesses (Gemini/Claude) to function by ensuring their required configuration files and home directory context are present in the remote container.

**Current Flaws Addressed:**
- **Flaw #1:** Incomplete Context Sync (HomeDir missing).
- **Flaw #5:** Hardcoded Paths.
- **Flaw #4:** Synchronous Sync Issues (Race conditions).

**Tasks:**
1.  **Home Directory Sync:** Extend `syncContext` to support multiple sources. It must explicitly `tar` relevant files from the local `AgentHome` (e.g., `.bashrc`, `.config/gcloud`, harness-specific settings) and stream them to the remote container's `$HOME`.
2.  **Wait-for-Init Logic:** Implement a "blocking start" mechanism. The main agent process (the Harness command) should not start until the context sync is complete.
    *   *Implementation Idea:* Override the container command to `tail -f /dev/null` initially, perform the `kubectl exec ... tar` upload, and *then* `kubectl exec` the actual harness command.
3.  **Workspace Path Configuration:** Respect `runCfg.Workspace` mapping instead of hardcoding `/workspace`.

**Verification (Manual QA):**
- Run: `scion start <agent-name>` (where agent uses a harness requiring a config file, e.g., `gemini-cli`).
- Success: The harness starts successfully without "config not found" errors.

---

## Milestone 3: SCM Integration (Git Clone on Start)

**Goal:** Transition from "uploading local workspace" to "cloning from source" for the project code, aligning with `@.design/kubernetes/scm.md`.

**Current Flaws Addressed:**
- **Kubernetes Challenge:** No Shared Filesystem.
- **Scalability:** Avoiding large tarball uploads for large repos.

**Tasks:**
1.  **Repository Detection:** Implement logic in `cmd/start.go` to detect the Git remote URL of the current grove.
2.  **Init Container Injection:**
    *   *Design Change:* Switch from `SandboxClaim` to `Pod` (or update `Sandbox` definition) to support an `initContainer`.
    *   Configure the init container to `git clone` the detected URL into an `EmptyDir` volume mounted at `/workspace`.
3.  **Credential Management (Basic):** Implement a method to copy the local user's Git credentials (PAT or SSH key) into a Kubernetes Secret and mount it for the Init Container.

**Verification (Manual QA):**
- Run: `scion start --runtime kubernetes` in a git repo.
- Success: Pod starts, `kubectl exec ls /workspace` shows the git repository content.

---

## Milestone 4: Interactive Synchronization

**Goal:** Restore the "local development" feel by allowing users to push/pull changes between their local machine and the remote agent.

**Current Flaws Addressed:**
- **Usability:** "I changed a file locally, but the remote agent doesn't see it."

**Tasks:**
1.  **Sync-To Command:** Implement `scion sync to <agent-name>`:
    *   Tars local workspace changes (diff against last sync or git status).
    *   Streams to remote `/workspace`.
2.  **Sync-From Command:** Implement `scion sync from <agent-name>`:
    *   Tars remote `/workspace` (or specific changed files).
    *   Unpacks to local directory.
3.  **Watch Mode (Optional):** A simple file watcher that triggers `sync to` on change.

**Verification (Manual QA):**
1.  Modify a file locally.
2.  Run `scion sync to <agent>`.
3.  Verify file change in remote Pod.

---

## Milestone 5: Production Hardening

**Goal:** Move from "Dev/Test" quality to "Production" quality, securing secrets and handling lifecycle events robustly.

**Current Flaws Addressed:**
- **Flaw #3:** Credential Propagation (Insecure Env Vars).
- **Flaw #4:** Race conditions.

**Tasks:**
1.  **Secret Management:** Replace Environment Variable injection (Milestone 1) with proper Kubernetes Secrets for API keys and Auth tokens.
2.  **Status Reconciliation:** Update `scion list` to accurately reflect K8s Pod status (Pending, Running, CrashLoopBackOff) rather than just local state.
3.  **Cleanup:** Ensure `scion delete` removes Secrets, ConfigMaps, and PVCs associated with the agent.

**Verification (Manual QA):**
- Inspect Pod: `kubectl get pod <agent> -o yaml`. Verify no API keys are visible in `spec.containers.env`.
- Run: `scion delete <agent>`. Verify all related K8s resources are gone.
