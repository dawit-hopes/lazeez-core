package table

import (
	"context"
	"database/sql"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
	"lazeez-core/internal/files"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type TableService interface {
	Create(ctx context.Context, req TableRequest) ([]*TableDTO, error)
	RegenerateQRCode(ctx context.Context, id string, branchID string) error
	Get(ctx context.Context, id string, branchID string) (*TableDTO, error)
	Delete(ctx context.Context, id string, branchID string) error
	UnDelete(ctx context.Context, id string, branchID string) error
	List(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*TableDTO], error)
	AttachOrder(ctx context.Context, id string, orderID string, branchID string) error
	DetachOrder(ctx context.Context, id string, branchID string) error
}

type tableService struct {
	repository    TableRepository
	fileService   files.FileService
	branchService branch.BranchService
	logger        config.Logger
}

func NewTableService(repository TableRepository, fileService files.FileService, branchService branch.BranchService, logger config.Logger) TableService {
	return &tableService{repository: repository, fileService: fileService, branchService: branchService, logger: logger}
}

func (s *tableService) generateQRCode(ctx context.Context, reference string) (string, error) {
	file, err := GenerateQRCodeHeader(reference)
	if err != nil {
		s.logger.Error("Failed to generate QR code", "error", err)
		return "", err
	}
	return s.fileService.UploadFile(ctx, file)
}

func (s *tableService) validateRequest(ctx context.Context, req TableRequest) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := s.branchService.Get(ctx, req.BranchID)
		if err != nil {
			s.logger.Error("Failed to get branch", "error", err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		innG, innCtx := errgroup.WithContext(ctx)
		for _, table := range req.Tables {
			tableName := table.TableName
			innG.Go(func() error {
				if err := s.repository.CheckExists(innCtx, tableName, req.BranchID); err != nil {
					s.logger.Error("Failed to check if table exists", "error", err)
					return err
				}
				return nil
			})
		}
		return innG.Wait()
	})
	return g.Wait()
}

func (s *tableService) Create(ctx context.Context, req TableRequest) ([]*TableDTO, error) {
	if err := s.validateRequest(ctx, req); err != nil {
		s.logger.Error("Failed to validate table create request", "error", err)
		return nil, err
	}

	// Build table records with references (no QR yet)
	tables := make([]Table, 0, len(req.Tables))
	for _, t := range req.Tables {
		tbl := Table{
			TableName: common.FormatText(t.TableName),
			BranchID:  req.BranchID,
			Reference: GenerateTableRef(),
			QRVersion: 1,
			Status:    "active",
		}
		tbl.ID = uuid.New().String()
		tables = append(tables, tbl)
	}

	// Generate QR codes in parallel for speed
	g, gctx := errgroup.WithContext(ctx)
	for i := range tables {
		i := i
		g.Go(func() error {
			url, err := s.generateQRCode(gctx, tables[i].Reference)
			if err != nil {
				return err
			}
			tables[i].QRCode = url
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		s.logger.Error("Failed to generate QR codes", "error", err)
		return nil, err
	}

	out := make([]*TableDTO, 0, len(tables))
	for i := range tables {
		if err := s.repository.Create(ctx, &tables[i]); err != nil {
			s.logger.Error("Failed to create table", "table_name", tables[i].TableName, "error", err)
			return nil, err
		}
		dto := tables[i].ToDTO()
		out = append(out, &dto)
	}
	return out, nil
}

func (s *tableService) RegenerateQRCode(ctx context.Context, id string, branchID string) error {
	table, err := s.repository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get table", "error", err)
		return err
	}
	if branchID != "" && table.BranchID != branchID {
		return common.ErrTableNotFound
	}
	qrCodeURL, err := s.generateQRCode(ctx, table.Reference)
	if err != nil {
		s.logger.Error("Failed to generate QR code", "error", err)
		return err
	}
	table.QRCode = qrCodeURL
	table.QRVersion++
	return s.repository.Update(ctx, *table)
}

func (s *tableService) Get(ctx context.Context, id string, branchID string) (*TableDTO, error) {
	table, err := s.repository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get table", "error", err)
		return nil, err
	}
	if branchID != "" && table.BranchID != branchID {
		return nil, common.ErrTableNotFound
	}
	tableDTO := table.ToDTO()
	return &tableDTO, nil
}

func (s *tableService) Delete(ctx context.Context, id string, branchID string) error {
	table, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if branchID != "" && table.BranchID != branchID {
		return common.ErrTableNotFound
	}
	return s.repository.Delete(ctx, id)
}

func (s *tableService) UnDelete(ctx context.Context, id string, branchID string) error {
	table, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if branchID != "" && table.BranchID != branchID {
		return common.ErrTableNotFound
	}
	return s.repository.UnDelete(ctx, id)
}

func (s *tableService) List(ctx context.Context, filter common.Filter, branchID string) (*common.PaginatedResponse[[]*TableDTO], error) {
	result, err := s.repository.List(ctx, filter, branchID)
	if err != nil {
		s.logger.Error("Failed to list tables", "error", err)
		return nil, err
	}
	tableDTOs := make([]*TableDTO, len(result.Data))
	for i, table := range result.Data {
		tableDTO := table.ToDTO()
		tableDTOs[i] = &tableDTO
	}
	return &common.PaginatedResponse[[]*TableDTO]{
		Data: tableDTOs,
		Meta: result.Meta,
	}, nil
}

func (s *tableService) AttachOrder(ctx context.Context, id string, orderID string, branchID string) error {
	table, err := s.repository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get table", "error", err)
		return err
	}
	if branchID != "" && table.BranchID != branchID {
		return common.ErrTableNotFound
	}
	table.ActiveOrderID = sql.NullString{String: orderID, Valid: true}
	return s.repository.Update(ctx, *table)
}

func (s *tableService) DetachOrder(ctx context.Context, id string, branchID string) error {
	table, err := s.repository.Get(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get table", "error", err)
		return err
	}
	if branchID != "" && table.BranchID != branchID {
		return common.ErrTableNotFound
	}
	table.ActiveOrderID = sql.NullString{Valid: false}
	return s.repository.Update(ctx, *table)
}
