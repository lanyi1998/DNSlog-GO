package sdk

import (
	"time"
)

type Request struct {
	Key        string    // 请求的参数
	ResultChan chan bool // 用于接收结果的专属通道
}
type KeyPool struct {
	queue     chan *Request // 接收请求的队列 (Key Pool)
	batchSize int           // 每次处理的最大数量（防止一次太多）
	interval  time.Duration // 处理间隔
	Client    *DnsLogClient
}

func NewKeyPool(bufferSize int, interval time.Duration, client *DnsLogClient) *KeyPool {
	return &KeyPool{
		queue:     make(chan *Request, bufferSize), // 带缓冲的通道
		batchSize: bufferSize,                      // 假设每次最多处理100个
		interval:  interval,
		Client:    client,
	}
}

// Start 启动后台处理协程 (Worker)
func (p *KeyPool) Start() {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.processBatch()
			}
		}
	}()
}

func (p *KeyPool) CloseChan(batch []*Request) {
	for _, req := range batch {
		close(req.ResultChan)
	}
}

// processBatch 从队列中取出当前积累的所有请求并处理
func (p *KeyPool) processBatch() {
	var batch []*Request
	var keyList []string
	for i := 0; i < p.batchSize; i++ {
		select {
		case req := <-p.queue:
			batch = append(batch, req)
			keyList = append(keyList, req.Key)
		default:
			break
		}
	}
	// 如果没有请求，直接返回
	if len(batch) == 0 {
		return
	}
	result, err := p.Client.BulkVerifyDns(keyList)
	if err != nil {
		p.CloseChan(batch)
		return
	}
	if len(result) == 0 {
		p.CloseChan(batch)
		return
	}
	successKey := make(map[string]struct{})
	for _, key := range result {
		for _, req := range batch {
			if _, ok := successKey[key]; ok {
				continue
			}
			if req.Key == key {
				successKey[key] = struct{}{}
				req.ResultChan <- true
			}
		}
	}
	p.CloseChan(batch)
}

func (p *KeyPool) DoRequest(key string) bool {
	// 1. 创建一个接收结果的通道
	resultChan := make(chan bool, 1) // buffered 1 防止接收方意外没读导致死锁

	// 2. 构造请求放入 Pool
	req := &Request{
		Key:        key,
		ResultChan: resultChan,
	}

	// 放入队列
	p.queue <- req

	// 3. 阻塞等待结果 (Blocked here)
	// 这里会一直停住，直到 Worker 处理完并通过 resultChan 发回数据
	result := <-resultChan
	return result
}
