package ekafka

import (
	"context"
	"errors"
	eventbus "github.com/SyaibanAhmadRamadhan/event-bus"
	"github.com/segmentio/kafka-go"
)

type MarshalFunc func(any) ([]byte, error)
type UnmarshalFunc func([]byte, any) error

type Reader struct {
	R            *kafka.Reader
	subTracer    TracerSub
	commitTracer TracerCommitMessage
	groupID      string
	marshal      MarshalFunc
	unmarshal    UnmarshalFunc
}

func (r *Reader) FetchMessage(ctx context.Context, v any) (kafka.Message, error) {
	msg, err := r.R.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, eventbus.Error(err)
	}
	var ctxOtel context.Context
	if r.subTracer != nil {
		ctxOtel = r.subTracer.TraceSubStart(ctx, r.groupID, &msg)
	}

	if v != nil {
		err = r.unmarshal(msg.Value, v)
		if err != nil {
			err = eventbus.Error(errors.Join(err, ErrJsonUnmarshal))
		}
	}

	if r.subTracer != nil {
		r.subTracer.TraceSubEnd(ctxOtel, err)
	}

	return msg, err
}

func (r *Reader) ReadMessage(ctx context.Context, v any) (kafka.Message, error) {
	msg, err := r.R.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, eventbus.Error(err)
	}

	var ctxOtel context.Context

	if r.subTracer != nil {
		ctxOtel = r.subTracer.TraceSubStart(ctx, r.groupID, &msg)
	}

	if v != nil {
		err = r.unmarshal(msg.Value, v)
		if err != nil {
			err = eventbus.Error(errors.Join(err, ErrJsonUnmarshal))
		}
	}

	if r.subTracer != nil {
		r.subTracer.TraceSubEnd(ctxOtel, err)
	}
	return msg, err
}

func (r *Reader) CommitMessages(ctx context.Context, messages ...kafka.Message) error {
	if messages == nil || len(messages) <= 0 {
		return nil
	}

	contexts := make([]context.Context, 0)
	if r.commitTracer != nil {
		contexts = r.commitTracer.TraceCommitMessagesStart(ctx, r.groupID, messages...)
	}

	err := r.R.CommitMessages(ctx, messages...)
	if err != nil {
		err = eventbus.Error(err)
	}

	if r.commitTracer != nil {
		r.commitTracer.TraceCommitMessagesEnd(contexts, err)
	}

	return err
}
