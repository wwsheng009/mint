// testing/reporter.go - 测试报告器
package testing

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Reporter 报告器接口
type Reporter interface {
	SuiteStart(suite *TestSuite)
	SuiteEnd(report *TestReport)
	CaseStart(tc *TestCase)
	CaseEnd(report *CaseReport)
}

// ConsoleReporter 控制台报告器
type ConsoleReporter struct {
	Verbose bool
}

// NewConsoleReporter 创建控制台报告器
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		Verbose: false,
	}
}

// SuiteStart 测试套件开始
func (r *ConsoleReporter) SuiteStart(suite *TestSuite) {
	fmt.Printf("=== RUN %s\n", suite.Name)
}

// SuiteEnd 测试套件结束
func (r *ConsoleReporter) SuiteEnd(report *TestReport) {
	fmt.Printf("\n--- %s: ", report.SuiteName)

	if report.Error != nil {
		fmt.Printf("ERROR: %v\n", report.Error)
		return
	}

	if report.Passed() {
		fmt.Printf("PASS (%v)\n", report.Duration.Round(time.Millisecond))
	} else {
		fmt.Printf("FAIL (%d passed, %d failed, %v)\n",
			report.PassedCount, report.FailedCount, report.Duration.Round(time.Millisecond))
	}

	// 打印失败的用例
	for _, cr := range report.CaseReports {
		if !cr.Passed {
			fmt.Printf("    --- FAIL: %s (%v)\n", cr.Name, cr.Duration.Round(time.Millisecond))
			if r.Verbose && cr.Error != nil {
				fmt.Printf("        Error: %v\n", cr.Error)
			}
		}
	}
}

// CaseStart 测试用例开始
func (r *ConsoleReporter) CaseStart(tc *TestCase) {
	if r.Verbose {
		fmt.Printf("    --- RUN: %s\n", tc.Name)
	}
}

// CaseEnd 测试用例结束
func (r *ConsoleReporter) CaseEnd(report *CaseReport) {
	if r.Verbose {
		status := "PASS"
		if !report.Passed {
			status = "FAIL"
		}
		fmt.Printf("    --- %s: %s (%v)\n", status, report.Name, report.Duration.Round(time.Millisecond))
	} else {
		if report.Passed {
			fmt.Print(".")
		} else {
			fmt.Print("F")
		}
	}
}

// JSONReporter JSON报告器
type JSONReporter struct {
	Output string // 输出文件路径
}

// NewJSONReporter 创建JSON报告器
func NewJSONReporter(output string) *JSONReporter {
	return &JSONReporter{
		Output: output,
	}
}

// SuiteStart 测试套件开始
func (r *JSONReporter) SuiteStart(suite *TestSuite) {
	// JSON报告器不实时输出
}

// SuiteEnd 测试套件结束
func (r *JSONReporter) SuiteEnd(report *TestReport) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return
	}

	if r.Output == "" {
		fmt.Println(string(data))
	} else {
		_ = os.WriteFile(r.Output, data, 0644)
	}
}

// CaseStart 测试用例开始
func (r *JSONReporter) CaseStart(tc *TestCase) {
	// JSON报告器不实时输出
}

// CaseEnd 测试用例结束
func (r *JSONReporter) CaseEnd(report *CaseReport) {
	// JSON报告器不实时输出
}

// JUnitReporter JUnit报告器
type JUnitReporter struct {
	Output string // 输出文件路径
}

// NewJUnitReporter 创建JUnit报告器
func NewJUnitReporter(output string) *JUnitReporter {
	return &JUnitReporter{
		Output: output,
	}
}

// SuiteStart 测试套件开始
func (r *JUnitReporter) SuiteStart(suite *TestSuite) {
	// JUnit报告器不实时输出
}

// SuiteEnd 测试套件结束
func (r *JUnitReporter) SuiteEnd(report *TestReport) {
	// 生成JUnit XML格式
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="%s" tests="%d" failures="%d" time="%f" timestamp="%s">
`,
		report.SuiteName,
		len(report.CaseReports),
		report.FailedCount,
		report.Duration.Seconds(),
		report.StartTime.Format("2006-01-02T15:04:05"),
	)

	for _, cr := range report.CaseReports {
		status := "passed"
		if !cr.Passed {
			status = fmt.Sprintf(`failure message="%s"`, cr.Error)
		}
		xml += fmt.Sprintf(`  <testcase name="%s" time="%f" %s/>
`, cr.Name, cr.Duration.Seconds(), status)
	}

	xml += "</testsuite>\n"

	if r.Output == "" {
		fmt.Print(xml)
	} else {
		_ = os.WriteFile(r.Output, []byte(xml), 0644)
	}
}

// CaseStart 测试用例开始
func (r *JUnitReporter) CaseStart(tc *TestCase) {
	// JUnit报告器不实时输出
}

// CaseEnd 测试用例结束
func (r *JUnitReporter) CaseEnd(report *CaseReport) {
	// JUnit报告器不实时输出
}
