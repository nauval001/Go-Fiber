package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type ReviewService struct {
	reviews repository.ReviewRepository
	courses repository.CourseRepository
	perms   *helper.PermissionSet
}

func NewReviewService(reviews repository.ReviewRepository, courses repository.CourseRepository, perms *helper.PermissionSet) *ReviewService {
	return &ReviewService{reviews: reviews, courses: courses, perms: perms}
}

// Create (POST /courses/:id/reviews)
func (s *ReviewService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id kelas harus berupa angka positif")
	}

	var req model.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}

	req.Comment = strings.TrimSpace(req.Comment)
	req.Tags = NormalizeTags(req.Tags)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	created, err := s.reviews.Create(ctx, model.Review{
		CourseID: courseID,
		UserID:   current.UserID,
		Username: current.Username,
		Rating:   req.Rating,
		Comment:  req.Comment,
		Tags:     req.Tags,
	})
	if err != nil {
		return translateReviewError(err)
	}

	return helper.Created(c, "ulasan berhasil disimpan", created, "/api/v1/reviews/"+created.ID)
}

// ListByCourse (GET /courses/:id/reviews)
func (s *ReviewService) ListByCourse(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id kelas harus berupa angka positif")
	}

	q := model.ReviewQuery{
		CourseID: courseID,
		Tag:      strings.ToLower(strings.TrimSpace(c.Query("tag"))),
		Before:   c.Query("before"),
		Limit:    c.QueryInt("limit", 10),
	}

	if q.Limit < 1 || q.Limit > 50 {
		q.Limit = 10
	}

	want := q.Limit
	q.Limit = want + 1

	reviews, err := s.reviews.FindByCourse(ctx, q)
	if err != nil {
		return translateReviewError(err)
	}

	hasMore := len(reviews) > want
	if hasMore {
		reviews = reviews[:want]
	}

	next := ""
	if hasMore && len(reviews) > 0 {
		next = reviews[len(reviews)-1].ID
	}

	return c.JSON(model.ReviewPage{
		Success: true,
		Message: "daftar ulasan berhasil diambil",
		Data:    reviews,
		Meta: model.ReviewPageMeta{
			Limit:      want,
			HasMore:    hasMore,
			NextBefore: next,
		},
	})
}

// translateReviewError memetakan kegagalan repository ke response HTTP.
func translateReviewError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id ulasan tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.NewError(fiber.StatusConflict, "ALREADY_REVIEWED", "Anda sudah menulis ulasan untuk kelas ini, ubah ulasan yang ada")
	case errors.Is(err, repository.ErrNotOwner):
		return helper.Forbidden("hanya pemilik ulasan yang boleh melakukan ini")
	case errors.Is(err, repository.ErrOwnReview):
		return helper.NewError(fiber.StatusConflict, "OWN_REVIEW", "tidak dapat menilai ulasan sendiri")
	case errors.Is(err, repository.ErrAlreadyVoted):
		return helper.NewError(fiber.StatusConflict, "ALREADY_VOTED", "Anda sudah menilai ulasan ini")
	case errors.Is(err, repository.ErrReplyLimit):
		return helper.NewError(fiber.StatusConflict, "REPLY_LIMIT", "jumlah balasan sudah mencapai batas")
	default:
		return helper.Internal(err)
	}
}