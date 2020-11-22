/*
 * @Author: chenjingwei
 * @Date: 2020-11-08 11:27:56
 * @Last Modified by: chenjingwei
 * @Last Modified time: 2020-11-12 23:15:52
 */

package util

import (
	"testing"
	"time"
)

func Test_TimeWheel_Base(t *testing.T) {
	i := int64(0)
	cb := func(userData interface{}) bool {
		println("user:", userData.(int64))
		return true
	}
	now := time.Now().Unix()
	tw := NewTimeWheel(nil, func() int64 { return now + i }, now, cb)

	guidBase := int64(888)
	for i = 0; i <= 5; i++ {
		tw.UpdateTask(now+int64(i), int64(guidBase+i), 0, i)
	}

	for i = 0; i <= 5; i++ {
		tw.Tick()
	}
}

func Test_TimeWheel_Trigger(t *testing.T) {
	i := int64(0)
	guidBase := int64(888)
	now := time.Now().Unix()

	cb := func(userData interface{}) bool {
		tw := userData.(*TimeWheel)
		println("tick time:", now, tw.Now())
		tw.UpdateTask(tw.Now()+1, guidBase, 0, tw)
		return true
	}

	tw := NewTimeWheel(nil, func() int64 { return now + i }, now, cb)

	tw.UpdateTask(now, guidBase+i, 0, tw)
	for i = 0; i <= 5; i++ {
		tw.Tick()
	}
}

func Test_TimeWheel_Long(t *testing.T) {
	i := int64(0)
	guidBase := int64(888)
	now := time.Now().Unix()
	cb := func(userData interface{}) bool {
		tw := userData.(*TimeWheel)
		println("tick time:", tw.Now()-now)
		return true
	}

	tw := NewTimeWheel(nil, func() int64 { return now + i }, now, cb)
	tw.UpdateTask(now+3600, guidBase, 0, tw)
	i = 3600
	tw.Tick()

	tw.UpdateTask(now+7215, guidBase, 0, tw)
	i = 7215
	tw.Tick()

	tw.UpdateTask(now+86400, guidBase, 0, tw)
	i = 86400
	tw.Tick()

	tw.UpdateTask(now+86400*360, guidBase, 0, tw)
	i = 86400 * 360
	tw.Tick()
}
