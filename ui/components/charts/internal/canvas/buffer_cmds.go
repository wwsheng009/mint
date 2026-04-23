package canvas

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
)

// BufferToDrawCmds converts a paint buffer into ordered draw commands.
// It preserves full row coverage by emitting style runs including spaces.
func BufferToDrawCmds(buf *paint.Buffer, offsetX, offsetY int) []paint.DrawCmd {
	if buf == nil || buf.Width <= 0 || buf.Height <= 0 {
		return nil
	}

	cmds := make([]paint.DrawCmd, 0, buf.Height*2)
	for y := 0; y < buf.Height; y++ {
		cmds = append(cmds, bufferRowToDrawCmds(buf, y, offsetX, offsetY+y)...)
	}
	return cmds
}

func bufferRowToDrawCmds(buf *paint.Buffer, row, offsetX, offsetY int) []paint.DrawCmd {
	cmds := make([]paint.DrawCmd, 0, 4)
	x := 0
	for x < buf.Width {
		cell := buf.Cells[row][x]
		if cell.IsContinuation {
			x++
			continue
		}

		startX := x
		runStyle := cell.Style
		var builder strings.Builder

		for x < buf.Width {
			cell = buf.Cells[row][x]
			if cell.IsContinuation {
				x++
				continue
			}
			if x > startX && cell.Style != runStyle {
				break
			}

			cluster := cell.Cluster
			if cluster == "" {
				cluster = " "
			}
			builder.WriteString(cluster)

			if cell.Width > 1 {
				x += cell.Width
				continue
			}
			x++
		}

		cmds = append(cmds, paint.DrawCmd{
			X:     offsetX + startX,
			Y:     offsetY,
			Text:  builder.String(),
			Style: runStyle,
		})
	}
	return cmds
}
