package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yq314/helm3-monitor/pkg"
	"helm.sh/helm/v3/pkg/action"
	"io"

	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type promCmd struct {
	releaseName   string
	out           io.Writer
	prometheusUrl string
	query         string
}

type prometheusQueryResponse struct {
	Data struct {
		Result []struct{} `json:"result"`
	} `json:"data"`
}

const monitorPrometheusDesc = `
This command monitor a release by querying Prometheus at a given interval and
take care of rolling back to the previous version if the query return a non-
empty result.
Example:
  $ helm monitor prometheus my-release 'rate(http_requests_total{code=~"^5.*$"}[5m]) > 0'
Reference:
  https://prometheus.io/docs/prometheus/latest/querying/basics/
`

func newPrometheusCmd(out io.Writer) *cobra.Command {
	c := &promCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:   "prometheus [flags] RELEASE PROMQL",
		Short: "query a prometheus server",
		Long:  monitorPrometheusDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("this command neeeds 2 arguments: release_name, promql")
			}

			c.releaseName = args[0]
			c.query = args[1]

			return c.run()
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&c.prometheusUrl, "prometheus", "http://localhost:9090", "prometheus url")

	return cmd
}

func (c *promCmd) run() error {
	kubeConfig := pkg.KubeConfig{
		Context: options.kubeContext,
		File:    options.kubeConfigFile,
	}

	actionConfig, err := pkg.GetActionConfig(kubeConfig)
	if err != nil {
		return err
	}

	_, err = actionConfig.Releases.Last(c.releaseName)
	if err != nil {
		return prettyError(err)
	}

	fmt.Fprintf(c.out, "Monitoring %s...\n", c.releaseName)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", c.prometheusUrl+"/api/v1/query", nil)
	if err != nil {
		return prettyError(err)
	}

	q := req.URL.Query()
	q.Add("query", c.query)
	req.URL.RawQuery = q.Encode()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(time.Second * time.Duration(options.interval))

	go func() {
		time.Sleep(time.Second * time.Duration(options.timeout))
		fmt.Fprintf(c.out, "No results after %d second(s)\n", options.timeout)
		close(quit)
	}()

	for {
		select {
		case <-ticker.C:
			debug("Processing URL %s", req.URL.String())

			res, err := client.Do(req)
			if err != nil {
				return prettyError(err)
			}

			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)

			if err != nil {
				return prettyError(err)
			}

			response := &prometheusQueryResponse{}
			err = json.Unmarshal(body, response)
			if err != nil {
				return prettyError(err)
			}

			debug("Response: %v", response)
			debug("Result count: %d", len(response.Data.Result))

			if len(response.Data.Result) > 0 {
				ticker.Stop()

				fmt.Fprintf(c.out, "Failure detected, rolling back...\n")

				rollback := action.NewRollback(actionConfig)
				rollback.DryRun = options.dryRun
				rollback.Recreate = false
				rollback.Force = options.force
				rollback.DisableHooks = options.disableHooks
				rollback.Version = 0
				rollback.Timeout = time.Duration(options.rollbackTimeout) * time.Second
				rollback.Wait = options.wait

				err := rollback.Run(c.releaseName)
				if err != nil {
					return prettyError(err)
				}

				fmt.Fprintf(c.out, "Successfully rolled back to previous revision!\n")
				if options.errorOnRollback {
					return rollbackDone
				}
				return nil
			}

		case <-quit:
			ticker.Stop()
			debug("Quitting...")
			return nil
		}
	}
}
