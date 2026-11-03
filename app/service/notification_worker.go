package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// NotificationWorker mengerjakan task "notification:send".
type NotificationWorker struct {
	users         repository.UserRepository
	notifications repository.NotificationRepository
	courses       repository.CourseRepository // Untuk Tugas D.2
	tx            *repository.TxManager
	mailer        helper.Mailer
	logger        *slog.Logger
}

func NewNotificationWorker(
	users repository.UserRepository,
	notifications repository.NotificationRepository,
	courses repository.CourseRepository,
	tx *repository.TxManager,
	mailer helper.Mailer,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		users: users, notifications: notifications, courses: courses,
		tx: tx, mailer: mailer, logger: logger,
	}
}

// HandleSend mengembalikan error untuk meminta Asynq melakukan retry nanti,
// atau error yang membungkus asynq.SkipRetry untuk menyatakan "jangan di-retry".
func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}

	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}

	// JIKA EVENT ADALAH PERUBAHAN KUOTA (Tugas Mandiri D.2)
	if job.EventType == model.EventQuotaChanged {
		return w.handleQuotaChanged(ctx, job, event)
	}

	// JIKA EVENT ADALAH PERUBAHAN KRS (Single User)
	user, err := w.users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err
	}

	msg, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID, UserID: user.ID, Channel: "email", Subject: msg.Subject, Body: msg.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})

	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %w", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID),
		slog.String("event", job.EventType),
		slog.Int("retry", retry),
		slog.Bool("already_sent", alreadySent))
	return nil
}

// handleQuotaChanged mengirim email ke SELURUH mahasiswa di kelas tersebut.
func (w *NotificationWorker) handleQuotaChanged(ctx context.Context, job model.NotificationJob, event model.EnrollmentEvent) error {
	userIDs, err := w.courses.GetEnrolledUserIDs(ctx, event.CourseID)
	if err != nil {
		return err
	}

	for _, uid := range userIDs {
		user, err := w.users.FindByID(ctx, uid)
		if err != nil {
			continue // Skip jika user dihapus
		}

		msg, _ := BuildMessage(job.EventType, event, user)

		err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
			claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
				OutboxID: job.OutboxID, UserID: user.ID, Channel: "email", Subject: msg.Subject, Body: msg.Body,
			})
			if err != nil {
				return err
			}
			if !claimed {
				return nil // Sudah dikirim (Idempotent)
			}
			return w.mailer.Send(ctx, msg)
		})
		
		if errors.Is(err, helper.ErrPermanentDelivery) {
			// Jika 1 email gagal permanen, lanjut ke mahasiswa lain
			w.logger.Warn("email_gagal_permanen", slog.Int("user_id", uid), slog.String("error", err.Error()))
			continue 
		}
		if err != nil {
			return err // Jika error sementara, batalkan dan biarkan Asynq retry
		}
	}

	w.logger.Info("quota_notifications_processed", slog.Int64("outbox_id", job.OutboxID), slog.Int("total_users", len(userIDs)))
	return nil
}