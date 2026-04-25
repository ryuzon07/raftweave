package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/raftweave/raftweave/internal/build/domain"
)

type BuildRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func NewBuildRepository(pool *pgxpool.Pool, tracer trace.Tracer) *BuildRepository {
	return &BuildRepository{
		pool:   pool,
		tracer: tracer,
	}
}

func (r *BuildRepository) Create(ctx context.Context, b *domain.Build) error {
	ctx, span := r.tracer.Start(ctx, "db.build.create", trace.WithAttributes(
		attribute.String("build_id", b.ID),
	))
	defer span.End()

	q := `
		INSERT INTO builds (
			id, workload_id, workspace_id, git_commit_sha, git_branch, status, language,
			created_at, updated_at
		) VALUES (
			@id, @workload_id, @workspace_id, @git_commit_sha, @git_branch, @status, @language,
			NOW(), NOW()
		)
	`

	args := pgx.NamedArgs{
		"id":             b.ID,
		"workload_id":    b.WorkloadID,
		"workspace_id":   b.WorkspaceID,
		"git_commit_sha": b.GitCommitSHA,
		"git_branch":     b.GitBranch,
		"status":         string(b.Status),
		"language":       string(b.Language),
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, q, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *BuildRepository) GetByID(ctx context.Context, id string) (*domain.Build, error) {
	ctx, span := r.tracer.Start(ctx, "db.build.get_by_id", trace.WithAttributes(
		attribute.String("build_id", id),
	))
	defer span.End()

	q := `SELECT id, workload_id, workspace_id, git_commit_sha, git_branch, status, language, image_ref, image_digest, size_bytes, error_message, started_at, completed_at, created_at, updated_at FROM builds WHERE id = @id`
	args := pgx.NamedArgs{"id": id}

	row := r.pool.QueryRow(ctx, q, args)
	b := &domain.Build{}

	var lang, imageRef, imageDigest, errMsg *string
	var sizeBytes *int64
	var startedAt, completedAt *time.Time

	err := row.Scan(
		&b.ID, &b.WorkloadID, &b.WorkspaceID, &b.GitCommitSHA, &b.GitBranch, &b.Status,
		&lang, &imageRef, &imageDigest, &sizeBytes, &errMsg,
		&startedAt, &completedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBuildNotFound
		}
		return nil, err
	}
	if startedAt != nil {
		b.StartedAt = *startedAt
	}
	b.CompletedAt = completedAt
	if lang != nil {
		b.Language = domain.Language(*lang)
	}
	if imageRef != nil {
		b.ImageRef = *imageRef
	}
	if imageDigest != nil {
		b.ImageDigest = *imageDigest
	}
	if sizeBytes != nil {
		b.SizeBytes = *sizeBytes
	}
	if errMsg != nil {
		b.ErrorMessage = *errMsg
	}

	return b, nil
}

