package data

import "time"

// SoftDeleteTime 返回用于写入 deleted_at 的当前时间。
func SoftDeleteTime() time.Time {
	return time.Now()
}
