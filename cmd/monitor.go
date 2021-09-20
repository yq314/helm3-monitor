package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"strconv"
)

var (
	options = new(cmdOptions)
	rollbackDone = errors.New("Rollback done")
)

type cmdOptions struct {
	dryRun           bool
	disableHooks     bool
	force            bool
	interval         int64
	rollbackTimeout  int64
	timeout          int64
	wait             bool
	errorOnRollback  bool
	kubeConfigFile   string
	kubeContext      string
	debug            bool
}

const monitorDesc = `
This command monitor a release by querying Prometheus at a
given interval and take care of rolling back to the previous version if the
query return a non-empty result.
`

func prettyError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(status.Convert(err).Message())
}

func debug(format string, v ...interface{}) {
	if options.debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Printf(format, v...)
	}
}

func NewRootCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "helm3-monitor",
		Short:        "Monitor helm3 release and rollback automatically",
		Long:         monitorDesc,
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("no arguments accepted")
			}
			return nil
		},
	}

	flags := cmd.PersistentFlags()
	flags.BoolVar(&options.dryRun, "dry-run", false, "simulate a rollback if triggered by query result")
	flags.BoolVar(&options.disableHooks, "no-hooks", false, "prevent hooks from running during rollback")
	flags.BoolVar(&options.force, "force", false, "force resource update through delete/recreate if needed")
	flags.Int64VarP(&options.interval, "interval", "i", 10, "time in seconds between each query")
	flags.Int64Var(&options.rollbackTimeout, "rollback-timeout", 300, "time in seconds to wait for any individual Kubernetes operation during the rollback (like Jobs for hooks)")
	flags.Int64Var(&options.timeout, "timeout", 300, "time in seconds to wait before assuming a monitoring action is successfull")
	flags.BoolVar(&options.wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment are in a ready state before marking a rollback as successful. It will wait for as long as --rollback-timeout")
	flags.BoolVarP(&options.errorOnRollback, "error-on-rollback", "e", true, "returns error on successful rollback, so we can fail CI")
	flags.StringVar(&options.kubeConfigFile, "kubeconfig", "", "path to the kubeconfig file")
	flags.StringVar(&options.kubeContext, "kube-context", options.kubeContext, "name of the kubeconfig context to use")

	if ctx := os.Getenv("HELM_KUBECONTEXT"); ctx != "" {
		options.kubeContext = ctx
	}

	if ctx := os.Getenv("KUBECONFIG"); ctx != "" {
		options.kubeConfigFile = ctx
	}

	if ctx := os.Getenv("HELM_DEBUG"); ctx != "" {
		options.debug, _ = strconv.ParseBool(ctx)
	}

	cmd.AddCommand(
		newPrometheusCmd(out),
	)

	return cmd
}
