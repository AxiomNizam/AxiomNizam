package timeutil

import (
	"fmt"
	"time"
)

// TimeFormat defines time formats
type TimeFormat string

const (
	RFC3339     TimeFormat = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano TimeFormat = "2006-01-02T15:04:05.999999999Z07:00"
	ISO8601     TimeFormat = "2006-01-02T15:04:05Z07:00"
	UnixDate    TimeFormat = "Mon Jan _2 15:04:05 MST 2006"
	RFC822      TimeFormat = "02 Jan 06 15:04 MST"
	Kitchen     TimeFormat = "3:04PM"
	Stamp       TimeFormat = "Jan _2 15:04:05"
)

// Now returns current time in UTC
func Now() time.Time {
	return time.Now().UTC()
}

// NowUnix returns current Unix timestamp
func NowUnix() int64 {
	return Now().Unix()
}

// NowUnixNano returns current Unix timestamp in nanoseconds
func NowUnixNano() int64 {
	return Now().UnixNano()
}

// NowUnixMilli returns current Unix timestamp in milliseconds
func NowUnixMilli() int64 {
	return Now().UnixMilli()
}

// Format formats time with specified format
func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatRFC3339 formats time as RFC3339
func FormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Parse parses time string
func Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseRFC3339 parses RFC3339 time string
func ParseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// Duration represents duration utilities
type Duration struct {
	d time.Duration
}

// NewDuration creates new duration
func NewDuration(d time.Duration) *Duration {
	return &Duration{d: d}
}

// Seconds returns duration in seconds
func (d *Duration) Seconds() int64 {
	return int64(d.d.Seconds())
}

// Milliseconds returns duration in milliseconds
func (d *Duration) Milliseconds() int64 {
	return d.d.Milliseconds()
}

// Microseconds returns duration in microseconds
func (d *Duration) Microseconds() int64 {
	return d.d.Microseconds()
}

// Nanoseconds returns duration in nanoseconds
func (d *Duration) Nanoseconds() int64 {
	return d.d.Nanoseconds()
}

// String returns duration as string
func (d *Duration) String() string {
	return d.d.String()
}

// Add adds duration to time
func Add(t time.Time, d time.Duration) time.Time {
	return t.Add(d)
}

// Sub subtracts time from another time
func Sub(t1, t2 time.Time) time.Duration {
	return t1.Sub(t2)
}

// AddSeconds adds seconds to time
func AddSeconds(t time.Time, seconds int64) time.Time {
	return Add(t, time.Duration(seconds)*time.Second)
}

// AddMinutes adds minutes to time
func AddMinutes(t time.Time, minutes int64) time.Time {
	return Add(t, time.Duration(minutes)*time.Minute)
}

// AddHours adds hours to time
func AddHours(t time.Time, hours int64) time.Time {
	return Add(t, time.Duration(hours)*time.Hour)
}

// AddDays adds days to time
func AddDays(t time.Time, days int64) time.Time {
	return Add(t, time.Duration(days)*24*time.Hour)
}

// AddMonths adds months to time
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears adds years to time
func AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// StartOfDay returns start of day
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns end of day
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfMonth returns start of month
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns end of month
func EndOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, t.Location())
}

// StartOfYear returns start of year
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns end of year
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.December, 31, 23, 59, 59, 999999999, t.Location())
}

// IsToday checks if time is today
func IsToday(t time.Time) bool {
	today := Now()
	return t.Year() == today.Year() &&
		t.Month() == today.Month() &&
		t.Day() == today.Day()
}

// IsYesterday checks if time is yesterday
func IsYesterday(t time.Time) bool {
	yesterday := Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() &&
		t.Month() == yesterday.Month() &&
		t.Day() == yesterday.Day()
}

// IsPast checks if time is in past
func IsPast(t time.Time) bool {
	return t.Before(Now())
}

// IsFuture checks if time is in future
func IsFuture(t time.Time) bool {
	return t.After(Now())
}

// IsExpired checks if time has expired relative to now
func IsExpired(t time.Time) bool {
	return IsPast(t)
}

// ExpiresIn returns time until expiration
func ExpiresIn(t time.Time) time.Duration {
	return time.Until(t)
}

// Age returns age of time since now
func Age(t time.Time) time.Duration {
	return Sub(Now(), t)
}

