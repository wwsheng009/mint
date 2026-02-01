// testing/runner.go - 测试运行器
package testing

import (
	"fmt"
	"time"
)

// TestCase 测试用例
type TestCase struct {
	Name    string
	Timeout time.Duration
	Run     func() error
}

// TestSuite 测试套件
type TestSuite struct {
	Name    string
	Cases   []*TestCase
	Setup   func() error
	Teardown func() error
}

// Runner 测试运行器
type Runner struct {
	suite    *TestSuite
	reporter Reporter
}

// NewRunner 创建测试运行器
func NewRunner(suite *TestSuite) *Runner {
	return &Runner{
		suite:    suite,
		reporter: NewConsoleReporter(),
	}
}

// SetReporter 设置报告器
func (r *Runner) SetReporter(reporter Reporter) {
	r.reporter = reporter
}

// Run 运行测试套件
func (r *Runner) Run() *TestReport {
	report := &TestReport{
		SuiteName:  r.suite.Name,
		StartTime:  time.Now(),
		CaseReports: make([]*CaseReport, 0),
	}

	r.reporter.SuiteStart(r.suite)

	// Setup
	if r.suite.Setup != nil {
		if err := r.suite.Setup(); err != nil {
			report.Error = fmt.Errorf("setup failed: %w", err)
			r.reporter.SuiteEnd(report)
			return report
		}
	}

	// 运行所有测试用例
	for _, tc := range r.suite.Cases {
		caseReport := r.runCase(tc)
		report.CaseReports = append(report.CaseReports, caseReport)

		if !caseReport.Passed {
			report.FailedCount++
		} else {
			report.PassedCount++
		}
	}

	// Teardown
	if r.suite.Teardown != nil {
		if err := r.suite.Teardown(); err != nil {
			report.Error = fmt.Errorf("teardown failed: %w", err)
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	r.reporter.SuiteEnd(report)

	return report
}

// runCase 运行单个测试用例
func (r *Runner) runCase(tc *TestCase) *CaseReport {
	caseReport := &CaseReport{
		Name:     tc.Name,
		StartTime: time.Now(),
	}

	r.reporter.CaseStart(tc)

	// 运行测试
	var err error
	if tc.Timeout > 0 {
		done := make(chan error, 1)
		go func() {
			done <- tc.Run()
		}()

		select {
		case e := <-done:
			err = e
		case <-time.After(tc.Timeout):
			err = fmt.Errorf("timeout after %v", tc.Timeout)
		}
	} else {
		err = tc.Run()
	}

	caseReport.EndTime = time.Now()
	caseReport.Duration = caseReport.EndTime.Sub(caseReport.StartTime)

	if err != nil {
		caseReport.Passed = false
		caseReport.Error = err
	} else {
		caseReport.Passed = true
	}

	r.reporter.CaseEnd(caseReport)

	return caseReport
}

// TestReport 测试报告
type TestReport struct {
	SuiteName   string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	CaseReports []*CaseReport
	PassedCount int
	FailedCount int
	Error       error
}

// Passed 所有测试是否通过
func (r *TestReport) Passed() bool {
	return r.FailedCount == 0 && r.Error == nil
}

// CaseReport 用例报告
type CaseReport struct {
	Name     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Passed    bool
	Error     error
}