func (r *BuildRepository) ListByWorkload(ctx context.Context, workloadID string, limit, offset int) ([]*domain.Build, error) {
	ctx, span := r.tracer.Start(ctx, "db.build.list_by_workload", trace.WithAttributes(
		attribute.String("workload_id", workloadID),
	))
	defer span.End()

	q := `
		SELECT id, workload_id, workspace_id, git_commit_sha, git_branch, status, language, image_ref, image_digest, size_bytes, error_message, started_at, completed_at, created_at, updated_at
		FROM builds
		WHERE workload_id = @workload_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`
	args := pgx.NamedArgs{
		"workload_id": workloadID,
		"limit":       limit,
		"offset":      offset,
	}

	rows, err := r.pool.Query(ctx, q, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []*domain.Build
	for rows.Next() {
		b := &domain.Build{}
		var lang, imageRef, imageDigest, errMsg *string
		var sizeBytes *int64
		var startedAt, completedAt *time.Time

		err := rows.Scan(
			&b.ID, &b.WorkloadID, &b.WorkspaceID, &b.GitCommitSHA, &b.GitBranch, &b.Status,
			&lang, &imageRef, &imageDigest, &sizeBytes, &errMsg,
			&startedAt, &completedAt, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if startedAt != nil {
			b.StartedAt = *startedAt
		}
		b.CompletedAt = completedAt

		if lang != nil {
			b.Language = domain.Language(*lang)
		}
		if imageRef != nil {
			b.ImageRef = *imageRef
		}
		if imageDigest != nil {
			b.ImageDigest = *imageDigest
		}
		if sizeBytes != nil {
			b.SizeBytes = *sizeBytes
		}
		if errMsg != nil {
			b.ErrorMessage = *errMsg
		}

		builds = append(builds, b)
	}

	return builds, nil
}

func (r *BuildRepository) UpdateStatus(ctx context.Context, id string, status domain.BuildStatus, errMsg string) error {
	ctx, span := r.tracer.Start(ctx, "db.build.update_status", trace.WithAttributes(
		attribute.String("build_id", id),
		attribute.String("new_status", string(status)),
	))
	defer span.End()

	q := `UPDATE builds SET status = @status, error_message = @err_msg, updated_at = NOW() WHERE id = @id`
	args := pgx.NamedArgs{"id": id, "status": string(status), "err_msg": errMsg}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, q, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *BuildRepository) UpdateImageRef(ctx context.Context, id, imageRef, imageDigest string, sizeBytes int64) error {
	ctx, span := r.tracer.Start(ctx, "db.build.update_image_ref", trace.WithAttributes(
		attribute.String("build_id", id),
	))
	defer span.End()

	q := `UPDATE builds SET image_ref = @image_ref, image_digest = @image_digest, size_bytes = @size_bytes, updated_at = NOW() WHERE id = @id`
	args := pgx.NamedArgs{"id": id, "image_ref": imageRef, "image_digest": imageDigest, "size_bytes": sizeBytes}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, q, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *BuildRepository) MarkCompleted(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "db.build.mark_completed", trace.WithAttributes(
		attribute.String("build_id", id),
	))
	defer span.End()

	q := `UPDATE builds SET completed_at = NOW(), status = 'SUCCEEDED', updated_at = NOW() WHERE id = @id`
	args := pgx.NamedArgs{"id": id}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, q, args)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// LogRepository struct and append lines
type LogRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func NewLogRepository(pool *pgxpool.Pool, tracer trace.Tracer) *LogRepository {
	return &LogRepository{pool: pool, tracer: tracer}
}

func (r *LogRepository) AppendLine(ctx context.Context, line *domain.LogLine) error {
	q := `INSERT INTO build_log_lines (build_id, sequence, stream, text, ts) VALUES (@build_id, @sequence, @stream, @text, @ts)`
	_, err := r.pool.Exec(ctx, q, pgx.NamedArgs{
		"build_id": line.BuildID,
		"sequence": line.Sequence,
		"stream":   line.Stream,
		"text":     line.Text,
		"ts":       line.Timestamp,
	})
	return err
}

func (r *LogRepository) GetLines(ctx context.Context, buildID string, fromSeq int64) ([]*domain.LogLine, error) {
	ctx, span := r.tracer.Start(ctx, "db.logs.get_lines", trace.WithAttributes(
		attribute.String("build_id", buildID),
	))
	defer span.End()

	q := `
		SELECT build_id, sequence, stream, text, ts
		FROM build_log_lines
		WHERE build_id = @build_id AND sequence > @from_seq
		ORDER BY sequence ASC
	`
	args := pgx.NamedArgs{
		"build_id":   buildID,
		"from_seq":   fromSeq,
	}

	rows, err := r.pool.Query(ctx, q, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []*domain.LogLine
	for rows.Next() {
		line := &domain.LogLine{}
		err := rows.Scan(&line.BuildID, &line.Sequence, &line.Stream, &line.Text, &line.Timestamp)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, nil
}
