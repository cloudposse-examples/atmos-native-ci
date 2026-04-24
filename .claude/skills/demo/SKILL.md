---
name: demo
description: Bump the COLOR environment variable on the app (alternating blue ↔ green) and open a PR with the `deploy` label. Use this when the user says "bump the color", "demo", "bump the demo", or otherwise wants to trigger a demo deploy via the CI/CD pipeline.
---

# Demo: Bump color and open a deploy PR

Performs the standard demo rotation that exercises the CI/CD pipeline: flip the `COLOR` env var on the app container, commit on a new branch, open a PR against `main`, and apply the `deploy` label so the pipeline picks it up.

## Steps

1. **Rename the current branch** to something concrete under 30 characters, prefixed `osterman/`. Example: `osterman/bump-color-green` or `osterman/bump-color-blue` — match the target color.
   ```sh
   git branch -m osterman/bump-color-<target>
   ```

2. **Read** `terraform/stacks/defaults/app.yaml` and find the `COLOR` env var (currently around line 19–20):
   ```yaml
   environment:
     - name: COLOR
       value: <current>
   ```

3. **Flip the value**: if current is `blue`, change to `green`; if `green`, change to `blue`. The repo alternates between these two — do not introduce other colors without the user asking.

4. **Commit** using the established convention:
   ```sh
   git add terraform/stacks/defaults/app.yaml
   git commit -m "Change environment variable COLOR from <old> to <new>"
   ```

5. **Push** and set upstream:
   ```sh
   git push -u origin <branch>
   ```

6. **Open the PR against `main` with the `deploy` label already attached**, using the repo's standard `what` / `why` / `references` body:
   ```sh
   gh pr create --base main --label deploy \
     --title "Change environment variable COLOR from <old> to <new>" \
     --body "$(cat <<'EOF'
   ## what

   - Change the deployment background color from <old> to <green> by updating the `COLOR` environment variable in the default app stack configuration.

   ## why

   - Flip the deployment color for a blue/green deployment switch.

   ## references

   - N/A
   EOF
   )"
   ```
   The `--label deploy` flag is required — this is what triggers the deploy pipeline. If the label is missing, the PR will sit idle.

   **Body format is mandatory.** Every demo PR in this repo uses the three H2 sections `## what`, `## why`, `## references` with bullet points under each. Do not substitute a "Summary / Test plan" template or any other format — match what prior PRs (e.g. #39, #42) used. If there are no references, write `- N/A` under that section.

7. **Return the PR URL** to the user.

## Notes

- Only `terraform/stacks/defaults/app.yaml` changes. The Go app at `app/main.go` reads `COLOR` from the environment — no code change.
- Do not amend or force-push; always create a fresh commit and new branch.
- If the user has uncommitted work, stop and ask before running `git branch -m`.
- If the `deploy` label is missing from the repo, create it first: `gh label create deploy --color 8F24A3`.
