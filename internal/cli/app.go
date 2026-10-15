package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/lanyi1998/DNSlog-GO/internal/cliconfig"
	"github.com/lanyi1998/DNSlog-GO/pkg/sdk"
)

const (
	defaultRandomLength         = 8
	defaultVerificationTimeout  = 10 * time.Second
	defaultVerificationInterval = 500 * time.Millisecond
)

type application struct {
	stdout io.Writer
	stderr io.Writer
}

type verificationResult struct {
	Found bool `json:"found"`
}

type verificationOptions struct {
	Query    string
	Timeout  time.Duration
	Interval time.Duration
}

type tokenStatus struct {
	Valid     bool   `json:"valid"`
	Subdomain string `json:"subdomain"`
}

type operationResult struct {
	Success bool `json:"success"`
}

type generatedValue struct {
	Value string `json:"value"`
}

type configPathResult struct {
	Path string `json:"path"`
}

type configSetResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

// Run executes the CLI and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	app := application{stdout: stdout, stderr: stderr}
	return app.run(args)
}

func (a application) run(args []string) int {
	if len(args) == 0 {
		a.printUsage()
		return 2
	}

	command := args[0]
	if command == "help" || command == "--help" || command == "-h" {
		a.printUsage()
		return 0
	}
	if command == "config-path" {
		return a.printConfigPath(args[1:])
	}
	if command == "config" {
		return a.runConfig(args[1:])
	}
	if !isCommand(command) {
		return a.usageError("unknown command %q", command)
	}

	configuration, err := cliconfig.Load()
	if err != nil {
		return a.runtimeError(err)
	}
	client, err := sdk.NewDnsLogClient(configuration.BaseURL, configuration.Token)
	if err != nil {
		return a.runtimeError(fmt.Errorf("connect to DNSlog-GO: %w", err))
	}

	return a.runCommand(client, command, args[1:])
}

func (a application) runCommand(client *sdk.DnsLogClient, command string, args []string) int {
	switch command {
	case "verify-token":
		return a.verifyToken(client, args)
	case "get-dns":
		return a.getDns(client, args, false)
	case "get-dns-clear":
		return a.getDns(client, args, true)
	case "verify-dns":
		return a.verifyDns(client, args)
	case "bulk-verify-dns":
		return a.bulkVerifyDns(client, args)
	case "verify-http":
		return a.verifyHTTP(client, args)
	case "bulk-verify-http":
		return a.bulkVerifyHTTP(client, args)
	case "clean":
		return a.clean(client, args)
	case "set-a-record":
		return a.setARecord(client, args)
	case "set-txt-record":
		return a.setTXTRecord(client, args)
	case "random-domain":
		return a.randomDomain(client, args)
	case "random-ssrf-url":
		return a.randomSSRFURL(client, args)
	default:
		return a.usageError("unknown command %q", command)
	}
}

func (a application) verifyToken(client *sdk.DnsLogClient, args []string) int {
	if err := requireArgumentCount(args, 0); err != nil {
		return a.usageError("verify-token: %v", err)
	}
	return a.writeJSON(tokenStatus{Valid: true, Subdomain: client.Subdomain})
}

func (a application) getDns(client *sdk.DnsLogClient, args []string, clear bool) int {
	if err := requireArgumentCount(args, 0); err != nil {
		return a.usageError("get-dns: %v", err)
	}

	var (
		records []sdk.DnsRecord
		err     error
	)
	if clear {
		records, err = client.GetDnsDataAndClear()
	} else {
		records, err = client.GetDnsData()
	}
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(records)
}

func (a application) verifyDns(client *sdk.DnsLogClient, args []string) int {
	options, err := parseVerificationOptions(args)
	if err != nil {
		return a.usageError("verify-dns: %v", err)
	}
	found, err := pollVerification(options, func(ctx context.Context) (bool, error) {
		return client.VerifyDnsContext(ctx, options.Query)
	})
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(verificationResult{Found: found})
}

func (a application) bulkVerifyDns(client *sdk.DnsLogClient, args []string) int {
	if len(args) == 0 {
		return a.usageError("bulk-verify-dns: at least one domain is required")
	}
	found, err := client.BulkVerifyDns(args)
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(found)
}

func (a application) verifyHTTP(client *sdk.DnsLogClient, args []string) int {
	options, err := parseVerificationOptions(args)
	if err != nil {
		return a.usageError("verify-http: %v", err)
	}
	found, err := pollVerification(options, func(ctx context.Context) (bool, error) {
		return client.VerifyHttpContext(ctx, options.Query)
	})
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(verificationResult{Found: found})
}

func (a application) bulkVerifyHTTP(client *sdk.DnsLogClient, args []string) int {
	if len(args) == 0 {
		return a.usageError("bulk-verify-http: at least one path is required")
	}
	found, err := client.BulkVerifyHttp(args)
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(found)
}

