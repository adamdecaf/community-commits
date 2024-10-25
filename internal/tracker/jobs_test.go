package tracker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJobs_FormatCrontabSchedule(t *testing.T) {
	got := formatCrontabSchedule(30 * time.Minute)
	require.Equal(t, "0 1/30 * * * *", got)

	got = formatCrontabSchedule(127 * time.Minute)
	require.Equal(t, "0 7 1/2 * * *", got)

	got = formatCrontabSchedule(time.Hour)
	require.Equal(t, "0 0 1/1 * * *", got)

	got = formatCrontabSchedule(12 * time.Hour)
	require.Equal(t, "0 0 1/12 * * *", got)

	// Over 24h isn't really supported by crontab
	got = formatCrontabSchedule(30 * time.Hour)
	require.Equal(t, "0 0 6 1/2 * *", got) // this is really every 48hrs

	// complex cases
	got = formatCrontabSchedule(3*time.Hour + 30*time.Minute)
	require.Equal(t, "0 30 1/3 * * *", got)
}
