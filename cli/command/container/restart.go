package container

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type restartOptions struct {
	nSeconds        int
	nSecondsChanged bool
	checkpoint    string
	checkpointDir string

	containers []string
}

// NewRestartCommand creates a new cobra.Command for `docker restart`
func NewRestartCommand(dockerCli command.Cli) *cobra.Command {
	var opts restartOptions

	cmd := &cobra.Command{
		Use:   "restart [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Restart one or more containers",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.containers = args
			opts.nSecondsChanged = cmd.Flags().Changed("time")
			return runRestart(dockerCli, &opts)
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.nSeconds, "time", "t", 10, "Seconds to wait for stop before killing the container")

	flags.StringVar(&opts.checkpoint, "checkpoint", "", "Restore from this checkpoint")
	flags.SetAnnotation("checkpoint", "experimental", nil)
	flags.SetAnnotation("checkpoint", "ostype", []string{"linux"})
	flags.StringVar(&opts.checkpointDir, "checkpoint-dir", "", "Use a custom checkpoint storage directory")
	flags.SetAnnotation("checkpoint-dir", "experimental", nil)
	flags.SetAnnotation("checkpoint-dir", "ostype", []string{"linux"})
	
	return cmd
}

func runRestart(dockerCli command.Cli, opts *restartOptions) error {
	ctx := context.Background()
	var errs []string
	var timeout *time.Duration
	if opts.nSecondsChanged {
		timeoutValue := time.Duration(opts.nSeconds) * time.Second
		timeout = &timeoutValue
	}

	for _, name := range opts.containers {
		startOptions := types.ContainerStartOptions{
			CheckpointID:  opts.checkpoint,
			CheckpointDir: opts.checkpointDir,
		}
		if err := dockerCli.Client().ContainerRestart(ctx, name, timeout, startOptions); err != nil {
			errs = append(errs, err.Error())
			continue
		}
		fmt.Fprintln(dockerCli.Out(), name)
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
