---
name: dnslog-go
description: Use the DNSlog-GO CLI to generate callback domains or URLs, inspect and verify DNS/HTTP interactions, clear interaction data, and manage A or TXT records. Use only for DNSlog-GO services, not for general DNS lookup or unrelated network diagnostics.
---

# DNSlog-GO

Use `dnslog-cli` as the only interface to DNSlog-GO. The CLI reads credentials from `DNSLOG_BASE_URL` and `DNSLOG_TOKEN`, or from `~/.config/dnslog-go/config.yaml`; environment variables take precedence.

## Locate the CLI

Resolve the executable in this order:

1. Use `dnslog-cli` when it is available in `PATH`.
2. In a DNSlog-GO source checkout, use `./build/dnslog-cli` when it exists.
3. If the source checkout contains `cmd/dnslog-cli`, build it with `make build-cli`, then use `./build/dnslog-cli`.

Do not recreate API requests with `curl` while the CLI is available.

## Protect credentials

- Never pass the token to ordinary API commands. `config set` is the only command that accepts a token argument; use it only when the user explicitly asks to persist configuration.
- Never print environment variables or read the configuration file merely to discover the token.
- Do not echo, log, summarize, or return the token.
- Use `dnslog-cli config path` when the user needs the effective configuration-file location.

If the CLI reports that `DNSLOG_BASE_URL`, `DNSLOG_TOKEN`, `base_url`, or `token` is missing:

1. Stop the requested DNSlog-GO operation; do not fall back to direct HTTP requests.
2. Tell the user to configure the CLI locally with:

   ```bash
   dnslog-cli config set --url "<DNSlog-GO URL>" --token "<TOKEN>"
   ```

3. Do not ask the user to paste the token into chat and do not run the command on their behalf unless they explicitly request it and have already provided the values through an authorized channel.
4. After the user says configuration is complete, run `dnslog-cli verify-token` once, then continue the original operation.

Run `dnslog-cli verify-token` before the requested operation when connectivity or authentication has not yet been established. Do not repeat it before every command in the same task.

## Select a command

Read-only operations:

- `dnslog-cli verify-token`: verify authentication and return the assigned subdomain.
- `dnslog-cli get-dns`: return all current interaction records.
- `dnslog-cli verify-dns <domain> [--timeout <duration>] [--interval <duration>]`: poll until a DNS interaction exists or the timeout expires.
- `dnslog-cli bulk-verify-dns <domain>...`: return matching DNS domains.
- `dnslog-cli verify-http <path-or-url> [--timeout <duration>] [--interval <duration>]`: poll until an HTTP interaction exists or the timeout expires.
- `dnslog-cli bulk-verify-http <path>...`: return matching HTTP paths.
- `dnslog-cli random-domain [length]`: generate a DNS callback domain.
- `dnslog-cli random-ssrf-url [length]`: generate an HTTP callback URL.

State-changing operations require the user's request to clearly authorize the change:

- `dnslog-cli config set --url <url> --token <token>`: persist connection settings with file permissions `0600`. A token provided this way may be visible in shell history or the local process list.
- `dnslog-cli get-dns-clear`: return current records and then remove them.
- `dnslog-cli clean`: remove all current interaction records.
- `dnslog-cli set-a-record <domain> <ip>`: set an A record.
- `dnslog-cli set-txt-record <domain> <text>`: set a TXT record.

Before `clean` or `get-dns-clear`, make sure the user understands that stored interaction data will be removed. Do not substitute either command for `get-dns`.

## Handle results

The CLI writes machine-readable JSON to standard output and errors to standard error.

- Parse JSON output instead of inferring success from prose.
- For verification commands, use the `found` boolean.
- For generated values, use the `value` field exactly as returned.
- For record queries, preserve record type, subdomain, source IP, location, request data, and timestamp when relevant to the user's request.
- Summarize results concisely unless the user asks for raw JSON.
- Treat a nonzero exit code as failure and report the error without exposing configuration values.

Single DNS and HTTP verification commands perform the first check immediately, then poll every `500ms` for up to `10s` by default. Use the defaults unless the user requests a different wait period. Pass the requested duration through `--timeout` and choose a reasonable `--interval`; use `--timeout 0` only when the user explicitly wants a single immediate check. Do not implement a separate shell polling loop around these commands, and do not clear records unless separately authorized.
