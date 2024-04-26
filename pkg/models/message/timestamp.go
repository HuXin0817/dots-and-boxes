package message

import "time"

const TimeFormatString = "2006-01-02 15:04:05"

type TimeStamp string

func NewTimeStamp(t time.Time) TimeStamp {
	return TimeStamp(t.Format(TimeFormatString))
}

func (ts TimeStamp) Time() time.Time {
	parsedTime, _ := time.Parse(TimeFormatString, string(ts))
	return parsedTime
}
