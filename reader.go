package disruptor

import "time"

type Reader struct {
	read     *Cursor
	written  *Cursor
	upstream Barrier
	consumer Consumer
	ready    bool
} // TODO: padding???

func NewReader(read, written *Cursor, upstream Barrier, consumer Consumer) *Reader {
	return &Reader{
		read:     read,
		written:  written,
		upstream: upstream,
		consumer: consumer,
		ready:    false,
	}
}

func (this *Reader) Start() {
	this.ready = true
	go this.receive()
}
func (this *Reader) Stop() {
	this.ready = false
}

func (this *Reader) receive() {
	previous := this.read.Sequence
	idling, gating := 0, 0

	for {
		lower := previous + 1
		upper := this.upstream.Read(lower)

		if lower <= upper {
			this.consumer.Consume(lower, upper)
			this.read.Sequence = upper
			previous = upper
		} else if upper = this.written.Load(); lower <= upper {
			// Gating--TODO: wait strategy (provide gating count to wait strategy for phased backoff)
			gating++
			idling = 0
		} else if this.ready {
			// Idling--TODO: wait strategy (provide idling count to wait strategy for phased backoff)
			idling++
			gating = 0
		} else {
			break
		}

		time.Sleep(time.Nanosecond)
	}
}
