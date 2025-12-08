package aocshared

import (
	"fmt"
	"time"
)

// TimeTask remains the core measurement function, returning the raw duration.
func TimeTask(run func()) time.Duration {
	startTime := time.Now()
	run()
	endTime := time.Now()
	return endTime.Sub(startTime)
}

// -------------------------------------------------------------------------
// Helper Struct and Function for Decomposition (Single Responsibility)
// -------------------------------------------------------------------------

// timeComponents holds the decomposed duration parts.
type timeComponents struct {
	hours     int
	minutes   int
	seconds   int
	subSecond time.Duration // The remaining fraction (ms, µs, ns)
}

// extractTimeComponents decomposes a duration into whole units and the sub-second fraction.
func extractTimeComponents(duration time.Duration) timeComponents {
	// 1. Extract whole hours
	hours := int(duration.Hours())
	remaining := duration - time.Duration(hours)*time.Hour

	// 2. Extract whole minutes
	minutes := int(remaining.Minutes())
	remaining -= time.Duration(minutes) * time.Minute

	// 3. Extract whole seconds
	seconds := int(remaining.Seconds())

	// 4. The remainder is the sub-second component
	subSecond := remaining - time.Duration(seconds)*time.Second

	return timeComponents{
		hours:     hours,
		minutes:   minutes,
		seconds:   seconds,
		subSecond: subSecond,
	}
}

// -------------------------------------------------------------------------
// Primary Formatting Function (Simplified)
// -------------------------------------------------------------------------

// FormatDuration accepts a name and duration, and returns a human-readable string.
func FormatDuration(name string, duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("[%s] completed in %s", name, duration.String())
	}

	components := extractTimeComponents(duration)

	var formattedTime string

	if components.hours > 0 {
		formattedTime += fmt.Sprintf("%dh", components.hours)
	}

	if components.minutes > 0 || components.hours > 0 {
		formattedTime += fmt.Sprintf("%dm", components.minutes)
	}

	if components.seconds > 0 || formattedTime == "" {
		formattedTime += fmt.Sprintf("%ds", components.seconds)
	}

	if components.subSecond > 0 {
		if components.subSecond >= time.Millisecond {
			formattedTime += fmt.Sprintf("%dms", components.subSecond.Milliseconds())
		} else if components.subSecond >= time.Microsecond {
			formattedTime += fmt.Sprintf("%dµs", components.subSecond.Microseconds())
		} else {
			formattedTime += fmt.Sprintf("%dns", components.subSecond.Nanoseconds())
		}
	}

	if formattedTime == "" && duration.Nanoseconds() == 0 {
		formattedTime = "0s"
	}

	return fmt.Sprintf("[%s] completed in %s", name, formattedTime)
}

// DebugAndLogTask runs the function, times it, and prints the result.
func DebugAndLogTask(name string, run func()) {
	duration := TimeTask(run)

	output := FormatDuration(name, duration)

	fmt.Println(output)
}

type Task struct {
	Name string
	Run  func()
}

// DebugAndLogTasks runs the functions, times it, and prints the result.
func DebugAndLogTasks(groupName string, tasks ...Task) {
	var totalDuration time.Duration

	for _, task := range tasks {
		duration := TimeTask(task.Run)
		output := FormatDuration(task.Name, duration)
		fmt.Println(output)
		totalDuration += duration
	}

	totalOutput := FormatDuration(groupName, totalDuration)
	fmt.Println(totalOutput)
}