func (a application) clean(client *sdk.DnsLogClient, args []string) int {
	if err := requireArgumentCount(args, 0); err != nil {
		return a.usageError("clean: %v", err)
	}
	if err := client.ClearRecords(); err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(operationResult{Success: true})
}

func (a application) setARecord(client *sdk.DnsLogClient, args []string) int {
	if err := requireArgumentCount(args, 2); err != nil {
		return a.usageError("set-a-record: %v", err)
	}
	if err := client.SetARecord(args[0], args[1]); err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(operationResult{Success: true})
}

func (a application) setTXTRecord(client *sdk.DnsLogClient, args []string) int {
	if err := requireArgumentCount(args, 2); err != nil {
		return a.usageError("set-txt-record: %v", err)
	}
	if err := client.SetTXTRecord(args[0], args[1]); err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(operationResult{Success: true})
}

func (a application) randomDomain(client *sdk.DnsLogClient, args []string) int {
	length, err := parseRandomLength(args)
	if err != nil {
		return a.usageError("random-domain: %v", err)
	}
	return a.writeJSON(generatedValue{Value: client.RandomSubDomain(length)})
}

func (a application) randomSSRFURL(client *sdk.DnsLogClient, args []string) int {
	length, err := parseRandomLength(args)
	if err != nil {
		return a.usageError("random-ssrf-url: %v", err)
	}
	return a.writeJSON(generatedValue{Value: client.RandomSSRFUrl(length)})
}

func (a application) printConfigPath(args []string) int {
	if err := requireArgumentCount(args, 0); err != nil {
		return a.usageError("config-path: %v", err)
	}
	path, err := cliconfig.Path()
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(configPathResult{Path: path})
}

func (a application) runConfig(args []string) int {
	if len(args) == 0 {
		return a.usageError("config: expected a subcommand")
	}
	switch args[0] {
	case "path":
		return a.printConfigPath(args[1:])
	case "set":
		return a.setConfig(args[1:])
	default:
		return a.usageError("config: unknown subcommand %q", args[0])
	}
}

func (a application) setConfig(args []string) int {
	updates, err := parseConfigUpdates(args)
	if err != nil {
		return a.usageError("config set: %v", err)
	}
	path, err := cliconfig.Save(updates)
	if err != nil {
		return a.runtimeError(err)
	}
	return a.writeJSON(configSetResult{Success: true, Path: path})
}

func (a application) writeJSON(value any) int {
	encoder := json.NewEncoder(a.stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return a.runtimeError(fmt.Errorf("write output: %w", err))
	}
	return 0
}

func (a application) usageError(format string, arguments ...any) int {
	fmt.Fprintf(a.stderr, "dnslog-cli: "+format+"\n", arguments...)
	fmt.Fprintln(a.stderr, "Run 'dnslog-cli help' for usage.")
	return 2
}

func (a application) runtimeError(err error) int {
	fmt.Fprintf(a.stderr, "dnslog-cli: %v\n", err)
	return 1
}

func (a application) printUsage() {
	fmt.Fprint(a.stdout, `Usage: dnslog-cli <command> [arguments]

Commands:
  verify-token                         Verify the configured token
  get-dns                              Get all interaction records
  get-dns-clear                        Get and clear interaction records
  verify-dns <domain> [options]        Check one DNS interaction
  bulk-verify-dns <domain>...          Check multiple DNS interactions
  verify-http <path-or-url> [options]  Check one HTTP interaction
  bulk-verify-http <path>...           Check multiple HTTP interactions
  clean                                Clear all interaction records
  set-a-record <domain> <ip>           Set an A record
  set-txt-record <domain> <text>       Set a TXT record
  random-domain [length]               Generate a DNS callback domain
  random-ssrf-url [length]             Generate an HTTP callback URL
  config set --url URL --token TOKEN  Save connection settings
  config path                         Print the configuration file path
  config-path                          Alias for 'config path'

Configuration precedence:
  DNSLOG_BASE_URL / DNSLOG_TOKEN environment variables
  ~/.config/dnslog-go/config.yaml

DNSLOG_CONFIG may override the configuration file path.
The config file is created with permissions 0600. A token passed on the
command line may be stored in shell history or briefly visible to local tools.

Verification options:
  --timeout DURATION                   Maximum polling time (default 10s)
  --interval DURATION                  Delay between checks (default 500ms)

The first check runs immediately. Use --timeout 0 to disable polling.
`)
}

func isCommand(command string) bool {
	switch command {
	case "verify-token", "get-dns", "get-dns-clear", "verify-dns",
		"bulk-verify-dns", "verify-http", "bulk-verify-http", "clean",
		"set-a-record", "set-txt-record", "random-domain", "random-ssrf-url":
		return true
	default:
		return false
	}
}

