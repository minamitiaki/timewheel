/*
 * @Author: chenjingwei
 * @Date: 2020-11-05 14:59:28
 * @Last Modified by: chenjingwei
 * @Last Modified time: 2020-11-10 16:48:59
 * @Desc 秒级精度的定时器
 */

package util

import (
	"container/list"
	"fmt"
	"time"
)

// Task 延时任务
type TWTask struct {
	time     int64 // 目标时间
	guid     int64 // 定时器唯一标识
	cycle    int64
	userData interface{} // 回调函数参数
}

type TWCallBack func(userData interface{}) bool

type TWOneWheel struct {
	scale      int64 //精度(秒)
	slots      []*list.List
	currentPos int64
}

type TimeWheel struct {
	tickTime int64
	userNow  func() int64      //用户时间可能不是标准时间
	wheel    []*TWOneWheel     //槽位scond,hour,day,year
	cb       TWCallBack        //定时器回调函数
	Index    map[int64]([]int) //guid <--> wheel, pos
}

func (tw *TimeWheel) tick(w *TWOneWheel, index int) {
	//递归终止条件
	if index >= len(tw.wheel) {
		return
	}

	slotMove := func() {
		w.currentPos++
		if w.currentPos >= int64(len(w.slots)) {
			w.currentPos = 0
			tw.tick(tw.wheel[index+1], index+1) //递归触发tick
		}
	}

	//处理当前槽位
	if index == 0 {
		l := w.slots[w.currentPos]
		if l != nil { //执行
			for e := l.Front(); e != nil; e = e.Next() {
				task := e.Value.(*TWTask)
				end := tw.cb(task.userData)
				if task.cycle == 0 || end {
					delete(tw.Index, task.guid) //删除索引
				} else {
					task.time += task.cycle
					tw.addTask(task)
				}
			}
			w.slots[w.currentPos] = nil
		}
		//偏移时刻
		if index == 0 {
			tw.tickTime++
		}
		//偏移槽位
		slotMove()
	} else {
		//偏移槽位
		slotMove()
		l := w.slots[w.currentPos]
		if l != nil { //重挂
			for e := l.Front(); e != nil; e = e.Next() {
				tw.addTask(e.Value.(*TWTask))
			}
			w.slots[w.currentPos] = nil
		}
	}
}

func (tw *TimeWheel) Now() int64 {
	if tw.userNow == nil {
		return time.Now().Unix()
	}
	return tw.userNow()
}

func (tw *TimeWheel) addTask(task *TWTask) error {
	delay := task.time - tw.tickTime
	if delay < 0 {
		delay = 0
	}
	for i, v := range tw.wheel {
		if delay < v.scale*int64(len(v.slots)) {
			pos := int(v.currentPos+delay/v.scale) % len(v.slots)
			if v.slots[pos] == nil {
				v.slots[pos] = &list.List{}
			}
			v.slots[pos].PushBack(task)
			tw.Index[task.guid] = nil
			tw.Index[task.guid] = append(tw.Index[task.guid], i, pos)
			return nil
		}
	}
	return fmt.Errorf("can't find wheel to add delay:%d %v", delay, task)
}

func (tw *TimeWheel) UpdateTask(time int64, guid int64, cycle int64, userData interface{}) error {
	tw.RemoveTask(guid)

	task := &TWTask{}
	task.time = time
	task.guid = guid
	task.cycle = cycle
	task.userData = userData

	return tw.addTask(task)
}

func (tw *TimeWheel) RemoveTask(guid int64) {
	index := tw.Index[guid]
	if index == nil {
		return
	}
	l := tw.wheel[index[0]].slots[index[1]]
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(*TWTask).guid == guid {
			delete(tw.Index, guid)
			l.Remove(e)
			return
		}
	}
}

//拨动时间轮
func (tw *TimeWheel) Tick() {
	now := tw.Now()
	for tw.tickTime <= now {
		tw.tick(tw.wheel[0], 0)
	}
}

/*
	plan: 时间轮倍率设置，基本精度1秒
	eg: plan = append(plan, 3600, 24, 360, 100) //100 year
	设置四个轮子以及时间跨度表达倍率的增长
	userNow: 用户自定义获得当前秒数的函数，不设置默认用系统时间
	tickTime:设置时间轮下个tick执行时间
*/
func NewTimeWheel(plan []int64, userNow func() int64, tickTime int64, cb TWCallBack) *TimeWheel {
	tw := &TimeWheel{}
	if len(plan) == 0 {
		plan = append(plan, 3600, 24, 360, 100) //100 year
	}

	tw.wheel = make([](*TWOneWheel), len(plan))
	for i, v := range plan {
		tw.wheel[i] = &TWOneWheel{}
		if i == 0 {
			tw.wheel[i].scale = 1
		} else {
			tw.wheel[i].scale = tw.wheel[i-1].scale * int64(len(tw.wheel[i-1].slots))
		}
		tw.wheel[i].slots = make([]*list.List, v)
	}
	tw.userNow = userNow
	tw.tickTime = tickTime
	tw.cb = cb
	tw.Index = make(map[int64][]int)
	return tw
}
