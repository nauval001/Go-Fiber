package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var (
	ErrSupervisorFull     = errors.New("kuota pembimbing penuh")
	ErrNotAssigned        = errors.New("tidak ada bimbingan")
	ErrVersionConflict    = errors.New("data sudah diubah pihak lain")
	ErrQuotaBelowFilled   = errors.New("kuota lebih kecil dari jumlah bimbingan")
)

type SupervisorRepository interface {
	WithTx(tx DBTX) SupervisorRepository
	LockStudent(ctx context.Context, studentID int) error
	LockSupervisors(ctx context.Context, ids ...int) error
	IncrementSupervised(ctx context.Context, supervisorID int) error
	DecrementSupervised(ctx context.Context, supervisorID int) error
	InsertSupervision(ctx context.Context, studentID, supervisorID int) error
	DeleteSupervision(ctx context.Context, studentID int) error
	UpdateQuota(ctx context.Context, id, quota, version int) error
}

type supervisorPostgresRepository struct {
	db DBTX
}

func NewSupervisorRepository(db DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: db}
}

func (r *supervisorPostgresRepository) WithTx(tx DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: tx}
}

// LockStudent mengunci baris mahasiswa agar tidak bisa mendaftar 2 pembimbing sekaligus
func (r *supervisorPostgresRepository) LockStudent(ctx context.Context, studentID int) error {
	var id int
	// PERBAIKAN BUG: Tambahkan FOR UPDATE
	err := r.db.QueryRow(ctx, "SELECT id FROM students WHERE id = $1 FOR UPDATE", studentID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// LockSupervisors mengunci beberapa pembimbing dengan urutan ID yang konsisten (mencegah deadlock)
func (r *supervisorPostgresRepository) LockSupervisors(ctx context.Context, ids ...int) error {
	rows, err := r.db.Query(ctx, "SELECT id FROM supervisors WHERE id = ANY($1) ORDER BY id FOR UPDATE", ids)
	if err != nil { return err }
	defer rows.Close()
	
	found := 0
	for rows.Next() { found++ }
	if found != len(ids) { return ErrNotFound }
	return nil
}

func (r *supervisorPostgresRepository) IncrementSupervised(ctx context.Context, supervisorID int) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE supervisors SET supervised_count = supervised_count + 1 
		WHERE id = $1 AND supervised_count < quota`, supervisorID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return ErrSupervisorFull }
	return nil
}

func (r *supervisorPostgresRepository) DecrementSupervised(ctx context.Context, supervisorID int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE supervisors SET supervised_count = supervised_count - 1 
		WHERE id = $1`, supervisorID)
	return err
}

func (r *supervisorPostgresRepository) InsertSupervision(ctx context.Context, studentID, supervisorID int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO supervisions (student_id, supervisor_id) VALUES ($1, $2)`, 
		studentID, supervisorID)
	if pgErrorCode(err) == codeUniqueViolation {
		return ErrDuplicate // Mahasiswa sudah punya pembimbing
	}
	return err
}

func (r *supervisorPostgresRepository) DeleteSupervision(ctx context.Context, studentID int) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM supervisions WHERE student_id = $1", studentID)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return ErrNotAssigned }
	return nil
}

func (r *supervisorPostgresRepository) UpdateQuota(ctx context.Context, id, quota, version int) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE supervisors SET quota = $1, version = version + 1 
		WHERE id = $2 AND version = $3 AND $1 >= supervised_count`, quota, id, version)
	if err != nil { return err }
	if tag.RowsAffected() == 0 { return ErrVersionConflict } // Bisa juga karena quota < filled
	return nil
}