# Narrated demo video

About 75 seconds. Screen recording with your face in a corner bubble. Show the broken cluster in the first 10 seconds, and burn in captions because most people watch muted.

Before recording, run `./demo/setup.sh`. It builds the cluster and the server but takes no snapshot. Run the `export` line it prints in a second terminal, so `kubin` is on the path, then open `k9s -n checkout` in the first one and wait until it shows `api` in CrashLoopBackOff, `postgres-0` Pending and `worker` failed. Have a Linear ticket open in a browser tab with the title and description empty. Clear the browser history so the URL bar doesn't suggest anything. Set the browser zoom to 125% so the text is readable in a small player.

## 1. Is this familiar? (0:00–0:12)

**Screen:** k9s on namespace `checkout`. Pods listed: `api` in CrashLoopBackOff, `postgres-0` Pending. Press `:jobs` or switch to kubectl to show `worker` failed.

**Say:** "Is this familiar? Something is broken in your cluster. The API is crash-looping, the database is pending, and a job failed."

## 2. The ticket (0:12–0:28)

**Screen:** the empty Linear ticket. Hover over the description. Optionally flash a scrolling wall of `kubectl describe` output.

**Say:** "Now you need a ticket, and whoever picks it up needs logs. So you copy `kubectl describe`, then a huge log dump, and paste it all in. Half of it is noise, and you still forgot something."

## 3. One command (0:28–0:43)

**Screen:** the terminal. Run `kubin create`. The link prints within a few seconds.

**Say:** "Instead, one command. `kubin create` captures everything in the cluster that isn't secret and prints a link, in seconds. Secrets and env var values stay out of the snapshot."

## 4. Paste the link (0:43–0:50)

**Screen:** copy the link, paste it into the Linear ticket, and save it.

**Say:** "Paste the link into the ticket. That's the whole report."

## 5. A quick look (0:50–1:08)

**Screen:**
1. Open the link from the ticket. The overview shows the problems first.
2. Click the `api` pod. The page lands on its logs.
3. Point at the kind sidebar on the left. Stop there.

**Say:** "Whoever opens the ticket sees what's broken first. Click the pod and you're in its logs. The rest of the cluster is in the sidebar, by kind. They never needed access to my cluster."

## 6. Self-host and close (1:08–1:15)

**Screen:** the README, showing the self-hosting commands.

**Say:** "It's one container. `docker run` with a volume and you have your own. Links are in the description."

## Don't claim

These aren't built yet, so the video shouldn't promise them: authentication or private links; expiry; a hosted instance; a published release image. Don't say it explains why something broke. See the "What does not" section of the README.
