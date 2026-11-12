package repository

import (
	"context"
	"errors"

	"api-students/app/model"
)

var (
	ErrInvalidID    = errors.New("format id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak boleh menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah pernah menilai ulasan ini")
	ErrReplyLimit   = errors.New("jumlah balasan sudah mencapai batas")
)

const MaxRepliesPerReview = 50

type ReviewRepository interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, r model.Review) (model.Review, error)
	FindByID(ctx context.Context, id string) (model.Review, error)
	FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error)
	UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error)
	Delete(ctx context.Context, id string, userID int, anyOwner bool) error
	AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error)
	MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error)
	Summary(ctx context.Context, courseID int) (model.ReviewSummary, error)
}