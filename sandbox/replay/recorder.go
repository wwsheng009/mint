// replay/recorder.go - 录制器
package replay

import (
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
)

// Recording 录制数据
type Recording struct {
	Metadata  RecordingMetadata
	Events    []platform.RawInput
	Snapshots []*sandbox.Snapshot
}

// RecordingMetadata 录制元数据
type RecordingMetadata struct {
	ID        string
	Timestamp time.Time
	Title     string
	Duration  time.Duration
	EventCount int
}

// Recorder 录制器
type Recorder struct {
	recording *Recording
	startTime time.Time
}

// NewRecorder 创建录制器
func NewRecorder(title string) *Recorder {
	return &Recorder{
		recording: &Recording{
			Metadata: RecordingMetadata{
				ID:        generateRecordingID(),
				Timestamp: time.Now(),
				Title:     title,
			},
			Events: make([]platform.RawInput, 0),
		},
	}
}

// Start 开始录制
func (r *Recorder) Start() {
	r.startTime = time.Now()
}

// Record 记录事件
func (r *Recorder) Record(event platform.RawInput) {
	r.recording.Events = append(r.recording.Events, event)
	r.recording.Metadata.EventCount = len(r.recording.Events)
}

// AddSnapshot 添加快照
func (r *Recorder) AddSnapshot(snap *sandbox.Snapshot) {
	r.recording.Snapshots = append(r.recording.Snapshots, snap)
}

// Stop 停止录制
func (r *Recorder) Stop() *Recording {
	r.recording.Metadata.Duration = time.Since(r.startTime)
	return r.recording
}

// GetRecording 获取录制数据
func (r *Recorder) GetRecording() *Recording {
	return r.recording
}

// generateRecordingID 生成录制ID
func generateRecordingID() string {
	return time.Now().Format("20060102-150405")
}
