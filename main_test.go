package main

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type logTest struct {
	name          string
	lines         []string
	expectedLines []string
}

func init() {
	debug = log.New(os.Stdout, "[pretty-test] ", 0)
}

func TestStandardNonJSON(t *testing.T) {
	runLogTests(t, []logTest{
		{
			name: "single_non_log_line",
			lines: []string{
				"A non-JSON string line",
			},
			expectedLines: []string{
				"A non-JSON string line",
			},
		},
		{
			name: "single_log_line_invalid_json",
			lines: []string{
				`{"severity":"s","time":"t","caller":"c:0"`,
			},
			expectedLines: []string{
				`{"severity":"s","time":"t","caller":"c:0"`,
			},
		},
	})
}

func TestStandardNewProduction(t *testing.T) {
	runLogTests(t, []logTest{
		{
			name: "single_log_line",
			lines: []string{
				`{"level":"info","ts":1545445711.144533,"caller":"c","msg":"m"}`,
			},
			expectedLines: []string{
				"[2018-12-21 21:28:31.144 EST] \x1b[32minfo\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
			},
		},
	})
}

func TestStandardFieldTs_ISO8601_string(t *testing.T) {
	runLogTests(t, []logTest{
		{
			name: "default",
			lines: []string{
				`{"level":"info","ts":"2019-12-06T19:40:20.627Z","caller":"c","msg":"m"}`,
			},
			expectedLines: []string{
				"[2019-12-06 14:40:20.627 EST] \x1b[32minfo\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
			},
		},
	})
}

