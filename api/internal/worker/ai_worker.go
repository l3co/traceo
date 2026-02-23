package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type JobType string

const (
	JobAgeProgression JobType = "age_progression"
	JobFaceMatching   JobType = "face_matching"
)

type AIJob struct {
	Type      JobType
	TargetID  string
	PhotoURL  string
	BirthDate time.Time
}

type JobProcessor interface {
	ProcessAgeProgression(ctx context.Context, missingID, photoURL string, birthDate time.Time) error
	ProcessFaceMatching(ctx context.Context, homelessID string) error
}

type AIWorker struct {
	jobs      chan AIJob
	processor JobProcessor
	wg        sync.WaitGroup
}

func NewAIWorker(processor JobProcessor, concurrency int) *AIWorker {
	w := &AIWorker{
		jobs:      make(chan AIJob, 100),
		processor: processor,
	}

	for i := range concurrency {
		w.wg.Add(1)
		go w.run(i)
	}

	return w
}

func (w *AIWorker) run(id int) {
	defer w.wg.Done()
	slog.Info("ai worker started", slog.Int("worker_id", id))

	for job := range w.jobs {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

		slog.Info("processing ai job",
			slog.String("type", string(job.Type)),
			slog.String("target_id", job.TargetID),
			slog.Int("worker_id", id),
		)

		var err error
		switch job.Type {
		case JobAgeProgression:
			err = w.processor.ProcessAgeProgression(ctx, job.TargetID, job.PhotoURL, job.BirthDate)
		case JobFaceMatching:
			err = w.processor.ProcessFaceMatching(ctx, job.TargetID)
		}

		if err != nil {
			slog.Error("ai job failed",
				slog.String("type", string(job.Type)),
				slog.String("target_id", job.TargetID),
				slog.String("error", err.Error()),
			)
		}

		cancel()
	}
}

func (w *AIWorker) Enqueue(job AIJob) {
	select {
	case w.jobs <- job:
	default:
		slog.Warn("ai job queue full, dropping job",
			slog.String("type", string(job.Type)),
			slog.String("target_id", job.TargetID),
		)
	}
}

func (w *AIWorker) Shutdown() {
	close(w.jobs)
	w.wg.Wait()
	slog.Info("ai worker shut down")
}
