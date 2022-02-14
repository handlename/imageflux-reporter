package reporter

import (
	"time"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
)

type Month struct {
	t time.Time
}

func ParseMonth(text string) (*Month, error) {
	t, err := time.Parse("2006-01", text)
	if err != nil {
		log.Debug("failed to parse month text", rz.String("text", text))
		return nil, err
	}

	return &Month{t.Local()}, nil
}

func (m Month) StartDate() string {
	return m.t.Format("2006-01-02")
}

func (m Month) EndDate() string {
	return m.t.AddDate(0, 1, 0).Format("2006-01-02")
}

func (m Month) String() string {
	return m.t.Format("2006-01")
}
