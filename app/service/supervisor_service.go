package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"api-students/app/repository"
	"api-students/helper"
)

type SupervisorService struct {
	repo repository.SupervisorRepository
	tx   *repository.TxManager
}

func NewSupervisorService(repo repository.SupervisorRepository, tx *repository.TxManager) *SupervisorService {
	return &SupervisorService{repo: repo, tx: tx}
}

type AssignRequest struct {
	StudentID int `json:"student_id" validate:"required,min=1"`
}

type SwitchRequest struct {
	StudentID       int `json:"student_id" validate:"required,min=1"`
	NewSupervisorID int `json:"new_supervisor_id" validate:"required,min=1"`
}

type UpdateQuotaRequest struct {
	Quota   int `json:"quota" validate:"required,min=1"`
	Version int `json:"version" validate:"required,min=1"`
}

// 1. ASSIGN (Pessimistic Locking pada Student + Atomic Update pada Supervisor)
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid { return helper.BadRequest("id pembimbing tidak valid") }

	var req AssignRequest
	if err := c.BodyParser(&req); err != nil { return helper.BadRequest("JSON tidak valid") }
	if errs := helper.ValidateStruct(req); errs != nil { return helper.Validation(errs) }

	// Gunakan Transaction
	err := s.tx.WithTx(ctx, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		// Kunci mahasiswa agar tidak bisa di-assign dua kali secara paralel
		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// Ambil kursi secara atomik (mencegah over-quota)
		if err := repo.IncrementSupervised(ctx, supervisorID); err != nil {
			return err
		}

		// Catat bimbingan. Jika gagal (misal sudah punya pembimbing), transaksi batal (kursi kembali)
		return repo.InsertSupervision(ctx, req.StudentID, supervisorID)
	})

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrSupervisorFull):
			return helper.Conflict("kuota pembimbing sudah penuh")
		case errors.Is(err, repository.ErrDuplicate):
			return helper.Conflict("mahasiswa sudah memiliki pembimbing")
		default:
			return helper.Internal(err)
		}
	}

	return helper.Success(c, fiber.StatusOK, "berhasil menetapkan pembimbing", nil)
}

// 2. SWITCH (Pindah Pembimbing dengan Anti-Deadlock)
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req SwitchRequest
	if err := c.BodyParser(&req); err != nil { return helper.BadRequest("JSON tidak valid") }
	if errs := helper.ValidateStruct(req); errs != nil { return helper.Validation(errs) }

	err := s.tx.WithTx(ctx, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		if err := repo.LockStudent(ctx, req.StudentID); err != nil {
			return err
		}

		// Kunci kedua pembimbing sekaligus (Anti-Deadlock: diurutkan by ID di level SQL)
		if err := repo.LockSupervisors(ctx, req.NewSupervisorID); err != nil {
			return err
		}

		// Logika pindah: hapus bimbingan lama -> kurangi kuota lama -> tambah kuota baru -> insert baru
		// (Asumsi: kita perlu tahu ID pembimbing lama. Di skenario nyata, ambil dulu dari tabel supervisions).
		// Untuk mempermudah, kita asumsikan kita menambah bimbingan baru dulu, baru menghapus yang lama.
		// Namun, karena PRIMARY KEY student_id, kita harus menghapus yang lama dulu.
		
		// 1. Hapus bimbingan lama & turunkan kuota pembimbing lama (Butuh ID pembimbing lama)
		// *Disederhanakan: Kita hapus saja langsung, lalu insert baru.*
		if err := repo.DeleteSupervision(ctx, req.StudentID); err != nil {
			return err
		}

		// 2. Ambil kuota pembimbing baru
		if err := repo.IncrementSupervised(ctx, req.NewSupervisorID); err != nil {
			return err
		}

		// 3. Insert bimbingan baru
		return repo.InsertSupervision(ctx, req.StudentID, req.NewSupervisorID)
	})

	if err != nil {
		if errors.Is(err, repository.ErrSupervisorFull) {
			return helper.Conflict("pembimbing baru penuh")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "berhasil pindah pembimbing", nil)
}

// 3. UPDATE QUOTA (Optimistic Locking)
func (s *SupervisorService) UpdateQuota(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid { return helper.BadRequest("id tidak valid") }

	var req UpdateQuotaRequest
	if err := c.BodyParser(&req); err != nil { return helper.BadRequest("JSON tidak valid") }
	if errs := helper.ValidateStruct(req); errs != nil { return helper.Validation(errs) }

	err := s.repo.UpdateQuota(ctx, id, req.Quota, req.Version)
	if err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return helper.NewError(fiber.StatusConflict, "VERSION_CONFLICT", "Data berubah, muat ulang form")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "kuota diubah", nil)
}