func requireArgumentCount(args []string, expected int) error {
	if len(args) != expected {
		return fmt.Errorf("expected %d argument(s), got %d", expected, len(args))
	}
	return nil
}

func parseRandomLength(args []string) (int, error) {
	if len(args) == 0 {
		return defaultRandomLength, nil
	}
	if len(args) > 1 {
		return 0, fmt.Errorf("expected at most one length argument")
	}
	length, err := strconv.Atoi(args[0])
	if err != nil || length < 1 || length > 63 {
		return 0, fmt.Errorf("length must be an integer between 1 and 63")
	}
	return length, nil
}

func parseVerificationOptions(args []string) (verificationOptions, error) {
	options := verificationOptions{
		Timeout:  defaultVerificationTimeout,
		Interval: defaultVerificationInterval,
	}
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "--timeout":
			value, nextIndex, err := optionValue(args, index, "--timeout")
			if err != nil {
				return verificationOptions{}, err
			}
			options.Timeout, err = parseDuration(value, "--timeout", true)
			if err != nil {
				return verificationOptions{}, err
			}
			index = nextIndex
		case strings.HasPrefix(argument, "--timeout="):
			value := strings.TrimPrefix(argument, "--timeout=")
			var err error
			options.Timeout, err = parseDuration(value, "--timeout", true)
			if err != nil {
				return verificationOptions{}, err
			}
		case argument == "--interval":
			value, nextIndex, err := optionValue(args, index, "--interval")
			if err != nil {
				return verificationOptions{}, err
			}
			options.Interval, err = parseDuration(value, "--interval", false)
			if err != nil {
				return verificationOptions{}, err
			}
			index = nextIndex
		case strings.HasPrefix(argument, "--interval="):
			value := strings.TrimPrefix(argument, "--interval=")
			var err error
			options.Interval, err = parseDuration(value, "--interval", false)
			if err != nil {
				return verificationOptions{}, err
			}
		case strings.HasPrefix(argument, "-"):
			return verificationOptions{}, fmt.Errorf("unknown option %q", argument)
		case options.Query == "":
			options.Query = argument
		default:
			return verificationOptions{}, fmt.Errorf("expected one query, got %q", argument)
		}
	}
	if options.Query == "" {
		return verificationOptions{}, fmt.Errorf("query is required")
	}
	return options, nil
}

func parseDuration(value, option string, allowZero bool) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration < 0 || (!allowZero && duration == 0) {
		if allowZero {
			return 0, fmt.Errorf("%s must be a non-negative duration such as 10s", option)
		}
		return 0, fmt.Errorf("%s must be a positive duration such as 500ms", option)
	}
	return duration, nil
}

func pollVerification(
	options verificationOptions,
	verify func(context.Context) (bool, error),
) (bool, error) {
	if options.Timeout == 0 {
		return verify(context.Background())
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()
	for {
		if ctx.Err() != nil {
			return false, nil
		}
		found, err := verify(ctx)
		if err != nil {
			return false, err
		}
		if found {
			return true, nil
		}

		timer := time.NewTimer(options.Interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return false, nil
		case <-timer.C:
		}
	}
}

func parseConfigUpdates(args []string) (cliconfig.Updates, error) {
	var updates cliconfig.Updates
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "--url":
			value, nextIndex, err := optionValue(args, index, "--url")
			if err != nil {
				return cliconfig.Updates{}, err
			}
			updates.BaseURL = &value
			index = nextIndex
		case strings.HasPrefix(argument, "--url="):
			value := strings.TrimPrefix(argument, "--url=")
			if value == "" {
				return cliconfig.Updates{}, fmt.Errorf("--url cannot be empty")
			}
			updates.BaseURL = &value
		case argument == "--token":
			value, nextIndex, err := optionValue(args, index, "--token")
			if err != nil {
				return cliconfig.Updates{}, err
			}
			updates.Token = &value
			index = nextIndex
		case strings.HasPrefix(argument, "--token="):
			value := strings.TrimPrefix(argument, "--token=")
			if value == "" {
				return cliconfig.Updates{}, fmt.Errorf("--token cannot be empty")
			}
			updates.Token = &value
		default:
			return cliconfig.Updates{}, fmt.Errorf("unknown option %q", argument)
		}
	}
	if updates.BaseURL == nil && updates.Token == nil {
		return cliconfig.Updates{}, fmt.Errorf("provide --url, --token, or both")
	}
	return updates, nil
}

func optionValue(args []string, index int, option string) (string, int, error) {
	valueIndex := index + 1
	if valueIndex >= len(args) || args[valueIndex] == "" {
		return "", index, fmt.Errorf("%s requires a value", option)
	}
	return args[valueIndex], valueIndex, nil
}