// DaysBetween returns days between two times
func DaysBetween(t1, t2 time.Time) int64 {
	return int64(Sub(t1, t2).Hours() / 24)
}

// HoursBetween returns hours between two times
func HoursBetween(t1, t2 time.Time) int64 {
	return int64(Sub(t1, t2).Hours())
}

// MinutesBetween returns minutes between two times
func MinutesBetween(t1, t2 time.Time) int64 {
	return int64(Sub(t1, t2).Minutes())
}

// SecondsBetween returns seconds between two times
func SecondsBetween(t1, t2 time.Time) int64 {
	return int64(Sub(t1, t2).Seconds())
}

// Timer represents a timer
type Timer struct {
	start time.Time
	end   time.Time
}

// NewTimer creates a new timer
func NewTimer() *Timer {
	return &Timer{
		start: Now(),
	}
}

// Stop stops the timer
func (t *Timer) Stop() time.Duration {
	t.end = Now()
	return t.Elapsed()
}

// Elapsed returns elapsed time
func (t *Timer) Elapsed() time.Duration {
	if t.end.IsZero() {
		return Sub(Now(), t.start)
	}
	return Sub(t.end, t.start)
}

// ElapsedSeconds returns elapsed seconds
func (t *Timer) ElapsedSeconds() float64 {
	return t.Elapsed().Seconds()
}

// ElapsedMillis returns elapsed milliseconds
func (t *Timer) ElapsedMillis() int64 {
	return t.Elapsed().Milliseconds()
}

// String returns elapsed time as string
func (t *Timer) String() string {
	return fmt.Sprintf("%.3fs", t.ElapsedSeconds())
}

// Stopwatch is an alias for Timer
type Stopwatch = Timer

// NewStopwatch creates a new stopwatch
func NewStopwatch() *Stopwatch {
	return NewTimer()
}

// Schedule represents a scheduled task
type Schedule struct {
	duration time.Duration
	next     time.Time
}

// NewSchedule creates a new schedule
func NewSchedule(interval time.Duration) *Schedule {
	return &Schedule{
		duration: interval,
		next:     Now().Add(interval),
	}
}

// IsReady checks if schedule is ready
func (s *Schedule) IsReady() bool {
	return Now().After(s.next)
}

// Mark marks the schedule as executed
func (s *Schedule) Mark() {
	s.next = Now().Add(s.duration)
}

// MarkForced marks schedule for forced execution
func (s *Schedule) MarkForced() {
	s.next = Now()
}

// NextExecutionTime returns next execution time
func (s *Schedule) NextExecutionTime() time.Time {
	return s.next
}

// TimeUntilNext returns time until next execution
func (s *Schedule) TimeUntilNext() time.Duration {
	return time.Until(s.next)
}

// Timezone represents timezone operations
type Timezone struct {
	location *time.Location
}

// NewTimezone creates a new timezone
func NewTimezone(name string) (*Timezone, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}
	return &Timezone{
		location: loc,
	}, nil
}

// Convert converts time to this timezone
func (tz *Timezone) Convert(t time.Time) time.Time {
	return t.In(tz.location)
}

// Now returns current time in this timezone
func (tz *Timezone) Now() time.Time {
	return Now().In(tz.location)
}

// Batch represents batch time operations
type Batch struct {
	times []time.Time
}

// NewBatch creates a new batch
func NewBatch() *Batch {
	return &Batch{
		times: make([]time.Time, 0),
	}
}

// Add adds time to batch
func (b *Batch) Add(t time.Time) *Batch {
	b.times = append(b.times, t)
	return b
}

// Latest returns latest time
func (b *Batch) Latest() time.Time {
	if len(b.times) == 0 {
		return time.Time{}
	}

	latest := b.times[0]
	for _, t := range b.times[1:] {
		if t.After(latest) {
			latest = t
		}
	}
	return latest
}

// Earliest returns earliest time
func (b *Batch) Earliest() time.Time {
	if len(b.times) == 0 {
		return time.Time{}
	}

	earliest := b.times[0]
	for _, t := range b.times[1:] {
		if t.Before(earliest) {
			earliest = t
		}
	}
	return earliest
}

// Count returns number of times
func (b *Batch) Count() int {
	return len(b.times)
}
