# Narrated demo video

About 75 seconds. Screen recording with your face in a corner bubble. Show the payoff in the first 10 seconds, and burn in captions because most people watch muted.

Before recording, run `./demo/run.sh` and leave the terminal on the printed link. Clear the browser history so the URL bar doesn't suggest anything. Set the browser zoom to 125% so the text is readable in a small player.

## 1. Hook (0:00–0:10)

**Screen:** the snapshot page is already open on the `payment-api` logs, with `cannot connect to postgres at db:5432` visible.

**Say:** "This pod is crash-looping in my cluster. You're looking at its logs, its events and its status, and you don't have access to my cluster. All I sent you was a link."

## 2. The problem (0:10–0:25)

**Screen:** your face bubble, a bit bigger, or a Slack thread with pasted `kubectl` output.

**Say:** "When something breaks, people paste `kubectl describe` into Slack, attach log files to tickets, or screen-share for an hour. Whoever helps you only sees what you remembered to copy."

## 3. One command (0:25–0:40)

**Screen:** the terminal. Run `kubin create --server http://localhost:8080`, or scroll back to the output of `run.sh`.

**Say:** "Kubin is a pastebin for clusters. `kubin create` captures the cluster: pods, events and the last lines of every container log. It uploads that and prints a link. Secrets and env var values stay out of the snapshot."

## 4. Click around (0:40–1:05)

**Screen:**
1. Open the link. Point at "1 pod needs attention".
2. Click `payment-api`. The page jumps to the logs.
3. Click "Previous run" and show it's the same error, so it isn't a fluke.
4. Scroll to the BackOff warning in Events.
5. Scroll to the YAML.

**Say:** "Broken things come first. Click the pod and you land on its logs: it can't reach Postgres. The previous run says the same thing, the events show Kubernetes backing off, and the full spec is below. That's the whole debugging session, frozen at the moment I ran the command."

## 5. Self-host and close (1:05–1:15)

**Screen:** the README, showing the three commands.

**Say:** "It's one container. `docker run` with a volume and you have your own. Clone the repo, run `demo/run.sh`, and you get exactly this. Links are in the description."

## Don't claim

These aren't built yet, so the video shouldn't promise them: other workloads, services or nodes; choosing what to capture; auth or private links; expiry; a hosted instance. See the "What does not" section of the README.
