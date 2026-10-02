# Atmos Dogfood Bugs

This file tracks bugs and dogfood gaps encountered while wiring this repository
to Atmos native CI, emulator fixtures, source-provisioned components, and
Terraform test variables.

The point of this branch is to dogfood Atmos. Workarounds below should not be
treated as final design decisions; they identify places Atmos or the Atmos
GitHub Actions integration needs to be fixed.

## Current release selection

As of 2026-10-02, local tooling and this file's revalidation target Atmos
`1.230.1` (`.tool-versions` and the `cache` action pins in
`.github/workflows/*.y*ml` have been bumped accordingly). Note that the
Homebrew `atmos` on a dev machine does not follow `.tool-versions`; use
`atmos --use-version=1.230.1 ...` to be sure. The live GitHub Actions
`ATMOS_VERSION` repository variable is still `1.226.0` pending a separate,
explicit bump.

## Status checked against Atmos 1.230.1

Checked on 2026-10-02 with `atmos --use-version=1.230.1` (host) and
`ghcr.io/cloudposse/atmos:1.230.1` (container), same isolated scratch-clone
methodology. No change from 1.229.0: the `terraform test --ci` summary/JUnit
fix (was #10) is still in place — `app.junit.xml` still reports `tests="1"`
with the real testcase `applies_ecs_service_against_emulator` — and the
`terraform/stacks/fixtures.yaml` identity-scoping fix still passes with a
`dev` emulator running concurrently. The 1.230.0/1.230.1 release notes don't
mention anything in the emulator-networking or CI-summary areas. One item
remains open, tracked below.

## Status checked against Atmos 1.229.0

Checked on 2026-09-19 with `atmos --use-version=1.229.0` (host) and
`ghcr.io/cloudposse/atmos:1.229.0` (container), using an isolated
`git clone --local` scratch copy of this worktree and the local Docker daemon
(no real AWS credentials).

Every bug previously tracked here has been confirmed fixed by direct local
reproduction and removed from this file, with one exception (below). Fixed in
this pass:

- **`terraform test --ci` summary/JUnit too sparse (was #10).** Fixed by
  cloudposse/atmos#3082 ("test summary/JUnit dropped OpenTofu runs and late
  diagnostics"), included in 1.229.0. OpenTofu's `-json` stream has no
  `progress` field, so the parser discarded every run event. On 1.229.0 the
  fixture test now produces `app.junit.xml` with `tests="1"`, testsuite
  `ecs_task.tftest.hcl`, testcase `applies_ecs_service_against_emulator`, and
  the step summary has a real results table plus a per-file "Detailed test
  results" table. (1.226.1-1.228.0 all still reported `tests="0"`.)

Repository-side hardening made during this pass (not an Atmos bug, but
surfaced while testing):

- `terraform/stacks/fixtures.yaml`: the `local-aws` identity used
  `emulator: aws`, which is ambiguous whenever any other stack's `aws`
  emulator is running (e.g. a `dev` emulator left up from another workspace):
  `emulator identity is ambiguous: emulator "aws" matches multiple running
  instances (dev/aws, fixtures/aws)`. This fails
  `atmos terraform test app -s fixtures` locally on 1.228.0 and 1.229.0
  alike. The identity now names its instance explicitly
  (`emulator: fixtures/aws`); verified passing on 1.229.0 with a `dev`
  emulator still running.

One item remains open, tracked below.

## 8. Emulator identity loopback endpoints fail inside GitHub job containers

**Status in 1.230.1: locally fixed, GHA confirmation still pending.**
Unchanged since 1.226.1 (re-checked on 1.226.1, 1.227.0, 1.228.0, 1.229.0,
1.230.1). The fix is cloudposse/atmos#2960 (built on #2942). Running
`ghcr.io/cloudposse/atmos:1.230.1` with the host Docker socket mounted, `atmos
emulator up aws -s fixtures` reports `emulator aws is up at
http://fixtures-aws:4566` (a container-network alias, not a loopback/gateway
address) and `curl http://fixtures-aws:4566/` from inside the same container
returns `HTTP 200`.

Caveat: this was validated against Docker Desktop's local VM networking, not
an actual GitHub-hosted Actions job-container network (which uses a
runner-managed Docker network with its own semantics). It cannot be marked
fully confirmed until a live Actions run of the `e2e` job on this repo passes.
When it does, drop this section.

**Observed behavior**

Running the e2e job as a GitHub Actions job container with:

```yaml
container:
  image: ghcr.io/cloudposse/atmos:${{ vars.ATMOS_VERSION }}
```

required installing a Docker CLI in the container so `atmos emulator up aws`
could use the mounted host Docker socket. After that, Atmos started the AWS
emulator successfully, but Terraform failed against the injected endpoint:

```text
Post "http://127.0.0.1:32768/": dial tcp 127.0.0.1:32768: connect: connection refused
```

The emulator container was started through the host Docker daemon and published
its port on the host. Inside the GitHub job container, `127.0.0.1` is the job
container itself, not the host where the emulator port is listening.

**Why this blocks dogfooding**

The intended CI shape was the Atmos container image approach without installing
Atmos on the runner. That works for ordinary Atmos commands, but emulator
identity injection currently produces a host-loopback endpoint that is invalid
from a sibling job container.

**Desired use case**

An app repository should be able to run Atmos from
`ghcr.io/cloudposse/atmos:<version>` in GitHub Actions, mount the host Docker
socket so Atmos can start emulator containers, and then run:

```text
atmos terraform test app -s fixtures --ci
```

That should work without wrapping the command in `docker run --network host`.
Terraform/OpenTofu should be able to reach the emulator endpoint that Atmos
injects into auth/profile/provider configuration.

**Minimal reproduction shape**

1. Run a GitHub Actions job from an Atmos container image.
2. Install or provide a Docker CLI inside that job container.
3. Mount the host Docker socket into the job container.
4. Run `atmos terraform test app -s fixtures --ci`.
5. Atmos starts the AWS emulator through the host Docker daemon.
6. Atmos injects `http://127.0.0.1:<published-port>` as the emulator endpoint.
7. Terraform fails because `127.0.0.1` resolves to the job container, not the
   Docker host where the emulator port was published.

**Current workaround (believed no longer needed, pending GHA confirmation)**

`docker run --network host` is no longer required in local testing — Atmos
1.226.1 injects a container-network alias instead of a loopback address. Keep
the workflow's current approach until this is confirmed against a real
GitHub Actions job container, then drop `--network host` if used.

**Expected fix (appears implemented, pending GHA confirmation)**

Atmos should make emulator endpoints container-context aware so the injected
endpoint is reachable from the process running Terraform:

- Host-native Atmos should keep using `127.0.0.1:<published-host-port>`.
- Containerized Atmos should prefer connecting emulator containers to the
  current Atmos/job container network, then inject a network alias plus the
  emulator container port. **Confirmed locally**: `atmos emulator up aws -s
  fixtures` inside `ghcr.io/cloudposse/atmos:1.226.1` (host Docker socket
  mounted) now reports `emulator aws is up at http://fixtures-aws:4566`, and
  that alias is reachable from inside the same container.
- If sharing the current container network is not possible, Atmos should fall
  back to a host-gateway reachable address plus the published host port.

