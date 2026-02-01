// sandbox/snapshot.go - 快照系统
package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
)

// Snapshot 快照
type Snapshot struct {
	Metadata SnapshotMetadata
	Buffer   *paint.Buffer
	Events   []platform.RawInput
	State    map[string]interface{}
	Checksum string
}

// SnapshotMetadata 快照元数据
type SnapshotMetadata struct {
	ID        string
	Timestamp time.Time
	Level     SnapshotLevel
	Tags      []string
	Size      int64
}

// SnapshotStorage 快照存储接口
type SnapshotStorage interface {
	Save(snap *Snapshot) error
	Load(id string) (*Snapshot, error)
	Delete(id string) error
	List() ([]*SnapshotMetadata, error)
}

// SnapshotManager 快照管理器
type SnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
	order     []string // 按时间顺序
	maxCount  int
	storage   SnapshotStorage
}

// NewSnapshotManager 创建快照管理器
func NewSnapshotManager(maxCount int) *SnapshotManager {
	if maxCount <= 0 {
		maxCount = 100
	}
	return &SnapshotManager{
		snapshots: make(map[string]*Snapshot),
		order:     make([]string, 0),
		maxCount:  maxCount,
	}
}

// SetStorage 设置持久化存储
func (sm *SnapshotManager) SetStorage(storage SnapshotStorage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.storage = storage
}

// Create 创建快照
func (sm *SnapshotManager) Create(level SnapshotLevel, buffer *paint.Buffer, events []platform.RawInput, state map[string]interface{}, tags ...string) (*Snapshot, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snap := &Snapshot{
		Metadata: SnapshotMetadata{
			ID:        generateSnapshotID(),
			Timestamp: time.Now(),
			Level:     level,
			Tags:      tags,
		},
	}

	// 根据级别捕获不同层次的数据
	switch level {
	case SnapshotMinimal:
		snap.Buffer = cloneBuffer(buffer)

	case SnapshotStandard:
		snap.Buffer = cloneBuffer(buffer)
		snap.Events = cloneEvents(events)

	case SnapshotFull:
		snap.Buffer = cloneBuffer(buffer)
		snap.Events = cloneEvents(events)
		snap.State = cloneState(state)
	}

	// 计算校验和
	snap.Checksum = computeChecksum(snap)
	snap.Metadata.Size = estimateSnapshotSize(snap)

	// 存储到内存
	sm.snapshots[snap.Metadata.ID] = snap
	sm.order = append(sm.order, snap.Metadata.ID)

	// 淘汰旧快照
	for len(sm.order) > sm.maxCount {
		oldID := sm.order[0]
		sm.order = sm.order[1:]
		delete(sm.snapshots, oldID)
	}

	// 持久化 (如果配置了)
	if sm.storage != nil {
		if err := sm.storage.Save(snap); err != nil {
			return nil, err
		}
	}

	return snap, nil
}

// Get 获取快照
func (sm *SnapshotManager) Get(id string) (*Snapshot, error) {
	sm.mu.RLock()
	snap, ok := sm.snapshots[id]
	sm.mu.RUnlock()

	if ok {
		return snap, nil
	}

	// 尝试从存储加载
	if sm.storage != nil {
		return sm.storage.Load(id)
	}

	return nil, ErrSnapshotNotFound
}

// List 列出所有快照元数据
func (sm *SnapshotManager) List() []*SnapshotMetadata {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*SnapshotMetadata, 0, len(sm.snapshots))
	for _, id := range sm.order {
		if snap, ok := sm.snapshots[id]; ok {
			meta := snap.Metadata
			result = append(result, &meta)
		}
	}
	return result
}

// Delete 删除快照
func (sm *SnapshotManager) Delete(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.snapshots[id]; !ok {
		return ErrSnapshotNotFound
	}

	delete(sm.snapshots, id)

	// 从顺序列表中移除
	for i, oid := range sm.order {
		if oid == id {
			sm.order = append(sm.order[:i], sm.order[i+1:]...)
			break
		}
	}

	// 从存储删除
	if sm.storage != nil {
		return sm.storage.Delete(id)
	}

	return nil
}

// Verify 验证快照完整性
func (sm *SnapshotManager) Verify(snap *Snapshot) bool {
	return snap.Checksum == computeChecksum(snap)
}

// ==============================================================================
// Helper Functions
// ==============================================================================

func generateSnapshotID() string {
	data := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(string(rune(data))))
	return hex.EncodeToString(hash[:8])
}

func computeChecksum(snap *Snapshot) string {
	data, _ := json.Marshal(struct {
		Level  SnapshotLevel
		Events int
		State  int
	}{
		Level:  snap.Metadata.Level,
		Events: len(snap.Events),
		State:  len(snap.State),
	})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func estimateSnapshotSize(snap *Snapshot) int64 {
	var size int64 = 100 // 基础元数据

	if snap.Buffer != nil {
		size += int64(snap.Buffer.Width * snap.Buffer.Height * 16) // 每个 Cell 约 16 字节
	}

	size += int64(len(snap.Events) * 64) // 每个事件约 64 字节

	return size
}

func cloneBuffer(buf *paint.Buffer) *paint.Buffer {
	if buf == nil {
		return nil
	}

	clone := paint.NewBuffer(buf.Width, buf.Height)
	for y := 0; y < buf.Height; y++ {
		copy(clone.Cells[y], buf.Cells[y])
	}
	return clone
}

func cloneEvents(events []platform.RawInput) []platform.RawInput {
	if events == nil {
		return nil
	}
	clone := make([]platform.RawInput, len(events))
	copy(clone, events)
	return clone
}

func cloneState(state map[string]interface{}) map[string]interface{} {
	if state == nil {
		return nil
	}
	clone := make(map[string]interface{}, len(state))
	for k, v := range state {
		clone[k] = v
	}
	return clone
}
