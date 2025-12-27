package document

import (
	"context"
	"errors"
	"fmt"

	documentDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// Sentinel errors for document operations.
var (
	// ErrDocumentNotFound is returned when a document cannot be found.
	ErrDocumentNotFound = errors.New("document not found")
	// ErrForbidden is returned when access to a document is denied.
	ErrForbidden = errors.New("access denied")
)

type service struct {
	repo   documentDomain.Repository
	ragSvc ragDomain.Service
	log    *logger.Logger
}

// ServiceConfig contains dependencies for creating a document service.
type ServiceConfig struct {
	Repo   documentDomain.Repository
	RAGSvc ragDomain.Service
	Log    *logger.Logger
}

// NewService creates a new document service with the given configuration.
func NewService(cfg ServiceConfig) documentDomain.Service {
	return &service{
		repo:   cfg.Repo,
		ragSvc: cfg.RAGSvc,
		log:    cfg.Log,
	}
}

func (s *service) logWarn(msg string, args ...any) {
	if s.log != nil {
		s.log.Warn(msg, args...)
	}
}

func (s *service) CreateDocument(ctx context.Context, userCtx documentDomain.UserContext, doc *documentDomain.Document) (string, error) {
	doc.UserID = userCtx.UserID

	id, err := s.repo.Create(ctx, doc)
	if err != nil {
		return "", fmt.Errorf("create document: %w", err)
	}

	if s.ragSvc != nil && doc.Content != "" {
		if err := s.ragSvc.IndexDocument(ctx, id, doc.Content); err != nil {
			s.logWarn("failed to index document", "document_id", id, "error", err)
		}
	}

	return id, nil
}

func (s *service) GetDocument(ctx context.Context, userCtx documentDomain.UserContext, id string) (*documentDomain.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	if doc == nil {
		return nil, ErrDocumentNotFound
	}

	if !userCtx.IsAdmin && doc.UserID != userCtx.UserID {
		return nil, ErrForbidden
	}

	return doc, nil
}

func (s *service) ListDocuments(ctx context.Context, userCtx documentDomain.UserContext, limit, offset int) ([]documentDomain.Document, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var docs []documentDomain.Document
	var total int64
	var err error

	if userCtx.IsAdmin {
		docs, err = s.repo.List(ctx, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("list documents: %w", err)
		}
		total, err = s.repo.Count(ctx)
	} else {
		docs, err = s.repo.ListByUser(ctx, userCtx.UserID, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("list user documents: %w", err)
		}
		total, err = s.repo.CountByUser(ctx, userCtx.UserID)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("count documents: %w", err)
	}

	return docs, total, nil
}

func (s *service) UpdateDocument(ctx context.Context, userCtx documentDomain.UserContext, doc *documentDomain.Document) error {
	existing, err := s.repo.GetByID(ctx, doc.ID)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if existing == nil {
		return ErrDocumentNotFound
	}

	if !userCtx.IsAdmin && existing.UserID != userCtx.UserID {
		return ErrForbidden
	}

	doc.UploadedAt = existing.UploadedAt
	doc.UserID = existing.UserID

	if err := s.repo.Update(ctx, doc); err != nil {
		return fmt.Errorf("update document: %w", err)
	}

	if s.ragSvc != nil && doc.Content != existing.Content {
		if err := s.ragSvc.DeleteDocumentChunks(ctx, doc.ID); err != nil {
			s.logWarn("failed to delete old chunks", "document_id", doc.ID, "error", err)
		}

		if doc.Content != "" {
			if err := s.ragSvc.IndexDocument(ctx, doc.ID, doc.Content); err != nil {
				s.logWarn("failed to index updated document", "document_id", doc.ID, "error", err)
			}
		}
	}

	return nil
}

func (s *service) DeleteDocument(ctx context.Context, userCtx documentDomain.UserContext, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}
	if existing == nil {
		return ErrDocumentNotFound
	}

	if !userCtx.IsAdmin && existing.UserID != userCtx.UserID {
		return ErrForbidden
	}

	if s.ragSvc != nil {
		if err := s.ragSvc.DeleteDocumentChunks(ctx, id); err != nil {
			s.logWarn("failed to delete chunks", "document_id", id, "error", err)
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	return nil
}
