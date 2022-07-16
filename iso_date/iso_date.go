package isoDate

import (
	"fmt"
	"time"
)

type IsoDate struct {
	Time time.Time
}

func NewIsoDateFromString(isoString string) (*IsoDate, error) {
	parsedTime, err := time.Parse(time.RFC3339, isoString+"T00:00:00+11:00")

	if err != nil {
		return nil, err
	}

	return &IsoDate{parsedTime}, nil
}

func NewIsoDate() *IsoDate {
	return &IsoDate{time.Now()}
}

func (isoDate *IsoDate) Stringify() string {
	return fmt.Sprintf("%d-%02d-%02d", isoDate.Time.Year(), int(isoDate.Time.Month()), isoDate.Time.Day())
}

func (isoDate *IsoDate) Timify() time.Time {
	return isoDate.Time
}

func (isoDate *IsoDate) NextDay() *IsoDate {
	return &IsoDate{isoDate.Time.AddDate(0, 0, 1)}
}

func (isoDate *IsoDate) PreviousDay() *IsoDate {
	return &IsoDate{isoDate.Time.AddDate(0, 0, -1)}
}