func TestZapDriverNewProduction(t *testing.T) {
	runLogTests(t, []logTest{
		{
			name: "single_log_line",
			lines: []string{
				zapdriverLine("INFO", "2018-12-21T23:06:49.435919-05:00"),
			},
			expectedLines: []string{
				"[2018-12-21 23:06:49.435 EST] \x1b[32mINFO\x1b[0m \x1b[38;5;244m(c:0)\x1b[0m \x1b[34mm\x1b[0m {\"folder\":\"f\"}",
			},
		},
		{
			name: "line_with_timestamp_vs_time",
			lines: []string{
				`{"severity":"INFO","timestamp":"2018-12-21T23:06:49.435919-05:00","caller":"c:0","message":"m","folder":"f","labels":{},"logging.googleapis.com/sourceLocation":{"file":"f","line":"1","function":"fn"}}`,
			},
			expectedLines: []string{
				"[2018-12-21 23:06:49.435 EST] \x1b[32mINFO\x1b[0m \x1b[38;5;244m(c:0)\x1b[0m \x1b[34mm\x1b[0m {\"folder\":\"f\"}",
			},
		},
		{
			name: "single_log_line_missing_fields",
			lines: []string{
				`{"severity":"s","time":"t","caller":"c:0"}`,
			},
			expectedLines: []string{
				`{"severity":"s","time":"t","caller":"c:0"}`,
			},
		},
		{
			name: "multi_log_line",
			lines: []string{
				zapdriverLine("INFO", "2018-12-21T23:06:49.435919-05:00"),
				zapdriverLine("DEBUG", "2018-12-21T23:06:49.436920-05:00"),
			},
			expectedLines: []string{
				"[2018-12-21 23:06:49.435 EST] \x1b[32mINFO\x1b[0m \x1b[38;5;244m(c:0)\x1b[0m \x1b[34mm\x1b[0m {\"folder\":\"f\"}",
				"[2018-12-21 23:06:49.436 EST] \x1b[34mDEBUG\x1b[0m \x1b[38;5;244m(c:0)\x1b[0m \x1b[34mm\x1b[0m {\"folder\":\"f\"}",
			},
		},
		{
			name: "multi_mixed",
			lines: []string{
				zapdriverLine("INFO", "2018-12-21T23:06:49.435919-05:00"),
				"A non-JSON string line",
				zapdriverLine("DEBUG", "2018-12-21T23:06:49.436920-05:00"),
			},
			expectedLines: []string{
				"[2018-12-21 23:06:49.435 EST] \x1b[32mINFO\x1b[0m \x1b[38;5;244m(c:0)\x1b[0m \x1b[34mm\x1b[0m {\"folder\":\"f\"}",
				"A non-JSON string line",
				"[2018-12-21 23:06:49.436 EST] \x1b[34mDEBUG\x1b[0m \x1b[38;5;244m(c:0)\x1b[0m \x1b[34mm\x1b[0m {\"folder\":\"f\"}",
			},
		},
		{
			name: "error_verbose_alone_right_format",
			lines: []string{
				`{"severity":"ERROR","time":"2019-04-15T15:49:55.676461-04:00","caller":"c","message":"m","errorVerbose":"initial message\nSectionA\nStack1a\n\tFile1a\nStack2a\n\tFile2a\nSectionB\nStack1b\n\tFile1b\nStack2b\n\tFile2b"}`,
			},
			expectedLines: []string{
				"[2019-04-15 15:49:55.676 EDT] \x1b[31mERROR\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
				`Error Verbose`,
				`  initial message`,
				``,
				`  SectionA`,
				`    Stack1a`,
				"    \tFile1a",
				`    Stack2a`,
				"    \tFile2a",
				``,
				`  SectionB`,
				`    Stack1b`,
				"    \tFile1b",
				`    Stack2b`,
				"    \tFile2b",
			},
		},
		{
			name: "error_verbose_alone_wrong_format_single_line",
			lines: []string{
				`{"severity":"ERROR","time":"2019-04-15T15:49:55.676461-04:00","caller":"c","message":"m","errorVerbose":"single line"}`,
			},
			expectedLines: []string{
				"[2019-04-15 15:49:55.676 EDT] \x1b[31mERROR\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
				`Error Verbose`,
				`  single line`,
			},
		},
		{
			name: "error_verbose_alone_wrong_format_multi_line",
			lines: []string{
				`{"severity":"ERROR","time":"2019-04-15T15:49:55.676461-04:00","caller":"c","message":"m","errorVerbose":"multi\nline"}`,
			},
			expectedLines: []string{
				"[2019-04-15 15:49:55.676 EDT] \x1b[31mERROR\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
				`Error Verbose`,
				`  multi`,
				`  line`,
			},
		},
		{
			name: "error_verbose_alone_wrong_format_for_stack_line",
			lines: []string{
				`{"severity":"ERROR","time":"2019-04-15T15:49:55.676461-04:00","caller":"c","message":"m","errorVerbose":"Stack1b\n\tFile1bStack2b\n\tFile2b"}`,
			},
			expectedLines: []string{
				"[2019-04-15 15:49:55.676 EDT] \x1b[31mERROR\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
				`Error Verbose`,
				`      Stack1b`,
				"    \tFile1bStack2b",
				"    \tFile2b",
			},
		},
		{
			name: "stacktrace_alone_right_format",
			lines: []string{
				`{"severity":"ERROR","time":"2019-04-15T15:49:55.676461-04:00","caller":"c","message":"m","stacktrace":"Stack1a\n\tFile1a\nStack2a\n\tFile2a"}`,
			},
			expectedLines: []string{
				"[2019-04-15 15:49:55.676 EDT] \x1b[31mERROR\x1b[0m \x1b[38;5;244m(c)\x1b[0m \x1b[34mm\x1b[0m",
				`Stacktrace`,
				`    Stack1a`,
				"    \tFile1a",
				`    Stack2a`,
				"    \tFile2a",
			},
		},
	})
}

func runLogTests(t *testing.T, tests []logTest) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := executeProcessorTest(test.lines)

			outputLines := strings.Split(writer.String(), "\n")
			require.Equal(t, test.expectedLines, outputLines)
		})
	}
}
