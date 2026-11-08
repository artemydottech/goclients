package service

import (
	"context"

	"github.com/artemydottech/goclients/internal/models"
)

type CompanyReportRepo interface {
	GetCompanyReport(context.Context, models.ReportQuery) (models.CompanyReport, error)
}
type CompanyReportService struct {
	repo      CompanyReportRepo
	companies CompanyLookup
}

func NewCompanyReportService(r CompanyReportRepo, c CompanyLookup) *CompanyReportService {
	return &CompanyReportService{repo: r, companies: c}
}
func (s *CompanyReportService) GetCompanyReport(ctx context.Context, q models.ReportQuery) (models.CompanyReport, error) {
	if err := q.Validate(); err != nil {
		return models.CompanyReport{}, err
	}
	if _, err := s.companies.GetCompanyById(ctx, q.CompanyID); err != nil {
		return models.CompanyReport{}, err
	}
	return s.repo.GetCompanyReport(ctx, q)
}
