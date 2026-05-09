package queue

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	mqv1 "telemetry-collector/api/mq/v1"
	"telemetry-collector/internal/infrastructure/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpcMQClient consumes mq.v1.MessageQueueService via JoinGroup, GetAssignment, Fetch, CommitOffset, Heartbeat, LeaveGroup.
// At most one in-flight message per partition so offsets stay consistent with concurrent workers.
type grpcMQClient struct {
	mu sync.Mutex

	cli          mqv1.MessageQueueServiceClient
	conn         *grpc.ClientConn
	group        string
	topic        string
	memberID     string
	generationID string

	partitions []int32
	partCursor int
	nextOffset map[int32]int64
	inFlight   map[int32]struct{}

	stopHB    chan struct{}
	hbStarted bool
	hbDone    sync.WaitGroup
}

// NewGRPCMQClient returns a queue.Client that consumes mq.v1.MessageQueueService (JoinGroup, Fetch, CommitOffset, …).
// The returned client also implements interface{ Close() error } for graceful shutdown.
func NewGRPCMQClient(cfg config.Config) (Client, error) {
	return dialGRPCMQ(cfg)
}

func dialGRPCMQ(cfg config.Config) (*grpcMQClient, error) {
	addr := strings.TrimSpace(cfg.MQGRPCAddr)
	if addr == "" {
		return nil, fmt.Errorf("MQ_GRPC_ADDR is required for grpc queue client")
	}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if cfg.MQGRPCPreferIPv4 {
		opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp4", address)
		}))
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %q: %w", addr, err)
	}

	c := &grpcMQClient{
		conn:       conn,
		cli:        mqv1.NewMessageQueueServiceClient(conn),
		group:      strings.TrimSpace(cfg.MQGroup),
		topic:      strings.TrimSpace(cfg.MQTopic),
		memberID:   mqMemberID(cfg),
		nextOffset: make(map[int32]int64),
		inFlight:   make(map[int32]struct{}),
		stopHB:     make(chan struct{}),
	}
	if c.group == "" {
		_ = conn.Close()
		return nil, fmt.Errorf("MQ_GROUP is empty")
	}
	if c.topic == "" {
		_ = conn.Close()
		return nil, fmt.Errorf("MQ_TOPIC is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.joinAssignLocked(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}

	c.hbStarted = true
	c.hbDone.Add(1)
	go c.runHeartbeat(cfg.MQHeartbeatInterval)

	log.Printf("mq grpc: connected %q group=%q topic=%q member_id=%q partitions=%v", addr, c.group, c.topic, c.memberID, c.partitions)
	return c, nil
}

func mqMemberID(cfg config.Config) string {
	if s := strings.TrimSpace(cfg.MQMemberID); s != "" {
		return s
	}
	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s-%d-%d", host, os.Getpid(), time.Now().UnixNano())
}

func (c *grpcMQClient) joinAssignLocked(ctx context.Context) error {
	joinResp, err := c.cli.JoinGroup(ctx, &mqv1.JoinRequest{
		Group:    c.group,
		Topic:    c.topic,
		MemberId: c.memberID,
	})
	if err != nil {
		return fmt.Errorf("JoinGroup: %w", err)
	}
	c.generationID = joinResp.GetGenerationId()

	assignResp, err := c.cli.GetAssignment(ctx, &mqv1.AssignRequest{
		Group:        c.group,
		MemberId:     c.memberID,
		GenerationId: c.generationID,
	})
	if err != nil {
		return fmt.Errorf("GetAssignment: %w", err)
	}

	c.inFlight = make(map[int32]struct{})
	c.nextOffset = make(map[int32]int64)
	c.partitions = c.partitions[:0]
	c.partCursor = 0

	for _, a := range assignResp.GetAssignments() {
		p := a.GetPartition()
		c.partitions = append(c.partitions, p)
		c.nextOffset[p] = a.GetStartOffset()
	}
	if len(c.partitions) == 0 {
		return fmt.Errorf("GetAssignment: no partitions assigned to this member")
	}
	return nil
}

func (c *grpcMQClient) rejoinLocked(ctx context.Context) error {
	if _, err := c.cli.LeaveGroup(ctx, &mqv1.LeaveRequest{
		Group:    c.group,
		MemberId: c.memberID,
	}); err != nil {
		log.Printf("mq grpc: LeaveGroup (rebalance): %v", err)
	}
	return c.joinAssignLocked(ctx)
}

func (c *grpcMQClient) runHeartbeat(interval time.Duration) {
	defer c.hbDone.Done()
	if interval <= 0 {
		interval = 3 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-c.stopHB:
			return
		case <-t.C:
			c.mu.Lock()
			if c.cli == nil {
				c.mu.Unlock()
				return
			}
			resp, err := c.cli.Heartbeat(context.Background(), &mqv1.HeartbeatRequest{
				Group:        c.group,
				MemberId:     c.memberID,
				GenerationId: c.generationID,
			})
			if err != nil {
				log.Printf("mq grpc: heartbeat: %v", err)
				c.mu.Unlock()
				continue
			}
			if resp.GetRebalanceNeeded() {
				if err := c.rejoinLocked(context.Background()); err != nil {
					log.Printf("mq grpc: rejoin after rebalance: %v", err)
				} else {
					log.Printf("mq grpc: rejoined generation_id=%q partitions=%v", c.generationID, c.partitions)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Close stops heartbeats, leaves the consumer group, and closes the gRPC connection.
func (c *grpcMQClient) Close() error {
	if c.hbStarted {
		close(c.stopHB)
		c.hbDone.Wait()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cli != nil {
		if _, err := c.cli.LeaveGroup(context.Background(), &mqv1.LeaveRequest{
			Group:    c.group,
			MemberId: c.memberID,
		}); err != nil {
			log.Printf("mq grpc: LeaveGroup: %v", err)
		}
		c.cli = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *grpcMQClient) Pull(ctx context.Context, batchSize int) ([]Message, error) {
	if batchSize <= 0 {
		batchSize = 1
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cli == nil {
		return nil, fmt.Errorf("mq grpc: client closed")
	}
	if len(c.partitions) == 0 {
		return nil, nil
	}

	out := make([]Message, 0, batchSize)
	for len(out) < batchSize {
		picked := false
		for k := 0; k < len(c.partitions); k++ {
			idx := (c.partCursor + k) % len(c.partitions)
			p := c.partitions[idx]
			if _, busy := c.inFlight[p]; busy {
				continue
			}
			off := c.nextOffset[p]
			resp, err := c.cli.Fetch(ctx, &mqv1.FetchRequest{
				Group:     c.group,
				Topic:     c.topic,
				Partition: p,
				Offset:    off,
				Max:       1,
			})
			if err != nil {
				return out, fmt.Errorf("mq grpc Fetch partition=%d offset=%d: %w", p, off, err)
			}
			msgs := resp.GetMessages()
			if len(msgs) == 0 {
				continue
			}
			m := msgs[0]
			c.inFlight[p] = struct{}{}
			c.partCursor = (idx + 1) % len(c.partitions)
			out = append(out, &grpcMQMessage{
				client:    c,
				partition: m.GetPartition(),
				offset:    m.GetOffset(),
				payload:   append([]byte(nil), m.GetPayload()...),
			})
			picked = true
			break
		}
		if !picked {
			break
		}
	}
	return out, nil
}

type grpcMQMessage struct {
	client    *grpcMQClient
	partition int32
	offset    int64
	payload   []byte
}

func (m *grpcMQMessage) Body() []byte { return m.payload }

func (m *grpcMQMessage) Ack(ctx context.Context) error {
	return m.client.commitInFlight(ctx, m.partition, m.offset, true)
}

func (m *grpcMQMessage) Retry(context.Context, time.Duration) error {
	m.client.clearInFlight(m.partition)
	return nil
}

func (m *grpcMQMessage) Reject(ctx context.Context) error {
	return m.client.commitInFlight(ctx, m.partition, m.offset, true)
}

func (c *grpcMQClient) clearInFlight(partition int32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.inFlight, partition)
}

func (c *grpcMQClient) commitInFlight(ctx context.Context, partition int32, offset int64, advance bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cli == nil {
		return fmt.Errorf("mq grpc: client closed")
	}
	_, err := c.cli.CommitOffset(ctx, &mqv1.CommitRequest{
		Group:     c.group,
		Topic:     c.topic,
		Partition: partition,
		Offset:    offset,
	})
	if err != nil {
		return err
	}
	delete(c.inFlight, partition)
	if advance {
		c.nextOffset[partition] = offset + 1
	}
	return nil
}
