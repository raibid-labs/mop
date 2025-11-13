package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// InvalidationMessage represents a cache invalidation message
type InvalidationMessage struct {
	Key       string    `json:"key"`
	Pattern   string    `json:"pattern,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// PubSub handles Redis pub/sub for cache invalidation
type PubSub struct {
	cache  *Cache
	pubsub *redis.PubSub
	stopCh chan struct{}
}

// NewPubSub creates a new PubSub instance
func NewPubSub(cache *Cache) *PubSub {
	return &PubSub{
		cache:  cache,
		stopCh: make(chan struct{}),
	}
}

// Subscribe starts listening for invalidation messages
func (p *PubSub) Subscribe(ctx context.Context) error {
	p.pubsub = p.cache.client.Subscribe(ctx, InvalidationChannel)

	// Wait for subscription confirmation
	if _, err := p.pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}

	// Start message handler in background
	go p.handleMessages(ctx)

	log.Printf("Subscribed to %s channel for cache invalidation", InvalidationChannel)
	return nil
}

// handleMessages processes incoming invalidation messages
func (p *PubSub) handleMessages(ctx context.Context) {
	ch := p.pubsub.Channel()

	for {
		select {
		case <-p.stopCh:
			log.Println("Stopping pub/sub message handler")
			return
		case <-ctx.Done():
			log.Println("Context cancelled, stopping pub/sub")
			return
		case msg := <-ch:
			if err := p.processMessage(ctx, msg); err != nil {
				log.Printf("Error processing invalidation message: %v", err)
			}
		}
	}
}

// processMessage handles a single invalidation message
func (p *PubSub) processMessage(ctx context.Context, msg *redis.Message) error {
	var invMsg InvalidationMessage
	if err := json.Unmarshal([]byte(msg.Payload), &invMsg); err != nil {
		return fmt.Errorf("unmarshal message failed: %w", err)
	}

	log.Printf("Received invalidation: key=%s, pattern=%s, source=%s",
		invMsg.Key, invMsg.Pattern, invMsg.Source)

	// Invalidate by key or pattern
	if invMsg.Key != "" {
		if err := p.cache.Delete(ctx, invMsg.Key); err != nil {
			return fmt.Errorf("delete key failed: %w", err)
		}
		log.Printf("Invalidated cache key: %s", invMsg.Key)
	}

	if invMsg.Pattern != "" {
		if err := p.cache.DeletePattern(ctx, invMsg.Pattern); err != nil {
			return fmt.Errorf("delete pattern failed: %w", err)
		}
		log.Printf("Invalidated cache pattern: %s", invMsg.Pattern)
	}

	return nil
}

// PublishInvalidation publishes a cache invalidation message
func (p *PubSub) PublishInvalidation(ctx context.Context, key, pattern, source string) error {
	msg := InvalidationMessage{
		Key:       key,
		Pattern:   pattern,
		Timestamp: time.Now(),
		Source:    source,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	if err := p.cache.client.Publish(ctx, InvalidationChannel, data).Err(); err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	log.Printf("Published invalidation: key=%s, pattern=%s, source=%s", key, pattern, source)
	return nil
}

// InvalidateKey publishes an invalidation message for a specific key
func (p *PubSub) InvalidateKey(ctx context.Context, key string) error {
	return p.PublishInvalidation(ctx, key, "", "api")
}

// InvalidatePattern publishes an invalidation message for a key pattern
func (p *PubSub) InvalidatePattern(ctx context.Context, pattern string) error {
	return p.PublishInvalidation(ctx, "", pattern, "api")
}

// Stop stops the pub/sub listener
func (p *PubSub) Stop() error {
	close(p.stopCh)
	if p.pubsub != nil {
		return p.pubsub.Close()
	}
	return nil
}
