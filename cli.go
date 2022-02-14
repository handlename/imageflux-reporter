package reporter

import (
	"context"
	"flag"
	"fmt"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
	"github.com/google/subcommands"
)

const binName = "imageflux-reporter"

type RateCmd struct {
	month string
}

func (*RateCmd) Name() string {
	return "rate"
}

func (*RateCmd) Synopsis() string {
	return "print rate report"
}

func (*RateCmd) Usage() string {
	return fmt.Sprintf(`%s rate -month <YYYY-MM>
`, binName)
}

func (c *RateCmd) SetFlags(f *flag.FlagSet) {
	// TODO: validation
	f.StringVar(&c.month, "month", "", "target month")
}

func (c *RateCmd) Execute(ctx context.Context, f *flag.FlagSet, opts ...interface{}) subcommands.ExitStatus {
	config := opts[0].(*Config)

	month, err := ParseMonth(c.month)
	if err != nil {
		log.Error("failed to parse -month", rz.String("month", c.month))
		return subcommands.ExitFailure
	}
	log.Debug("succeeded to ParseMonth", rz.String("month", month.String()))

	client, err := NewImageFluxClient(config.Email, config.Password)
	if err != nil {
		log.Error("failed to new ImageFluxClient", rz.Err(err))
		return subcommands.ExitFailure
	}

	r, err := NewRateReporter(ctx, client, config.Origins)
	if err != nil {
		log.Error("failed to init RateReporter", rz.Err(err))
		return subcommands.ExitFailure
	}

	reports, err := r.Run(ctx, *month)
	if err != nil {
		log.Error("failed to run reporter", rz.Err(err))
		return subcommands.ExitFailure
	}

	for _, report := range reports.Reports {
		fmt.Printf("%s\t%d\t%f\n", report.Project, report.Volume, report.Rate)
	}

	return subcommands.ExitSuccess
}
