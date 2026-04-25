package logstream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/raftweave/raftweave/internal/build/domain"
	"github.com/redis/go-redis/v9"
)

// Subscriber receives log lines in real time.
type Subscriber interface {
	Send(line *domain.LogLine) error
	Done() <-chan struct{}
}

type chanSubscriber struct {
	ch   chan *domain.LogLine
	done chan struct{}
}

func (s *chanSubscriber) Send(line *domain.LogLine) error {
	select {
	case s.ch <- line:
		return nil
	case <-s.done:
		return fmt.Errorf("subscriber closed")
	default:
		return fmt.Errorf("subscriber buffer full")
	}
}

func (s *chanSubscriber) Done() <-chan struct{} {
	return s.done
}

// Broadcaster routes log lines from a producer to multiple subscribers.
type Broadcaster interface {
	// Publish sends a log line to all active subscribers for this buildID.
	Publish(ctx context.Context, line *domain.LogLine) error

	// Subscribe registers a subscriber for a buildID. Returns an unsubscribe func.
	Subscribe(ctx context.Context, buildID string) (Subscriber, func(), error)

	// MarkComplete signals end-of-stream to all subscribers of a buildID.
	MarkComplete(ctx context.Context, buildID string)
}

type memBroadcaster struct {
	mu          sync.RWMutex
	subscribers map[string][]*chanSubscriber
}

// New returns an in-memory broadcaster backed by sync.Map and channels.
func New() Broadcaster {
	return &memBroadcaster{
		subscribers: make(map[string][]*chanSubscriber),
	}
}

func (b *memBroadcaster) Publish(ctx context.Context, line *domain.LogLine) error {
	b.mu.RLock()
	subs := b.subscribers[line.BuildID]
	b.mu.RUnlock()

	for _, sub := range subs {
		_ = sub.Send(line)
	}
	return nil
}

func (b *memBroadcaster) Subscribe(ctx context.Context, buildID string) (Subscriber, func(), error) {
	sub := &chanSubscriber{
		ch:   make(chan *domain.LogLine, 100),
		done: make(chan struct{}),
	}

	b.mu.Lock()
	b.subscribers[buildID] = append(b.subscribers[buildID], sub)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[buildID]
		for i, s := range subs {
			if s == sub {
				b.subscribers[buildID] = append(subs[:i], subs[i+1:]...)
				close(sub.done)
				break
			}
		}
	}

	return sub, unsubscribe, nil
}

func (b *memBroadcaster) MarkComplete(ctx context.Context, buildID string) {
	b.mu.RLock()
	subs := b.subscribers[buildID]
	b.mu.RUnlock()

	for range subs {
	}
}

type redisBroadcaster struct {
	client *redis.Client
}

// NewRedis returns a broadcaster backed by Redis Pub/Sub.
func NewRedis(redisAddr string) Broadcaster {
	return &redisBroadcaster{
		client: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
	}
}

func (b *redisBroadcaster) Publish(ctx context.Context, line *domain.LogLine) error {
	data, err := json.Marshal(line)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, fmt.Sprintf("raftweave:build:logs:%s", line.BuildID), data).Err()
}

func (b *redisBroadcaster) Subscribe(ctx context.Context, buildID string) (Subscriber, func(), error) {
	pubsub := b.client.Subscribe(ctx, fmt.Sprintf("raftweave:build:logs:%s", buildID))
	sub := &chanSubscriber{
		ch:   make(chan *domain.LogLine, 100),
		done: make(chan struct{}),
	}

	go func() {
		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var line domain.LogLine
				if err := json.Unmarshal([]byte(msg.Payload), &line); err == nil {
					_ = sub.Send(&line)
				}
			case <-sub.done:
				pubsub.Close()
				return
			}
		}
	}()

	unsubscribe := func() {
		close(sub.done)
	}

	return sub, unsubscribe, nil
}

func (b *redisBroadcaster) MarkComplete(ctx context.Context, buildID string) {
	// Signal end of stream if needed
}
