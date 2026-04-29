package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	sanitizer "github.com/joekrill/subsonic-sanitizer"
	"github.com/urfave/cli/v3"
)

var (
	logLevels = []string{slog.LevelDebug.String(), slog.LevelInfo.String(), slog.LevelWarn.String(), slog.LevelError.String()}
)

var Server = &cli.Command{
	Name:    "subsonic-sanitizer",
	Usage:   "Run the sanitizing proxy server",
	Version: sanitizer.Version,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "target-url",
			Aliases:  []string{"t"},
			Usage:    "the target URL to proxy reqquests to",
			Required: true,
			Sources:  cli.EnvVars("TARGET_URL"),
			Validator: func(v string) error {
				_, err := url.Parse(v)
				if err != nil {
					return err
				}
				return nil
			},
		},
		&cli.IntFlag{
			Name:    "listen-port",
			Aliases: []string{"p"},
			Usage:   "port number to listen for requests on",
			Sources: cli.EnvVars("LISTEN_PORT"),
			Value:   3000,
		},
		&cli.StringFlag{
			Name:    "listen-host",
			Aliases: []string{"h"},
			Usage:   "the host to listen for requests on",
			Sources: cli.EnvVars("LISTEN_HOST"),
		},
		&cli.StringFlag{
			Name:    "log-level",
			Usage:   "the minimum log level to output (info|debug|warn|error)",
			Sources: cli.EnvVars("LOG_LEVEL"),
			Value:   "info",
			Validator: func(v string) error {
				var level slog.Level
				var err = level.UnmarshalText([]byte(v))

				if err != nil {
					return fmt.Errorf("log-level must be one of: %s", strings.Join(logLevels, "|"))
				}
				return nil
			},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var logLevel slog.Level
		var err = logLevel.UnmarshalText([]byte(cmd.String("log-level")))
		if err != nil {
			return cli.Exit(fmt.Errorf("invalid log level '%s': %w", cmd.String("log-level"), err), 1)
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
		logger.Debug("initializing", "version", cmd.Version, "name", cmd.Name)
		logger.Debug("set log-level", "log-level", logLevel)

		targetUrl, err := url.Parse(cmd.String("target-url"))
		if err != nil {
			return cli.Exit(fmt.Errorf("invalid target URL '%s': %w", cmd.String("target-url"), err), 1)
		}
		logger.Debug("set target-url", "target-url", targetUrl)

		options := []sanitizer.SanitizerProxyConfig{
			sanitizer.WithLogger(logger),
		}

		proxy, err := sanitizer.NewSanitizerProxy(targetUrl, options...)
		if err != nil {
			return cli.Exit(fmt.Errorf("failed to initialize proxy: %w", err), 1)
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})

		addr := fmt.Sprintf("%s:%d", cmd.String("listen-host"), cmd.Int("listen-port"))
		logger.Info("server starting", "addr", addr)
		err = http.ListenAndServe(addr, nil)

		if err != nil {
			logger.Error(err.Error())
			return cli.Exit(err.Error(), 1)
		}

		logger.Info("server stopped", "addr", addr)

		return nil
	},
}

func main() {
	if err := Server.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
