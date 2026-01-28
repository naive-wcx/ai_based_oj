package queue

import (
	"log"
	"sync"
	"time"

	"oj-system/internal/model"
)

// JudgeTask 判题任务
type JudgeTask struct {
	Submission *model.Submission
	Problem    *model.Problem
	Testcases  []model.Testcase
}

// JudgeQueue 判题队列
type JudgeQueue struct {
	tasks    chan *JudgeTask
	mu       sync.Mutex
	running  bool
	handlers []func(*JudgeTask)
}

var queue *JudgeQueue

// Init 初始化判题队列
func Init(bufferSize int) {
	queue = &JudgeQueue{
		tasks:    make(chan *JudgeTask, bufferSize),
		handlers: make([]func(*JudgeTask), 0),
	}
}

// GetQueue 获取队列实例
func GetQueue() *JudgeQueue {
	return queue
}

// Push 添加判题任务
func (q *JudgeQueue) Push(task *JudgeTask) {
	select {
	case q.tasks <- task:
		log.Printf("[Queue] 添加判题任务: submission_id=%d", task.Submission.ID)
	default:
		log.Printf("[Queue] 队列已满，丢弃任务: submission_id=%d", task.Submission.ID)
	}
}

// RegisterHandler 注册判题处理器
func (q *JudgeQueue) RegisterHandler(handler func(*JudgeTask)) {
	q.handlers = append(q.handlers, handler)
}

// Start 启动队列处理
func (q *JudgeQueue) Start(workers int) {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return
	}
	q.running = true
	q.mu.Unlock()

	log.Printf("[Queue] 启动判题队列，workers=%d", workers)

	for i := 0; i < workers; i++ {
		go q.worker(i)
	}
}

// worker 工作协程
func (q *JudgeQueue) worker(id int) {
	log.Printf("[Worker-%d] 启动", id)
	for task := range q.tasks {
		log.Printf("[Worker-%d] 处理任务: submission_id=%d", id, task.Submission.ID)
		
		startTime := time.Now()
		for _, handler := range q.handlers {
			handler(task)
		}
		elapsed := time.Since(startTime)
		
		log.Printf("[Worker-%d] 任务完成: submission_id=%d, 耗时=%v", id, task.Submission.ID, elapsed)
	}
}

// Stop 停止队列
func (q *JudgeQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if !q.running {
		return
	}
	
	close(q.tasks)
	q.running = false
	log.Printf("[Queue] 判题队列已停止")
}

// Size 获取队列大小
func (q *JudgeQueue) Size() int {
	return len(q.tasks)
}
